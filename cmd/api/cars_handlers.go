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
	vehicleModel := data.VehicleModel{DB: a.getDB()}
	vehicles, err := vehicleModel.GetAll()
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	resp := envelope{"vehicles": vehicles}
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
