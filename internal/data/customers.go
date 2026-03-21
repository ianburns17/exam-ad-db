package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Customer struct {
	ID                  int64      `json:"id" db:"customer_id"`
	FirstName           string     `json:"first_name" db:"first_name"`
	LastName            string     `json:"last_name" db:"last_name"`
	Email               string     `json:"email" db:"email"`
	Phone               string     `json:"phone" db:"phone"`
	DriverLicenseNumber string     `json:"driver_license_number" db:"driver_license_number"`
	LicenseExpiry       *time.Time `json:"license_expiry" db:"license_expiry"`
	IsActive            *bool      `json:"is_active" db:"is_active"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}

type CustomerModel struct {
	DB *sql.DB
}

func (m *Customer) Validate() error {
	return nil
}

func (mod *CustomerModel) Insert(m *Customer) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO customers (first_name, last_name, email, phone, driver_license_number, license_expiry, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING customer_id, created_at`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.FirstName, m.LastName, m.Email, m.Phone, m.DriverLicenseNumber, m.LicenseExpiry, m.IsActive,
	).Scan(&m.ID, &m.CreatedAt)
}

func (mod *CustomerModel) Get(id int64) (*Customer, error) {
	var m Customer

	query := `SELECT customer_id, first_name, last_name, email, phone, driver_license_number, license_expiry, is_active, created_at FROM customers WHERE customer_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.FirstName, &m.LastName, &m.Email, &m.Phone, &m.DriverLicenseNumber, &m.LicenseExpiry, &m.IsActive, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *CustomerModel) Update(id int64, m *Customer) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE customers SET first_name=$1, last_name=$2, email=$3, phone=$4, driver_license_number=$5, license_expiry=$6, is_active=$7 WHERE customer_id=$8`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.FirstName, m.LastName, m.Email, m.Phone, m.DriverLicenseNumber, m.LicenseExpiry, m.IsActive, id,
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

func (mod *CustomerModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE customers SET %s WHERE customer_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedCustomerSorts = map[string]string{

	"customer_id": "customer_id",

	"first_name": "first_name",

	"last_name": "last_name",

	"email": "email",

	"phone": "phone",

	"driver_license_number": "driver_license_number",

	"license_expiry": "license_expiry",

	"is_active": "is_active",

	"created_at": "created_at",
}

func (mod *CustomerModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*Customer, int, error) {

	sortCol, ok := allowedCustomerSorts[sort]
	if !ok {
		sortCol = "customer_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM customers").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT customer_id, first_name, last_name, email, phone, driver_license_number, license_expiry, is_active, created_at FROM customers ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Customer
	for rows.Next() {
		var m Customer

		if err := rows.Scan(&m.ID, &m.FirstName, &m.LastName, &m.Email, &m.Phone, &m.DriverLicenseNumber, &m.LicenseExpiry, &m.IsActive, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *CustomerModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM customers WHERE customer_id = $1`, id)
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
