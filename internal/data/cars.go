package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (m *VehicleModel) Update(id int64, v *Vehicle) error {
	if err := v.Validate(); err != nil {
		return err
	}
	query := `UPDATE vehicles SET vin=$1, make=$2, model=$3, year=$4, category=$5, daily_rate=$6, mileage=$7, fuel_capacity=$8, current_location_id=$9, status=$10 WHERE vehicle_id=$11`
	res, err := m.DB.ExecContext(context.Background(), query,
		v.VIN, v.Make, v.Model, v.Year, v.Category, v.DailyRate, v.Mileage, v.FuelCapacity, v.CurrentLocationID, v.Status, id,
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

// PartialUpdate updates only provided fields (PATCH)
func (m *VehicleModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE vehicles SET %s WHERE vehicle_id=$%d", strings.Join(setParts, ", "), i)
	res, err := m.DB.ExecContext(context.Background(), query, args...)
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

type Vehicle struct {
	ID                int64     `json:"id" db:"vehicle_id"`
	VIN               string    `json:"vin" db:"vin"`
	Make              string    `json:"make" db:"make"`
	Model             string    `json:"model" db:"model"`
	Year              int       `json:"year" db:"year"`
	Category          string    `json:"category" db:"category"`
	DailyRate         float64   `json:"daily_rate" db:"daily_rate"`
	Mileage           int       `json:"mileage" db:"mileage"`
	FuelCapacity      float64   `json:"fuel_capacity" db:"fuel_capacity"`
	CurrentLocationID *int      `json:"current_location_id,omitempty" db:"current_location_id"`
	Status            string    `json:"status" db:"status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type VehicleModel struct {
	DB *sql.DB
}

func (v *Vehicle) Validate() error {
	if len(strings.TrimSpace(v.VIN)) < 5 {
		return errors.New("vin must be at least 5 characters")
	}
	if v.Year < 1886 || v.Year > time.Now().Year()+1 {
		return fmt.Errorf("year must be between 1886 and %d", time.Now().Year()+1)
	}
	if v.Make == "" {
		return errors.New("make is required")
	}
	if v.Model == "" {
		return errors.New("model is required")
	}
	if v.Status != "available" && v.Status != "rented" && v.Status != "maintenance" && v.Status != "inactive" && v.Status != "" {
		return errors.New("invalid status")
	}
	return nil
}

func (m *VehicleModel) Insert(vehicle *Vehicle) error {
	if err := vehicle.Validate(); err != nil {
		return err
	}
	query := `INSERT INTO vehicles (vin, make, model, year, category, daily_rate, mileage, fuel_capacity, current_location_id, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING vehicle_id, created_at`
	return m.DB.QueryRowContext(context.Background(), query,
		vehicle.VIN, vehicle.Make, vehicle.Model, vehicle.Year, vehicle.Category,
		vehicle.DailyRate, vehicle.Mileage, vehicle.FuelCapacity, vehicle.CurrentLocationID, vehicle.Status,
	).Scan(&vehicle.ID, &vehicle.CreatedAt)
}

func (m *VehicleModel) Get(id int64) (*Vehicle, error) {
	var v Vehicle
	query := `SELECT vehicle_id, vin, make, model, year, category, daily_rate, mileage, fuel_capacity, current_location_id, status, created_at
		 FROM vehicles WHERE vehicle_id = $1`
	err := m.DB.QueryRowContext(context.Background(), query, id).Scan(
		&v.ID, &v.VIN, &v.Make, &v.Model, &v.Year, &v.Category, &v.DailyRate, &v.Mileage, &v.FuelCapacity, &v.CurrentLocationID, &v.Status, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

var allowedSorts = map[string]string{
	"id":            "vehicle_id",
	"vin":           "vin",
	"make":          "make",
	"model":         "model",
	"year":          "year",
	"category":      "category",
	"daily_rate":    "daily_rate",
	"mileage":       "mileage",
	"fuel_capacity": "fuel_capacity",
	"status":        "status",
	"created_at":    "created_at",
}

func (m *VehicleModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*Vehicle, int, error) {
	// Validate sort
	sortCol, ok := allowedSorts[sort]
	if !ok {
		sortCol = "vehicle_id"
	}
	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	// Get total count
	var total int  
	err := m.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM vehicles").Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	query := `SELECT vehicle_id, vin, make, model, year, category, daily_rate, mileage, fuel_capacity, current_location_id, status, created_at FROM vehicles ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`
	rows, err := m.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var vehicles []*Vehicle
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.VIN, &v.Make, &v.Model, &v.Year, &v.Category, &v.DailyRate, &v.Mileage, &v.FuelCapacity, &v.CurrentLocationID, &v.Status, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		vehicles = append(vehicles, &v)
	}
	return vehicles, total, nil
}

func (m *VehicleModel) Delete(id int64) error {
	res, err := m.DB.ExecContext(context.Background(), `DELETE FROM vehicles WHERE vehicle_id = $1`, id)
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
