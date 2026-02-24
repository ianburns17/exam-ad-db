package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	// setup a new router
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	// setup routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/comments", a.createCommentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/cars/read", a.readCarsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/cars/write", a.writeCarsHandler)

	return a.recoverPanic(router)

}
