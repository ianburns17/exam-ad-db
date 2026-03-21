package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type CustomerViolation struct {
	ID          int64     `json:"id" db:"violation_id"`
	CustomerID  int64     `json:"customer_id" db:"customer_id"`
	Description string    `json:"description" db:"description"`
	FeeAmount   *float64  `json:"fee_amount" db:"fee_amount"`
	IsPaid      *bool     `json:"is_paid" db:"is_paid"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CustomerViolationModel struct {
	DB *sql.DB
}

func (m *CustomerViolation) Validate() error {
	return nil
}

func (mod *CustomerViolationModel) Insert(m *CustomerViolation) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO customerviolations (customer_id, description, fee_amount, is_paid)
		VALUES ($1, $2, $3, $4)
		 RETURNING violation_id, created_at`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.CustomerID, m.Description, m.FeeAmount, m.IsPaid,
	).Scan(&m.ID, &m.CreatedAt)
}

func (mod *CustomerViolationModel) Get(id int64) (*CustomerViolation, error) {
	var m CustomerViolation

	query := `SELECT violation_id, customer_id, description, fee_amount, is_paid, created_at FROM customerviolations WHERE violation_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.CustomerID, &m.Description, &m.FeeAmount, &m.IsPaid, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *CustomerViolationModel) Update(id int64, m *CustomerViolation) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE customerviolations SET customer_id=$1, description=$2, fee_amount=$3, is_paid=$4 WHERE violation_id=$5`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.CustomerID, m.Description, m.FeeAmount, m.IsPaid, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (mod *CustomerViolationModel) PartialUpdate(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}
	setParts := []string{}
	args := []interface{}{}
	i := 1
	for k, v := range updates {
		setParts = append(setParts, fmt.Sprintf("%s=$%d", k, i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE customerviolations SET %s WHERE violation_id=$%d", strings.Join(setParts, ", "), i)
	res, err := mod.DB.ExecContext(context.Background(), query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

var allowedCustomerViolationSorts = map[string]string{

	"violation_id": "violation_id",

	"customer_id": "customer_id",

	"description": "description",

	"fee_amount": "fee_amount",

	"is_paid": "is_paid",

	"created_at": "created_at",
}

func (mod *CustomerViolationModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*CustomerViolation, int, error) {

	sortCol, ok := allowedCustomerViolationSorts[sort]
	if !ok {
		sortCol = "violation_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM customerviolations").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT violation_id, customer_id, description, fee_amount, is_paid, created_at FROM customerviolations ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*CustomerViolation
	for rows.Next() {
		var m CustomerViolation

		if err := rows.Scan(&m.ID, &m.CustomerID, &m.Description, &m.FeeAmount, &m.IsPaid, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *CustomerViolationModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM customerviolations WHERE violation_id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
