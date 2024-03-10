package server

import (
	"autotools-golang-api/kubecloudsinc/handler" // Adjust this import path to your project structure
	"autotools-golang-api/kubecloudsinc/middleware"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Initialize and return a new HTTP router
func NewRouter(app *newrelic.Application) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.NewRelicMiddleware(app))

	r.HandleFunc("/v2/login", middleware.Login).Methods("POST")
	r.HandleFunc("/v2/employees", middleware.IsAuthorized("admin", "editor", "viewer")(handler.GetEmployees)).Methods("GET")
	r.HandleFunc("/v2/employee", middleware.IsAuthorized("admin", "editor", "viewer")(handler.GetEmployee)).Methods("GET")
	r.HandleFunc("/v2/employee/{employeeId}", middleware.IsAuthorized("admin", "editor", "viewer")(handler.GetEmployeeProfile)).Methods("GET")
	r.HandleFunc("/v2/employee", middleware.IsAuthorized("admin", "editor")(handler.AddEmployee)).Methods("POST")
	r.HandleFunc("/v2/employee/{employeeId}", middleware.IsAuthorized("admin", "editor")(handler.UpdateEmployee)).Methods("PUT")
	r.HandleFunc("/v2/employee/{employeeId}", middleware.IsAuthorized("admin")(handler.DeleteEmployee)).Methods("DELETE")

	// Manually register pprof handlers
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Register other pprof handlers
	r.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)

	return r
}

// StartServer starts the HTTP server on a specified port
func StartServer(port string, app *newrelic.Application) error {
	r := NewRouter(app)
	http.Handle("/", r)
	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(port, r)
}
