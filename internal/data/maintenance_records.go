package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type MaintenanceRecord struct {
	ID                 int64      `json:"id" db:"maintenance_id"`
	VehicleID          *int64     `json:"vehicle_id" db:"vehicle_id"`
	ServiceType        string     `json:"service_type" db:"service_type"`
	ServiceDate        *time.Time `json:"service_date" db:"service_date"`
	MileageAtService   *int       `json:"mileage_at_service" db:"mileage_at_service"`
	NextServiceMileage *int       `json:"next_service_mileage" db:"next_service_mileage"`
	NextServiceDate    *time.Time `json:"next_service_date" db:"next_service_date"`
	Cost               *float64   `json:"cost" db:"cost"`
	Notes              string     `json:"notes" db:"notes"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type MaintenanceRecordModel struct {
	DB *sql.DB
}

func (m *MaintenanceRecord) Validate() error {
	return nil
}

func (mod *MaintenanceRecordModel) Insert(m *MaintenanceRecord) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO maintenancerecords (vehicle_id, service_type, service_date, mileage_at_service, next_service_mileage, next_service_date, cost, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING maintenance_id, created_at`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.VehicleID, m.ServiceType, m.ServiceDate, m.MileageAtService, m.NextServiceMileage, m.NextServiceDate, m.Cost, m.Notes,
	).Scan(&m.ID, &m.CreatedAt)
}

func (mod *MaintenanceRecordModel) Get(id int64) (*MaintenanceRecord, error) {
	var m MaintenanceRecord

	query := `SELECT maintenance_id, vehicle_id, service_type, service_date, mileage_at_service, next_service_mileage, next_service_date, cost, notes, created_at FROM maintenancerecords WHERE maintenance_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.VehicleID, &m.ServiceType, &m.ServiceDate, &m.MileageAtService, &m.NextServiceMileage, &m.NextServiceDate, &m.Cost, &m.Notes, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *MaintenanceRecordModel) Update(id int64, m *MaintenanceRecord) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE maintenancerecords SET vehicle_id=$1, service_type=$2, service_date=$3, mileage_at_service=$4, next_service_mileage=$5, next_service_date=$6, cost=$7, notes=$8 WHERE maintenance_id=$9`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.VehicleID, m.ServiceType, m.ServiceDate, m.MileageAtService, m.NextServiceMileage, m.NextServiceDate, m.Cost, m.Notes, id,
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

func (mod *MaintenanceRecordModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE maintenancerecords SET %s WHERE maintenance_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedMaintenanceRecordSorts = map[string]string{

	"maintenance_id": "maintenance_id",

	"vehicle_id": "vehicle_id",

	"service_type": "service_type",

	"service_date": "service_date",

	"mileage_at_service": "mileage_at_service",

	"next_service_mileage": "next_service_mileage",

	"next_service_date": "next_service_date",

	"cost": "cost",

	"notes": "notes",

	"created_at": "created_at",
}

func (mod *MaintenanceRecordModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*MaintenanceRecord, int, error) {

	sortCol, ok := allowedMaintenanceRecordSorts[sort]
	if !ok {
		sortCol = "maintenance_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM maintenancerecords").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT maintenance_id, vehicle_id, service_type, service_date, mileage_at_service, next_service_mileage, next_service_date, cost, notes, created_at FROM maintenancerecords ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*MaintenanceRecord
	for rows.Next() {
		var m MaintenanceRecord

		if err := rows.Scan(&m.ID, &m.VehicleID, &m.ServiceType, &m.ServiceDate, &m.MileageAtService, &m.NextServiceMileage, &m.NextServiceDate, &m.Cost, &m.Notes, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *MaintenanceRecordModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM maintenancerecords WHERE maintenance_id = $1`, id)
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
