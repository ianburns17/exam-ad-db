package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Fee struct {
	ID          int64     `json:"id" db:"fee_id"`
	RentalID    int64     `json:"rental_id" db:"rental_id"`
	CustomerID  int64     `json:"customer_id" db:"customer_id"`
	FeeType     string    `json:"fee_type" db:"fee_type"`
	Amount      *float64  `json:"amount" db:"amount"`
	Description string    `json:"description" db:"description"`
	IsPaid      *bool     `json:"is_paid" db:"is_paid"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type FeeModel struct {
	DB *sql.DB
}

func (m *Fee) Validate() error {
	return nil
}

func (mod *FeeModel) Insert(m *Fee) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO fees (rental_id, customer_id, fee_type, amount, description, is_paid)
		VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING fee_id, created_at`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.RentalID, m.CustomerID, m.FeeType, m.Amount, m.Description, m.IsPaid,
	).Scan(&m.ID, &m.CreatedAt)
}

func (mod *FeeModel) Get(id int64) (*Fee, error) {
	var m Fee

	query := `SELECT fee_id, rental_id, customer_id, fee_type, amount, description, is_paid, created_at FROM fees WHERE fee_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.RentalID, &m.CustomerID, &m.FeeType, &m.Amount, &m.Description, &m.IsPaid, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *FeeModel) Update(id int64, m *Fee) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE fees SET rental_id=$1, customer_id=$2, fee_type=$3, amount=$4, description=$5, is_paid=$6 WHERE fee_id=$7`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.RentalID, m.CustomerID, m.FeeType, m.Amount, m.Description, m.IsPaid, id,
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

func (mod *FeeModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE fees SET %s WHERE fee_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedFeeSorts = map[string]string{

	"fee_id": "fee_id",

	"rental_id": "rental_id",

	"customer_id": "customer_id",

	"fee_type": "fee_type",

	"amount": "amount",

	"description": "description",

	"is_paid": "is_paid",

	"created_at": "created_at",
}

func (mod *FeeModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*Fee, int, error) {

	sortCol, ok := allowedFeeSorts[sort]
	if !ok {
		sortCol = "fee_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM fees").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT fee_id, rental_id, customer_id, fee_type, amount, description, is_paid, created_at FROM fees ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Fee
	for rows.Next() {
		var m Fee

		if err := rows.Scan(&m.ID, &m.RentalID, &m.CustomerID, &m.FeeType, &m.Amount, &m.Description, &m.IsPaid, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *FeeModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM fees WHERE fee_id = $1`, id)
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
