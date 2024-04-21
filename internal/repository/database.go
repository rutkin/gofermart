package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	my_errors "github.com/rutkin/gofermart/internal/errors"
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

	_, err = tx.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		logger.Log.Error("Failed to create extension", zap.String("error", err.Error()))
		return nil, err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS users (userID uuid DEFAULT uuid_generate_v4(), userName VARCHAR (50) UNIQUE NOT NULL, password VARCHAR (100) NOT NULL)")
	if err != nil {
		logger.Log.Error("Failed to create users table", zap.String("error", err.Error()))
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
	var userID uuid.UUID
	err := r.db.QueryRow("INSERT INTO users (userName, password) Values ($1, $2) RETURNING userID", name, password).Scan(userID)

	if err != nil {
		logger.Log.Error("Failed to insert user", zap.String("error", err.Error()))
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = my_errors.ErrExists
		}
		return "", err
	}

	return userID.String(), nil
}

func (r *Database) GetUserID(name string, password string) (string, error) {
	var userID uuid.UUID
	err := r.db.QueryRow("SELECT userID FROM users WHERE userName=$1 AND password=$2", name, password).Scan(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", my_errors.ErrNotFound
		}
		return "", err
	}
	return userID.String(), nil
}
