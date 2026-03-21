package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Rental struct {
	ID                 int64      `json:"id" db:"rental_id"`
	VehicleID          *int64     `json:"vehicle_id" db:"vehicle_id"`
	CustomerID         int64      `json:"customer_id" db:"customer_id"`
	PickupLocationID   *int64     `json:"pickup_location_id" db:"pickup_location_id"`
	DropoffLocationID  *int64     `json:"dropoff_location_id" db:"dropoff_location_id"`
	PickupDatetime     *time.Time `json:"pickup_datetime" db:"pickup_datetime"`
	DropoffDatetime    *time.Time `json:"dropoff_datetime" db:"dropoff_datetime"`
	StartMileage       *int       `json:"start_mileage" db:"start_mileage"`
	EndMileage         *int       `json:"end_mileage" db:"end_mileage"`
	StartFuelLevel     *float64   `json:"start_fuel_level" db:"start_fuel_level"`
	EndFuelLevel       *float64   `json:"end_fuel_level" db:"end_fuel_level"`
	DailyRate          *float64   `json:"daily_rate" db:"daily_rate"`
	MileageLimitPerDay *int       `json:"mileage_limit_per_day" db:"mileage_limit_per_day"`
	ExtraMileageFee    *float64   `json:"extra_mileage_fee" db:"extra_mileage_fee"`
	TotalCost          *float64   `json:"total_cost" db:"total_cost"`
	Status             string     `json:"status" db:"status"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type RentalModel struct {
	DB *sql.DB
}

func (m *Rental) Validate() error {
	return nil
}

func (mod *RentalModel) Insert(m *Rental) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO rentals (vehicle_id, customer_id, pickup_location_id, dropoff_location_id, pickup_datetime, dropoff_datetime, start_mileage, end_mileage, start_fuel_level, end_fuel_level, daily_rate, mileage_limit_per_day, extra_mileage_fee, total_cost, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 RETURNING rental_id, created_at`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.VehicleID, m.CustomerID, m.PickupLocationID, m.DropoffLocationID, m.PickupDatetime, m.DropoffDatetime, m.StartMileage, m.EndMileage, m.StartFuelLevel, m.EndFuelLevel, m.DailyRate, m.MileageLimitPerDay, m.ExtraMileageFee, m.TotalCost, m.Status,
	).Scan(&m.ID, &m.CreatedAt)
}

func (mod *RentalModel) Get(id int64) (*Rental, error) {
	var m Rental

	query := `SELECT rental_id, vehicle_id, customer_id, pickup_location_id, dropoff_location_id, pickup_datetime, dropoff_datetime, start_mileage, end_mileage, start_fuel_level, end_fuel_level, daily_rate, mileage_limit_per_day, extra_mileage_fee, total_cost, status, created_at FROM rentals WHERE rental_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.VehicleID, &m.CustomerID, &m.PickupLocationID, &m.DropoffLocationID, &m.PickupDatetime, &m.DropoffDatetime, &m.StartMileage, &m.EndMileage, &m.StartFuelLevel, &m.EndFuelLevel, &m.DailyRate, &m.MileageLimitPerDay, &m.ExtraMileageFee, &m.TotalCost, &m.Status, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *RentalModel) Update(id int64, m *Rental) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE rentals SET vehicle_id=$1, customer_id=$2, pickup_location_id=$3, dropoff_location_id=$4, pickup_datetime=$5, dropoff_datetime=$6, start_mileage=$7, end_mileage=$8, start_fuel_level=$9, end_fuel_level=$10, daily_rate=$11, mileage_limit_per_day=$12, extra_mileage_fee=$13, total_cost=$14, status=$15 WHERE rental_id=$16`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.VehicleID, m.CustomerID, m.PickupLocationID, m.DropoffLocationID, m.PickupDatetime, m.DropoffDatetime, m.StartMileage, m.EndMileage, m.StartFuelLevel, m.EndFuelLevel, m.DailyRate, m.MileageLimitPerDay, m.ExtraMileageFee, m.TotalCost, m.Status, id,
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

func (mod *RentalModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE rentals SET %s WHERE rental_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedRentalSorts = map[string]string{

	"rental_id": "rental_id",

	"vehicle_id": "vehicle_id",

	"customer_id": "customer_id",

	"pickup_location_id": "pickup_location_id",

	"dropoff_location_id": "dropoff_location_id",

	"pickup_datetime": "pickup_datetime",

	"dropoff_datetime": "dropoff_datetime",

	"start_mileage": "start_mileage",

	"end_mileage": "end_mileage",

	"start_fuel_level": "start_fuel_level",

	"end_fuel_level": "end_fuel_level",

	"daily_rate": "daily_rate",

	"mileage_limit_per_day": "mileage_limit_per_day",

	"extra_mileage_fee": "extra_mileage_fee",

	"total_cost": "total_cost",

	"status": "status",

	"created_at": "created_at",
}

func (mod *RentalModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*Rental, int, error) {

	sortCol, ok := allowedRentalSorts[sort]
	if !ok {
		sortCol = "rental_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM rentals").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT rental_id, vehicle_id, customer_id, pickup_location_id, dropoff_location_id, pickup_datetime, dropoff_datetime, start_mileage, end_mileage, start_fuel_level, end_fuel_level, daily_rate, mileage_limit_per_day, extra_mileage_fee, total_cost, status, created_at FROM rentals ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Rental
	for rows.Next() {
		var m Rental

		if err := rows.Scan(&m.ID, &m.VehicleID, &m.CustomerID, &m.PickupLocationID, &m.DropoffLocationID, &m.PickupDatetime, &m.DropoffDatetime, &m.StartMileage, &m.EndMileage, &m.StartFuelLevel, &m.EndFuelLevel, &m.DailyRate, &m.MileageLimitPerDay, &m.ExtraMileageFee, &m.TotalCost, &m.Status, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *RentalModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM rentals WHERE rental_id = $1`, id)
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
