package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"final/internal/data"
)

func (a *applicationDependencies) updateVehicleHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	var input data.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	vehicleModel := data.VehicleModel{DB: a.getDB()}
	if err := vehicleModel.Update(id, &input); err != nil {
		if err == sql.ErrNoRows {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}
	// Return the updated vehicle
	vehicle, err := vehicleModel.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	resp := envelope{"vehicle": vehicle}
	a.writeJSON(w, http.StatusOK, resp, nil)
}

// PATCH /v1/vehicles/:id
func (a *applicationDependencies) patchVehicleHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	vehicleModel := data.VehicleModel{DB: a.getDB()}
	if err := vehicleModel.PartialUpdate(id, updates); err != nil {
		if err == sql.ErrNoRows {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}
	// Return the updated vehicle
	vehicle, err := vehicleModel.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	resp := envelope{"vehicle": vehicle}
	a.writeJSON(w, http.StatusOK, resp, nil)
}

// POST /v1/vehicles
func (a *applicationDependencies) createVehicleHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := input.Validate(); err != nil {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, err.Error())
		return
	}
	vehicleModel := data.VehicleModel{DB: a.getDB()}
	if err := vehicleModel.Insert(&input); err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	resp := envelope{"vehicle": input}
	a.writeJSON(w, http.StatusCreated, resp, nil)
}

// GET /v1/vehicles/:id

func (a *applicationDependencies) getVehicleHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	vehicleModel := data.VehicleModel{DB: a.getDB()}
	vehicle, err := vehicleModel.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if vehicle == nil {
		a.notFoundResponse(w, r)
		return
	}
	resp := envelope{"vehicle": vehicle}
	a.writeJSON(w, http.StatusOK, resp, nil)
}

// GET /v1/vehicles
func (a *applicationDependencies) listVehiclesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query params for pagination and sorting
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	sort := q.Get("sort")
	if sort == "" {
		sort = "id"
	}
	direction := q.Get("direction")
	if direction != "desc" {
		direction = "asc"
	}
	vehicleModel := data.VehicleModel{DB: a.getDB()}
	vehicles, total, err := vehicleModel.GetAllPaginated(page, pageSize, sort, direction)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	resp := envelope{
		"vehicles": vehicles,
		"pagination": map[string]any{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	}
	a.writeJSON(w, http.StatusOK, resp, nil)
}

// DELETE /v1/vehicles/:id
func (a *applicationDependencies) deleteVehicleHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	idStr := params.ByName("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	vehicleModel := data.VehicleModel{DB: a.getDB()}
	err = vehicleModel.Delete(id)
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

// getDB returns the application's *sql.DB connection
func (a *applicationDependencies) getDB() *sql.DB {
	return a.db
}
