package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Location struct {
	ID        int64     `json:"id" db:"location_id"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	City      string    `json:"city" db:"city"`
	State     string    `json:"state" db:"state"`
	Country   string    `json:"country" db:"country"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LocationModel struct {
	DB *sql.DB
}

func (m *Location) Validate() error {
	return nil
}

func (mod *LocationModel) Insert(m *Location) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `INSERT INTO locations (name, address, city, state, country)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING location_id, created_at`
	return mod.DB.QueryRowContext(context.Background(), query,
		m.Name, m.Address, m.City, m.State, m.Country,
	).Scan(&m.ID, &m.CreatedAt)
}

func (mod *LocationModel) Get(id int64) (*Location, error) {
	var m Location

	query := `SELECT location_id, name, address, city, state, country, created_at FROM locations WHERE location_id = $1`
	err := mod.DB.QueryRowContext(context.Background(), query, id).Scan(
		&m.ID, &m.Name, &m.Address, &m.City, &m.State, &m.Country, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mod *LocationModel) Update(id int64, m *Location) error {
	if err := m.Validate(); err != nil {
		return err
	}

	query := `UPDATE locations SET name=$1, address=$2, city=$3, state=$4, country=$5 WHERE location_id=$6`
	res, err := mod.DB.ExecContext(context.Background(), query,
		m.Name, m.Address, m.City, m.State, m.Country, id,
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

func (mod *LocationModel) PartialUpdate(id int64, updates map[string]interface{}) error {
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
	query := fmt.Sprintf("UPDATE locations SET %s WHERE location_id=$%d", strings.Join(setParts, ", "), i)
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

var allowedLocationSorts = map[string]string{

	"location_id": "location_id",

	"name": "name",

	"address": "address",

	"city": "city",

	"state": "state",

	"country": "country",

	"created_at": "created_at",
}

func (mod *LocationModel) GetAllPaginated(page, pageSize int, sort, direction string) ([]*Location, int, error) {

	sortCol, ok := allowedLocationSorts[sort]
	if !ok {
		sortCol = "location_id"
	}

	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	offset := (page - 1) * pageSize
	var total int

	err := mod.DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM locations").Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `SELECT location_id, name, address, city, state, country, created_at FROM locations ORDER BY ` + sortCol + ` ` + dir + ` LIMIT $1 OFFSET $2`

	rows, err := mod.DB.QueryContext(context.Background(), query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*Location
	for rows.Next() {
		var m Location

		if err := rows.Scan(&m.ID, &m.Name, &m.Address, &m.City, &m.State, &m.Country, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (mod *LocationModel) Delete(id int64) error {
	res, err := mod.DB.ExecContext(context.Background(), `DELETE FROM locations WHERE location_id = $1`, id)
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
