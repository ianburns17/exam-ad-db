package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"final/internal/data"
)

func (a *applicationDependencies) createCustomerViolationHandler(w http.ResponseWriter, r *http.Request) {
	var input data.CustomerViolation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := input.Validate(); err != nil {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, err.Error())
		return
	}
	mod := data.CustomerViolationModel{DB: a.getDB()}
	if err := mod.Insert(&input); err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	a.writeJSON(w, http.StatusCreated, envelope{"customer_violations": input}, nil)
}

func (a *applicationDependencies) getCustomerViolationHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	mod := data.CustomerViolationModel{DB: a.getDB()}
	item, err := mod.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if item == nil {
		a.notFoundResponse(w, r)
		return
	}
	a.writeJSON(w, http.StatusOK, envelope{"customer_violations": item}, nil)
}

func (a *applicationDependencies) updateCustomerViolationHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	var input data.CustomerViolation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	mod := data.CustomerViolationModel{DB: a.getDB()}
	if err := mod.Update(id, &input); err != nil {
		if err == sql.ErrNoRows {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}
	item, err := mod.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	a.writeJSON(w, http.StatusOK, envelope{"customer_violations": item}, nil)
}

func (a *applicationDependencies) patchCustomerViolationHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	mod := data.CustomerViolationModel{DB: a.getDB()}
	if err := mod.PartialUpdate(id, updates); err != nil {
		if err == sql.ErrNoRows {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}
	item, err := mod.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	a.writeJSON(w, http.StatusOK, envelope{"customer_violations": item}, nil)
}

func (a *applicationDependencies) deleteCustomerViolationHandler(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid id")
		return
	}
	mod := data.CustomerViolationModel{DB: a.getDB()}
	err = mod.Delete(id)
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

func (a *applicationDependencies) listCustomerViolationsHandler(w http.ResponseWriter, r *http.Request) {
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
	mod := data.CustomerViolationModel{DB: a.getDB()}
	items, total, err := mod.GetAllPaginated(page, pageSize, sort, direction)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	a.writeJSON(w, http.StatusOK, envelope{
		"customer_violationss": items,
		"pagination": map[string]any{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	}, nil)
}
