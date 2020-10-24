package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hamburghammer/gsave/db"
)

// NewHostsRouter is a constructor for the HostsRouter.
func NewHostsRouter(db db.HostDB) *HostsRouter {
	return &HostsRouter{db: db}
}

// HostsRouter represents the controller for the hosts routes.
type HostsRouter struct {
	subrouter *mux.Router
	db        db.HostDB
}

// Register registers all routes to the given subrouter.
func (hr *HostsRouter) Register(subrouter *mux.Router) {
	hr.subrouter = subrouter
	subrouter.HandleFunc("", hr.GetHosts).Methods(http.MethodGet).Name("GetHosts")
	subrouter.HandleFunc("/{hostname}", hr.GetHost).Methods(http.MethodGet).Name("GetHost")
	subrouter.HandleFunc("/{hostname}/stats", hr.getStats).Methods(http.MethodGet).Name("GetStats")
	subrouter.HandleFunc("/{hostname}/stats", hr.postStats).Methods(http.MethodPost).Name("PostStats")
}

// GetPrefix returns the the pre route for this controller.
func (hr *HostsRouter) GetPrefix() string {
	return "/hosts"
}

// GetRouteName returns the Name of this controller.
func (hr *HostsRouter) GetRouteName() string {
	return "Hosts"
}

// GetHosts is a HandleFunc to get hosts out of the db with optional pagination as query params.
func (hr *HostsRouter) GetHosts(w http.ResponseWriter, r *http.Request) {
	pagination, err := hr.getSkipAndLimit(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logBadRequest.Error(err)
		return
	}

	hosts, err := hr.db.GetHosts(pagination)
	if err != nil {
		if errors.Is(err, db.ErrHostsNotFound) || errors.Is(err, db.ErrAllEntriesSkipped) {
			http.Error(w, err.Error(), http.StatusNotFound)
			logNotFound.Error(err)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logInternalServerError.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hosts)
}

// GetHost is a HandleFunc to get one host. The host name gets read out of the request path.
func (hr *HostsRouter) GetHost(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]

	host, err := hr.db.GetHost(hostname)
	if err != nil {
		if errors.Is(err, db.ErrHostNotFound) {
			http.Error(w, fmt.Sprintf("No host with the name '%s' found", hostname), http.StatusNotFound)
			logNotFound.Error(err)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logInternalServerError.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(host)
}

func (hr *HostsRouter) getStats(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]

	pagination, err := hr.getSkipAndLimit(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logBadRequest.Error(err)
		return
	}

	stats, err := hr.db.GetStatsByHostname(hostname, pagination)
	if err != nil {
		if errors.Is(err, db.ErrHostNotFound) || errors.Is(err, db.ErrAllEntriesSkipped) {
			http.Error(w, err.Error(), http.StatusNotFound)
			logNotFound.Error(err)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logInternalServerError.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (hr *HostsRouter) postStats(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]

	var stats db.Stats
	err := json.NewDecoder(r.Body).Decode(&stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logBadRequest.Error(err)
		return
	}
	logPackage.Debugf("Received stat: %+v", stats)

	err = hr.db.InsertStats(hostname, stats)
	if err != nil {
		http.Error(w, "Something with the DB went wrong.", http.StatusInternalServerError)
		logInternalServerError.Error(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// getSkipAndLimit from the query of the request.
func (hr *HostsRouter) getSkipAndLimit(r *http.Request) (db.Pagination, error) {
	defaultLimit := "10"
	defaultSkip := "0"

	strLimit := r.FormValue("limit")
	if strLimit == "" {
		strLimit = defaultLimit
	}
	limit, err := strconv.ParseInt(strLimit, 10, 64)
	if err != nil {
		return db.Pagination{}, fmt.Errorf("Query param 'limit' expected to be a number: %s is not a number", strLimit)
	}
	if limit < 0 {
		return db.Pagination{}, fmt.Errorf("No negative number allowed for the query param 'limit'")
	}

	strSkip := r.FormValue("skip")
	if strSkip == "" {
		strSkip = defaultSkip
	}
	skip, err := strconv.ParseInt(strSkip, 10, 64)
	if err != nil {
		return db.Pagination{}, fmt.Errorf("Query param 'skip' expected to be a number: %s is not a number", strSkip)
	}
	if skip < 0 {
		return db.Pagination{}, fmt.Errorf("No negative number allowed for the query param 'skip'")
	}

	return db.Pagination{Skip: int(skip), Limit: int(limit)}, nil
}
