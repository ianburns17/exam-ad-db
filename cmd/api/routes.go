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

	router.HandlerFunc(http.MethodPost, "/v1/cars", a.createCarHandler)
	router.HandlerFunc(http.MethodGet, "/v1/cars/:id", a.getCarHandler)
	router.HandlerFunc(http.MethodGet, "/v1/cars", a.listCarsHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/cars/:id", a.deleteCarHandler)

	return a.recoverPanic(router)

}
