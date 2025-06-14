// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: transaction.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createTransaction = `-- name: CreateTransaction :exec
INSERT INTO transaction (type, category_id, amount, date, description) VALUES ($1, $2, $3, $4, $5)
`

type CreateTransactionParams struct {
	Type        TransactionType `json:"type"`
	CategoryID  uuid.UUID       `json:"category_id"`
	Amount      int32           `json:"amount"`
	Date        time.Time       `json:"date"`
	Description string          `json:"description"`
}

func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) error {
	_, err := q.db.ExecContext(ctx, createTransaction,
		arg.Type,
		arg.CategoryID,
		arg.Amount,
		arg.Date,
		arg.Description,
	)
	return err
}

const deleteTransaction = `-- name: DeleteTransaction :exec
DELETE FROM transaction WHERE id = $1
`

func (q *Queries) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteTransaction, id)
	return err
}

const getTotalExpenses = `-- name: GetTotalExpenses :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total_expenses FROM transaction WHERE type = 'expense'
`

func (q *Queries) GetTotalExpenses(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, getTotalExpenses)
	var total_expenses int64
	err := row.Scan(&total_expenses)
	return total_expenses, err
}

const getTotalIncome = `-- name: GetTotalIncome :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total_income FROM transaction WHERE type = 'income'
`

func (q *Queries) GetTotalIncome(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, getTotalIncome)
	var total_income int64
	err := row.Scan(&total_income)
	return total_income, err
}

const getTransaction = `-- name: GetTransaction :one
SELECT transaction.id, transaction.type, transaction.category_id, transaction.amount, transaction.date, transaction.description, transaction.created_at, category.name as category_name 
FROM transaction JOIN category ON transaction.category_id = category.id WHERE transaction.id = $1
`

type GetTransactionRow struct {
	ID           uuid.UUID       `json:"id"`
	Type         TransactionType `json:"type"`
	CategoryID   uuid.UUID       `json:"category_id"`
	Amount       int32           `json:"amount"`
	Date         time.Time       `json:"date"`
	Description  string          `json:"description"`
	CreatedAt    time.Time       `json:"created_at"`
	CategoryName string          `json:"category_name"`
}

func (q *Queries) GetTransaction(ctx context.Context, id uuid.UUID) (GetTransactionRow, error) {
	row := q.db.QueryRowContext(ctx, getTransaction, id)
	var i GetTransactionRow
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.CategoryID,
		&i.Amount,
		&i.Date,
		&i.Description,
		&i.CreatedAt,
		&i.CategoryName,
	)
	return i, err
}

const getTransactions = `-- name: GetTransactions :many
SELECT transaction.id, transaction.type, transaction.category_id, transaction.amount, transaction.date, transaction.description, transaction.created_at, category.name as category_name 
FROM transaction JOIN category ON transaction.category_id = category.id
`

type GetTransactionsRow struct {
	ID           uuid.UUID       `json:"id"`
	Type         TransactionType `json:"type"`
	CategoryID   uuid.UUID       `json:"category_id"`
	Amount       int32           `json:"amount"`
	Date         time.Time       `json:"date"`
	Description  string          `json:"description"`
	CreatedAt    time.Time       `json:"created_at"`
	CategoryName string          `json:"category_name"`
}

func (q *Queries) GetTransactions(ctx context.Context) ([]GetTransactionsRow, error) {
	rows, err := q.db.QueryContext(ctx, getTransactions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetTransactionsRow{}
	for rows.Next() {
		var i GetTransactionsRow
		if err := rows.Scan(
			&i.ID,
			&i.Type,
			&i.CategoryID,
			&i.Amount,
			&i.Date,
			&i.Description,
			&i.CreatedAt,
			&i.CategoryName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTransactionsByType = `-- name: GetTransactionsByType :many
WITH filtered_transactions AS (
    SELECT transaction.id, transaction.type, transaction.category_id, transaction.amount, transaction.date, transaction.description, transaction.created_at, category.name as category_name 
    FROM transaction JOIN category ON transaction.category_id = category.id
    WHERE transaction.type = $1
    AND ($2::text = '' OR category.name = $2)
)
SELECT 
    t.id, t.type, t.category_id, t.amount, t.date, t.description, t.created_at, t.category_name,
    (SELECT COALESCE(SUM(amount), 0)::bigint FROM filtered_transactions) AS total_sum
FROM filtered_transactions t
ORDER BY t.created_at DESC
`

type GetTransactionsByTypeParams struct {
	Type         TransactionType `json:"type"`
	CategoryName string          `json:"category_name"`
}

type GetTransactionsByTypeRow struct {
	ID           uuid.UUID       `json:"id"`
	Type         TransactionType `json:"type"`
	CategoryID   uuid.UUID       `json:"category_id"`
	Amount       int32           `json:"amount"`
	Date         time.Time       `json:"date"`
	Description  string          `json:"description"`
	CreatedAt    time.Time       `json:"created_at"`
	CategoryName string          `json:"category_name"`
	TotalSum     int64           `json:"total_sum"`
}

func (q *Queries) GetTransactionsByType(ctx context.Context, arg GetTransactionsByTypeParams) ([]GetTransactionsByTypeRow, error) {
	rows, err := q.db.QueryContext(ctx, getTransactionsByType, arg.Type, arg.CategoryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetTransactionsByTypeRow{}
	for rows.Next() {
		var i GetTransactionsByTypeRow
		if err := rows.Scan(
			&i.ID,
			&i.Type,
			&i.CategoryID,
			&i.Amount,
			&i.Date,
			&i.Description,
			&i.CreatedAt,
			&i.CategoryName,
			&i.TotalSum,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTransaction = `-- name: UpdateTransaction :exec
UPDATE transaction SET type = $2, category_id = $3, amount = $4, date = $5, description = $6 WHERE id = $1
`

type UpdateTransactionParams struct {
	ID          uuid.UUID       `json:"id"`
	Type        TransactionType `json:"type"`
	CategoryID  uuid.UUID       `json:"category_id"`
	Amount      int32           `json:"amount"`
	Date        time.Time       `json:"date"`
	Description string          `json:"description"`
}

func (q *Queries) UpdateTransaction(ctx context.Context, arg UpdateTransactionParams) error {
	_, err := q.db.ExecContext(ctx, updateTransaction,
		arg.ID,
		arg.Type,
		arg.CategoryID,
		arg.Amount,
		arg.Date,
		arg.Description,
	)
	return err
}
