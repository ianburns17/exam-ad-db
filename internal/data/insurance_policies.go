package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type InsurancePolicy struct {
	ID              int64      `json:"id" db:"policy_id"`
	VehicleID       *int64     `json:"vehicle_id" db:"vehicle_id"`
	Provider        string     `json:"provider" db:"provider"`
	PolicyNumber    string     `json:"policy_number" db:"policy_number"`
	CoverageDetails string     `json:"coverage_details" db:"coverage_details"`
	StartDate       *time.Time `json:"start_date" db:"start_date"`
	EndDate         *time.Time `json:"end_date" db:"end_date"`
	Premium         *float64   `json:"premium" db:"premium"`
}

type InsurancePolicyModel struct {
	DB *sql.DB
}

func (m *InsurancePolicy) Validate() error {
	return nil
}

func (mod *InsurancePolicyModel) Insert(m *InsurancePolicy) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO insurancepolicies (vehicle_id, provider, policy_number, coverage_details, start_date, end_date, premium)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING policy_id`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.VehicleID, m.Provider, m.PolicyNumber, m.CoverageDetails, m.StartDate, m.EndDate, m.Premium,
	).Scan(&m.ID)
}

func (mod *InsurancePolicyModel) Get(id int64) (*InsurancePolicy, error) {
	var m InsurancePolicy

	query := `SELECT policy_id, vehicle_id, provider, policy_number, coverage_details, start_date, end_date, premium FROM insurancepolicies WHERE policy_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.VehicleID, &m.Provider, &m.PolicyNumber, &m.CoverageDetails, &m.StartDate, &m.EndDate, &m.Premium,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *InsurancePolicyModel) Update(id int64, m *InsurancePolicy) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE insurancepolicies SET vehicle_id=$1, provider=$2, policy_number=$3, coverage_details=$4, start_date=$5, end_date=$6, premium=$7 WHERE policy_id=$8`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.VehicleID, m.Provider, m.PolicyNumber, m.CoverageDetails, m.StartDate, m.EndDate, m.Premium, id,
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

func (mod *InsurancePolicyModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE insurancepolicies SET %s WHERE policy_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedInsurancePolicySorts = map[string]string{

	"policy_id": "policy_id",

	"vehicle_id": "vehicle_id",

	"provider": "provider",

	"policy_number": "policy_number",

	"coverage_details": "coverage_details",

	"start_date": "start_date",

	"end_date": "end_date",

	"premium": "premium",
}

func (mod *InsurancePolicyModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*InsurancePolicy, int, error) {

	sortCol, ok := allowedInsurancePolicySorts[sort]
	if !ok {
		sortCol = "policy_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM insurancepolicies").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT policy_id, vehicle_id, provider, policy_number, coverage_details, start_date, end_date, premium FROM insurancepolicies ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*InsurancePolicy
	for rows.Next() {
		var m InsurancePolicy

		if err := rows.Scan(&m.ID, &m.VehicleID, &m.Provider, &m.PolicyNumber, &m.CoverageDetails, &m.StartDate, &m.EndDate, &m.Premium); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *InsurancePolicyModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM insurancepolicies WHERE policy_id = $1`, id)
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
