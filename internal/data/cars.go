package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Car struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type CarModel struct {
	DB *sql.DB
}

func (c *Car) Validate() error {
	if strings.TrimSpace(c.Content) == "" {
		return errors.New("content must not be empty")
	}
	if strings.TrimSpace(c.Author) == "" {
		return errors.New("author must not be empty")
	}
	return nil
}

func (m *CarModel) Insert(car *Car) error {
	if err := car.Validate(); err != nil {
		return err
	}
	query := `INSERT INTO cars (content, author) VALUES ($1, $2) RETURNING id`
	return m.DB.QueryRowContext(context.Background(), query, car.Content, car.Author).Scan(&car.ID)
}

func (m *CarModel) Get(id int64) (*Car, error) {
	var car Car
	query := `SELECT id, content, author FROM cars WHERE id = $1`
	err := m.DB.QueryRowContext(context.Background(), query, id).Scan(&car.ID, &car.Content, &car.Author)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &car, nil
}

func (m *CarModel) GetAll() ([]*Car, error) {
	query := `SELECT id, content, author FROM cars`
	rows, err := m.DB.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cars []*Car
	for rows.Next() {
		var car Car
		if err := rows.Scan(&car.ID, &car.Content, &car.Author); err != nil {
			return nil, err
		}
		cars = append(cars, &car)
	}
	return cars, nil
}

func (m *CarModel) Delete(id int64) error {
	_, err := m.DB.ExecContext(context.Background(), `DELETE FROM cars WHERE id = $1`, id)
	return err
}
