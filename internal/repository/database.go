package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	myerrors "github.com/rutkin/gofermart/internal/errors"
	"github.com/rutkin/gofermart/internal/logger"
	"github.com/rutkin/gofermart/internal/models"
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

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS orders (userID VARCHAR(50), number VARCHAR (50) UNIQUE NOT NULL, status VARCHAR (50), accrual REAL, date DATE)")
	if err != nil {
		logger.Log.Error("Failed to create orders table", zap.String("error", err.Error()))
		return nil, err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS balance (userID VARCHAR(50) UNIQUE NOT NULL, sum REAL, withDrawn REAL)")
	if err != nil {
		logger.Log.Error("Failed to create balance table", zap.String("error", err.Error()))
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
	err := r.db.QueryRow("SELECT userID FROM users WHERE userName=$1 AND password=$2", name, password).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", myerrors.ErrNotFound
		}
		return "", err
	}
	return userID, nil
}

func (r *Database) CreateOrder(userID string, number string) error {
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
	if rows.Next() {
		err = rows.Err()
		if err != nil {
			logger.Log.Error("Failed to iterate db", zap.String("error", err.Error()))
			return err
		}
		var currentUserID string
		err = rows.Scan(&currentUserID)
		if err != nil {
			logger.Log.Error("Failed to scan value", zap.String("error", err.Error()))
			return err
		}

		if currentUserID != userID {
			return myerrors.ErrConflict
		}
		return myerrors.ErrExists
	}
	_, err = tx.Exec("INSERT INTO orders (userID, number, status, accrual, date) Values ($1, $2, 'NEW', 0, current_timestamp)", userID, number)
	if err != nil {
		logger.Log.Error("failed to insert order", zap.String("error", err.Error()))
		return err
	}
	tx.Commit()
	return nil
}

func (r *Database) GetOrders(userID string) (models.OrdersResponse, error) {
	rows, err := r.db.Query("SELECT number, status, accrual, date FROM orders WHERE userID=$1;", userID)
	if err != nil {
		logger.Log.Error("Failed to get orders from db", zap.String("error", err.Error()))
		return nil, err
	}

	var result []models.OrderRecord
	for rows.Next() {
		err := rows.Err()
		if err != nil {
			logger.Log.Error("Failed to iterate db", zap.String("error", err.Error()))
			return nil, err
		}
		var record models.OrderRecord
		if err := rows.Scan(&record.Number, &record.Status, &record.Accrual, &record.UploadetAt); err != nil {
			logger.Log.Error("Failed to scan get urls result", zap.String("error", err.Error()))
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

func (r *Database) UpdateOrder(userID string, number string, status string, accrual float32) error {
	tx, err := r.db.Begin()
	if err != nil {
		logger.Log.Error("Failed to create transaction", zap.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE orders SET status=$1, accrual=$2 WHERE number=$3", status, accrual, number)
	if err != nil {
		logger.Log.Error("Failed to delete urls from db", zap.String("error", err.Error()))
		return err
	}

	_, err = tx.Exec("INSERT INTO balance VALUES (userID, sum, withDrawn) Values ($1, $2, 0) ON CONFLICT (userID) DO UPDATE SET sum = sum+$2", userID, accrual)
	if err != nil {
		logger.Log.Error("Failed to update balance", zap.String("error", err.Error()))
		return err
	}

	return tx.Commit()
}

func (r *Database) GetBalance(userID string) (models.BalanceRecord, error) {
	var result models.BalanceRecord
	err := r.db.QueryRow("SELECT sum, withDrawn FROM balance WHERE userID=$1", userID).Scan(&result.Current, &result.Withdrawn)
	if err != nil {
		logger.Log.Error("Failed to get balance", zap.String("error", err.Error()))
		return models.BalanceRecord{}, err
	}
	return result, nil
}
