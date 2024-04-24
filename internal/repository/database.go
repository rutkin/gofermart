package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	myerrors "github.com/rutkin/gofermart/internal/errors"
	"github.com/rutkin/gofermart/internal/logger"
	"go.uber.org/zap"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDatabase(databaseURI string) (*Database, error) {
	db, err := sql.Open("pgx", databaseURI)
	if err != nil {
		logger.Log.Error("failed to open db", zap.String("error", err.Error()))
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create transaction", zap.String("error", err.Error()))
		return nil, err
	}

	defer tx.Rollback()

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS users (userID VARCHAR(50), userName VARCHAR (50) UNIQUE NOT NULL, password VARCHAR (100) NOT NULL)")
	if err != nil {
		logger.Log.Error("Failed to create users table", zap.String("error", err.Error()))
		return nil, err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS orders (userID VARCHAR(50), number VARCHAR (50) UNIQUE NOT NULL, status VARCHAR (50), accrual INTEGER, date DATE)")
	if err != nil {
		logger.Log.Error("Failed to create orders table", zap.String("error", err.Error()))
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Error("Failed to prepare db", zap.String("error", err.Error()))
		return nil, err
	}

	return &Database{db}, nil
}

type Database struct {
	db *sql.DB
}

func (r *Database) CreateUser(name string, password string) (string, error) {
	userID := uuid.New().String()
	_, err := r.db.Exec("INSERT INTO users (userID, userName, password) Values ($1, $2, $3)", userID, name, password)

	if err != nil {
		logger.Log.Error("Failed to insert user", zap.String("error", err.Error()))
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = myerrors.ErrExists
		}
		return "", err
	}

	return userID, nil
}

func (r *Database) GetUserID(name string, password string) (string, error) {
	var userID string
	err := r.db.QueryRow("SELECT userID FROM users WHERE userName=$1 AND password=$2", name, password).Scan(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", myerrors.ErrNotFound
		}
		return "", err
	}
	return userID, nil
}

func (r *Database) CreateOrder(userID string, number string) error {
	var currentUserID string
	tx, err := r.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create transaction", zap.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT userID FROM orders WHERE number=$1;", number)
	if err != nil {
		logger.Log.Error("failed to select user from order", zap.String("error", err.Error()))
		return err
	}
	rows.Next()
	err = rows.Scan(currentUserID)
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if currentUserID != "" {
		if currentUserID != userID {
			return myerrors.ErrConflict
		}
		return myerrors.ErrExists
	}
	_, err = tx.Exec("INSERT INTO orders (userID, number, status, accrual, date) Values ($1, $2, NEW, 0, current_timestamp)", userID, number)
	if err != nil {
		logger.Log.Error("failed to insert order", zap.String("error", err.Error()))
		return err
	}
	tx.Commit()
	return nil
}
