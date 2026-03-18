package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// Healthcheck
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/vehicles", a.createVehicleHandler)
	router.HandlerFunc(http.MethodGet, "/v1/vehicles/:id", a.getVehicleHandler)
	router.HandlerFunc(http.MethodGet, "/v1/vehicles", a.listVehiclesHandler)
	router.HandlerFunc(http.MethodPut, "/v1/vehicles/:id", a.updateVehicleHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/vehicles/:id", a.patchVehicleHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/vehicles/:id", a.deleteVehicleHandler)

	// Chain rate limiting and panic recovery middleware
	return a.recoverPanic(a.rateLimit(router))
}
