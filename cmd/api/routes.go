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

	// Locations
	router.HandlerFunc(http.MethodPost, "/v1/locations", a.createLocationHandler)
	router.HandlerFunc(http.MethodGet, "/v1/locations/:id", a.getLocationHandler)
	router.HandlerFunc(http.MethodGet, "/v1/locations", a.listLocationsHandler)
	router.HandlerFunc(http.MethodPut, "/v1/locations/:id", a.updateLocationHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/locations/:id", a.patchLocationHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/locations/:id", a.deleteLocationHandler)

	// Customers
	router.HandlerFunc(http.MethodPost, "/v1/customers", a.createCustomerHandler)
	router.HandlerFunc(http.MethodGet, "/v1/customers/:id", a.getCustomerHandler)
	router.HandlerFunc(http.MethodGet, "/v1/customers", a.listCustomersHandler)
	router.HandlerFunc(http.MethodPut, "/v1/customers/:id", a.updateCustomerHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/customers/:id", a.patchCustomerHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/customers/:id", a.deleteCustomerHandler)

	// Customer Violations
	router.HandlerFunc(http.MethodPost, "/v1/customer_violations", a.createCustomerViolationHandler)
	router.HandlerFunc(http.MethodGet, "/v1/customer_violations/:id", a.getCustomerViolationHandler)
	router.HandlerFunc(http.MethodGet, "/v1/customer_violations", a.listCustomerViolationsHandler)
	router.HandlerFunc(http.MethodPut, "/v1/customer_violations/:id", a.updateCustomerViolationHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/customer_violations/:id", a.patchCustomerViolationHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/customer_violations/:id", a.deleteCustomerViolationHandler)

	// Rentals
	router.HandlerFunc(http.MethodPost, "/v1/rentals", a.createRentalHandler)
	router.HandlerFunc(http.MethodGet, "/v1/rentals/:id", a.getRentalHandler)
	router.HandlerFunc(http.MethodGet, "/v1/rentals", a.listRentalsHandler)
	router.HandlerFunc(http.MethodPut, "/v1/rentals/:id", a.updateRentalHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/rentals/:id", a.patchRentalHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/rentals/:id", a.deleteRentalHandler)

	// Maintenance Records
	router.HandlerFunc(http.MethodPost, "/v1/maintenance_records", a.createMaintenanceRecordHandler)
	router.HandlerFunc(http.MethodGet, "/v1/maintenance_records/:id", a.getMaintenanceRecordHandler)
	router.HandlerFunc(http.MethodGet, "/v1/maintenance_records", a.listMaintenanceRecordsHandler)
	router.HandlerFunc(http.MethodPut, "/v1/maintenance_records/:id", a.updateMaintenanceRecordHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/maintenance_records/:id", a.patchMaintenanceRecordHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/maintenance_records/:id", a.deleteMaintenanceRecordHandler)

	// Insurance Policies
	router.HandlerFunc(http.MethodPost, "/v1/insurance_policies", a.createInsurancePolicyHandler)
	router.HandlerFunc(http.MethodGet, "/v1/insurance_policies/:id", a.getInsurancePolicyHandler)
	router.HandlerFunc(http.MethodGet, "/v1/insurance_policies", a.listInsurancePolicysHandler)
	router.HandlerFunc(http.MethodPut, "/v1/insurance_policies/:id", a.updateInsurancePolicyHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/insurance_policies/:id", a.patchInsurancePolicyHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/insurance_policies/:id", a.deleteInsurancePolicyHandler)

	// Insurance Claims
	router.HandlerFunc(http.MethodPost, "/v1/insurance_claims", a.createInsuranceClaimHandler)
	router.HandlerFunc(http.MethodGet, "/v1/insurance_claims/:id", a.getInsuranceClaimHandler)
	router.HandlerFunc(http.MethodGet, "/v1/insurance_claims", a.listInsuranceClaimsHandler)
	router.HandlerFunc(http.MethodPut, "/v1/insurance_claims/:id", a.updateInsuranceClaimHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/insurance_claims/:id", a.patchInsuranceClaimHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/insurance_claims/:id", a.deleteInsuranceClaimHandler)

	// Inspections
	router.HandlerFunc(http.MethodPost, "/v1/inspections", a.createInspectionHandler)
	router.HandlerFunc(http.MethodGet, "/v1/inspections/:id", a.getInspectionHandler)
	router.HandlerFunc(http.MethodGet, "/v1/inspections", a.listInspectionsHandler)
	router.HandlerFunc(http.MethodPut, "/v1/inspections/:id", a.updateInspectionHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/inspections/:id", a.patchInspectionHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/inspections/:id", a.deleteInspectionHandler)

	// Fees
	router.HandlerFunc(http.MethodPost, "/v1/fees", a.createFeeHandler)
	router.HandlerFunc(http.MethodGet, "/v1/fees/:id", a.getFeeHandler)
	router.HandlerFunc(http.MethodGet, "/v1/fees", a.listFeesHandler)
	router.HandlerFunc(http.MethodPut, "/v1/fees/:id", a.updateFeeHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/fees/:id", a.patchFeeHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/fees/:id", a.deleteFeeHandler)

	// Expose /metrics endpoint
	router.HandlerFunc(http.MethodGet, "/metrics", a.metricsHandler)

	// Chain CORS, metrics, gzip, rate limiting, and panic recovery middleware
	return a.recoverPanic(
		a.rateLimit(
			a.gzipMiddleware(
				a.metricsMiddleware(
					a.corsMiddleware(router),
				),
			),
		),
	)
}
