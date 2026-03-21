package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"final/internal/data"
)

func (a *applicationDependencies) createFeeHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Fee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := input.Validate(); err != nil {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, err.Error())
		return
	}
	mod := data.FeeModel{DB: a.getDB()}
	if err := mod.Insert(&input); err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	a.writeJSON(w, http.StatusCreated, envelope{"fees": input}, nil)
}

func (a *applicationDependencies) getFeeHandler(w http.ResponseWriter, r *http.Request) {
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
	mod := data.FeeModel{DB: a.getDB()}
	item, err := mod.Get(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if item == nil {
		a.notFoundResponse(w, r)
		return
	}
	a.writeJSON(w, http.StatusOK, envelope{"fees": item}, nil)
}

func (a *applicationDependencies) updateFeeHandler(w http.ResponseWriter, r *http.Request) {
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
	var input data.Fee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}
	mod := data.FeeModel{DB: a.getDB()}
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
	a.writeJSON(w, http.StatusOK, envelope{"fees": item}, nil)
}

func (a *applicationDependencies) patchFeeHandler(w http.ResponseWriter, r *http.Request) {
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
	mod := data.FeeModel{DB: a.getDB()}
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
	a.writeJSON(w, http.StatusOK, envelope{"fees": item}, nil)
}

func (a *applicationDependencies) deleteFeeHandler(w http.ResponseWriter, r *http.Request) {
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
	mod := data.FeeModel{DB: a.getDB()}
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

func (a *applicationDependencies) listFeesHandler(w http.ResponseWriter, r *http.Request) {
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
	mod := data.FeeModel{DB: a.getDB()}
	items, total, err := mod.GetAllPaginated(page, pageSize, sort, direction)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	a.writeJSON(w, http.StatusOK, envelope{
		"feess": items,
		"pagination": map[string]any{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	}, nil)
}
