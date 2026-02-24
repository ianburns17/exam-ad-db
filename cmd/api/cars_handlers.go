package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"final/internal/data"
)

// POST /v1/cars
func (a *applicationDependencies) createCarHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Car
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := input.Validate(); err != nil {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, err.Error())
		return
	}
	carModel := data.CarModel{DB: a.getDB()}
	if err := carModel.Insert(&input); err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	resp := envelope{"car": input}
	a.writeJSON(w, http.StatusCreated, resp, nil)
}

// GET /v1/cars/:id
func (a *applicationDependencies) getCarHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	carModel := data.CarModel{DB: a.getDB()}
	car, err := carModel.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if car == nil {
		a.notFoundResponse(w, r)
		return
	}
	resp := envelope{"car": car}
	a.writeJSON(w, http.StatusOK, resp, nil)
}

// GET /v1/cars
func (a *applicationDependencies) listCarsHandler(w http.ResponseWriter, r *http.Request) {
	carModel := data.CarModel{DB: a.getDB()}
	cars, err := carModel.GetAll()
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	resp := envelope{"cars": cars}
	a.writeJSON(w, http.StatusOK, resp, nil)
}

// DELETE /v1/cars/:id
func (a *applicationDependencies) deleteCarHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	carModel := data.CarModel{DB: a.getDB()}
	err = carModel.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getDB is a helper to get the *sql.DB from dependencies (stub, replace with your actual DB connection)
func (a *applicationDependencies) getDB() *sql.DB {
	// TODO: wire your actual DB connection here
	return nil
}
