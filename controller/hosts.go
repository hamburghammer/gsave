package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	subrouter.HandleFunc("", hr.getHosts).Methods(http.MethodGet).Name("GetHosts")
	subrouter.HandleFunc("/{hostname}", hr.getHost).Methods(http.MethodGet).Name("GetHost")
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

func (hr *HostsRouter) getHosts(w http.ResponseWriter, r *http.Request) {
	pagination, err := hr.getSkipAndLimit(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hosts, err := hr.db.GetHosts(pagination)
	if err != nil {
		if errors.Is(err, db.ErrHostsNotFound) || errors.Is(err, db.ErrAllEntriesSkipped) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hosts)
}

func (hr *HostsRouter) getHost(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]
	if hostname == "" {
		http.Error(w, fmt.Sprintf("Missing hostname: '%s' is not a valid hostname\n", hostname), http.StatusBadRequest)
		return
	}

	host, err := hr.db.GetHost(hostname)
	if err != nil {
		if errors.Is(err, db.ErrHostNotFound) {
			http.Error(w, fmt.Sprintf("No host with the name '%s' found\n", hostname), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(host)
}

func (hr *HostsRouter) getStats(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]

	pagination, err := hr.getSkipAndLimit(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stats, err := hr.db.GetStatsByHostname(hostname, pagination)
	if err != nil {
		if errors.Is(err, db.ErrHostsNotFound) || errors.Is(err, db.ErrAllEntriesSkipped) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		log.Println(err)
		return
	}
	log.Printf("Received stat: %+v", stats)

	err = hr.db.InsertStats(hostname, stats)
	if err != nil {
		http.Error(w, "Something with the DB went wrong.", http.StatusInternalServerError)
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

	strSkip := r.FormValue("skip")
	if strSkip == "" {
		strSkip = defaultSkip
	}
	skip, err := strconv.ParseInt(strSkip, 10, 64)
	if err != nil {
		return db.Pagination{}, fmt.Errorf("Query param 'skip' expected to be a number: %s is not a number", strSkip)
	}

	return db.Pagination{Skip: int(skip), Limit: int(limit)}, nil
}
