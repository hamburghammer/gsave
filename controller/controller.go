package controller

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	logPackage             = log.WithField("Package", "controller")
	logRequestError        = logPackage.WithField("RequestStatus", "Error")
	logBadRequest          = logRequestError.WithField("StatusCode", http.StatusBadRequest)
	logNotFound            = logRequestError.WithField("StatusCode", http.StatusNotFound)
	logInternalServerError = logRequestError.WithField("StatusCode", http.StatusInternalServerError)
)

// Router is an interface that should be implemented by any controller
// to give some information and to register the routes.
type Router interface {
	// GetPrefix returns the path prefix for the router.
	GetPrefix() string
	// GetRouteName returns the name that this router route should have.
	GetRouteName() string
	// Register should register all routes of an controller to a subrouter.
	Register(subrouter *mux.Router)
}
