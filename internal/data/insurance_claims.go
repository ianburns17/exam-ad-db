package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type InsuranceClaim struct {
	ID          int64      `json:"id" db:"claim_id"`
	VehicleID   *int64     `json:"vehicle_id" db:"vehicle_id"`
	RentalID    int64      `json:"rental_id" db:"rental_id"`
	ClaimDate   *time.Time `json:"claim_date" db:"claim_date"`
	Description string     `json:"description" db:"description"`
	ClaimAmount *float64   `json:"claim_amount" db:"claim_amount"`
	Status      string     `json:"status" db:"status"`
}

type InsuranceClaimModel struct {
	DB *sql.DB
}

func (m *InsuranceClaim) Validate() error {
	return nil
}

func (mod *InsuranceClaimModel) Insert(m *InsuranceClaim) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO insuranceclaims (vehicle_id, rental_id, claim_date, description, claim_amount, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING claim_id`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.VehicleID, m.RentalID, m.ClaimDate, m.Description, m.ClaimAmount, m.Status,
	).Scan(&m.ID)
}

func (mod *InsuranceClaimModel) Get(id int64) (*InsuranceClaim, error) {
	var m InsuranceClaim

	query := `SELECT claim_id, vehicle_id, rental_id, claim_date, description, claim_amount, status FROM insuranceclaims WHERE claim_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.VehicleID, &m.RentalID, &m.ClaimDate, &m.Description, &m.ClaimAmount, &m.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *InsuranceClaimModel) Update(id int64, m *InsuranceClaim) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE insuranceclaims SET vehicle_id=$1, rental_id=$2, claim_date=$3, description=$4, claim_amount=$5, status=$6 WHERE claim_id=$7`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.VehicleID, m.RentalID, m.ClaimDate, m.Description, m.ClaimAmount, m.Status, id,
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

func (mod *InsuranceClaimModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE insuranceclaims SET %s WHERE claim_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedInsuranceClaimSorts = map[string]string{

	"claim_id": "claim_id",

	"vehicle_id": "vehicle_id",

	"rental_id": "rental_id",

	"claim_date": "claim_date",

	"description": "description",

	"claim_amount": "claim_amount",

	"status": "status",
}

func (mod *InsuranceClaimModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*InsuranceClaim, int, error) {

	sortCol, ok := allowedInsuranceClaimSorts[sort]
	if !ok {
		sortCol = "claim_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM insuranceclaims").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT claim_id, vehicle_id, rental_id, claim_date, description, claim_amount, status FROM insuranceclaims ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*InsuranceClaim
	for rows.Next() {
		var m InsuranceClaim

		if err := rows.Scan(&m.ID, &m.VehicleID, &m.RentalID, &m.ClaimDate, &m.Description, &m.ClaimAmount, &m.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *InsuranceClaimModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM insuranceclaims WHERE claim_id = $1`, id)
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
