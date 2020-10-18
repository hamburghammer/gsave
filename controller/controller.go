package controller

import "github.com/gorilla/mux"

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
