package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
)

type PaymentRepository interface {
	ListPayments(status, sort, id string) ([]*entity.Payment, error)
}

type Payment struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *Payment {
	return &Payment{db: db}
}

func (r *Payment) ListPayments(status, sort, id string) ([]*entity.Payment, error) {
	query := `SELECT id, merchant, amount, status, created_at FROM payments WHERE 1=1`
	args := []any{}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	if id != "" {
		query += ` AND id = ?`
		args = append(args, id)
	}

	orderCol := "created_at"
	orderDir := "DESC"
	switch sort {
	case "created_at":
		orderDir = "ASC"
	case "-amount":
		orderCol = "amount"
		orderDir = "DESC"
	case "amount":
		orderCol = "amount"
		orderDir = "ASC"
	}
	query += fmt.Sprintf(` ORDER BY %s %s`, orderCol, orderDir)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, entity.WrapError(err, entity.ErrorCodeInternal, "db error")
	}
	defer rows.Close()

	var payments []*entity.Payment
	for rows.Next() {
		var p entity.Payment
		var createdAt string
		if err := rows.Scan(&p.ID, &p.Merchant, &p.Amount, &p.Status, &createdAt); err != nil {
			return nil, entity.WrapError(err, entity.ErrorCodeInternal, "scan error")
		}
		parsed, err := time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			return nil, entity.WrapError(err, entity.ErrorCodeInternal, "time parse error")
		}
		p.CreatedAt = parsed
		payments = append(payments, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, entity.WrapError(err, entity.ErrorCodeInternal, "row iteration error")
	}
	return payments, nil
}
