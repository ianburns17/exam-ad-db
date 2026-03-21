package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Inspection struct {
	ID             int64      `json:"id" db:"inspection_id"`
	VehicleID      *int64     `json:"vehicle_id" db:"vehicle_id"`
	RentalID       int64      `json:"rental_id" db:"rental_id"`
	InspectionDate *time.Time `json:"inspection_date" db:"inspection_date"`
	InspectorName  string     `json:"inspector_name" db:"inspector_name"`
	DamageFound    *bool      `json:"damage_found" db:"damage_found"`
	Notes          string     `json:"notes" db:"notes"`
}

type InspectionModel struct {
	DB *sql.DB
}

func (m *Inspection) Validate() error {
	return nil
}

func (mod *InspectionModel) Insert(m *Inspection) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO inspections (vehicle_id, rental_id, inspection_date, inspector_name, damage_found, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING inspection_id`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.VehicleID, m.RentalID, m.InspectionDate, m.InspectorName, m.DamageFound, m.Notes,
	).Scan(&m.ID)
}

func (mod *InspectionModel) Get(id int64) (*Inspection, error) {
	var m Inspection

	query := `SELECT inspection_id, vehicle_id, rental_id, inspection_date, inspector_name, damage_found, notes FROM inspections WHERE inspection_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.VehicleID, &m.RentalID, &m.InspectionDate, &m.InspectorName, &m.DamageFound, &m.Notes,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *InspectionModel) Update(id int64, m *Inspection) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE inspections SET vehicle_id=$1, rental_id=$2, inspection_date=$3, inspector_name=$4, damage_found=$5, notes=$6 WHERE inspection_id=$7`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.VehicleID, m.RentalID, m.InspectionDate, m.InspectorName, m.DamageFound, m.Notes, id,
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

func (mod *InspectionModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE inspections SET %s WHERE inspection_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedInspectionSorts = map[string]string{

	"inspection_id": "inspection_id",

	"vehicle_id": "vehicle_id",

	"rental_id": "rental_id",

	"inspection_date": "inspection_date",

	"inspector_name": "inspector_name",

	"damage_found": "damage_found",

	"notes": "notes",
}

func (mod *InspectionModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*Inspection, int, error) {

	sortCol, ok := allowedInspectionSorts[sort]
	if !ok {
		sortCol = "inspection_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM inspections").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT inspection_id, vehicle_id, rental_id, inspection_date, inspector_name, damage_found, notes FROM inspections ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Inspection
	for rows.Next() {
		var m Inspection

		if err := rows.Scan(&m.ID, &m.VehicleID, &m.RentalID, &m.InspectionDate, &m.InspectorName, &m.DamageFound, &m.Notes); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *InspectionModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM inspections WHERE inspection_id = $1`, id)
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
