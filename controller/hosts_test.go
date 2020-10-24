package controller_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hamburghammer/gsave/controller"
	"github.com/hamburghammer/gsave/db"
	"github.com/stretchr/testify/assert"
)

func TestHostsRouter_GetHosts(t *testing.T) {
	t.Run("db has one item", func(t *testing.T) {
		stats := []db.HostInfo{
			{Hostname: "foo"},
		}
		hostDB := &MockHostDB{}
		hostDB.SetHosts(stats)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHosts)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var gotBody []db.HostInfo
		json.Unmarshal(rr.Body.Bytes(), &gotBody)

		assert.Equal(t, stats, gotBody)
	})

	t.Run("db is empty", func(t *testing.T) {
		stats := []db.HostInfo{}
		hostDB := &MockHostDB{}
		hostDB.SetHosts(stats)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHosts)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var gotBody []db.HostInfo
		json.Unmarshal(rr.Body.Bytes(), &gotBody)

		assert.Equal(t, stats, gotBody)
	})

	t.Run("db returns hosts not found error", func(t *testing.T) {
		hostDB := &MockHostDB{}
		hostDB.SetHostsError(db.ErrHostsNotFound)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHosts)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		wantBody := fmt.Sprintf("%s\n", db.ErrHostsNotFound.Error())
		assert.Equal(t, wantBody, rr.Body.String())
	})

	t.Run("db returns all entities skipped error", func(t *testing.T) {
		hostDB := &MockHostDB{}
		hostDB.SetHostsError(db.ErrAllEntriesSkipped)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHosts)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		wantBody := fmt.Sprintf("%s\n", db.ErrAllEntriesSkipped.Error())
		assert.Equal(t, wantBody, rr.Body.String())
	})

	t.Run("db returns unknown error", func(t *testing.T) {
		unknownErr := errors.New("some error")
		hostDB := &MockHostDB{}
		hostDB.SetHostsError(unknownErr)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/hosts", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHosts)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		wantBody := fmt.Sprintf("%s\n", unknownErr.Error())
		assert.Equal(t, wantBody, rr.Body.String())
	})

	t.Run("pagination", func(t *testing.T) {
		t.Run("default pagination has a limit of 10", func(t *testing.T) {
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/hosts", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetHosts)
			handler.ServeHTTP(rr, req)

			assert.NotEqual(t, db.Pagination{}, hostDB.GetPagination())
			assert.Equal(t, 10, hostDB.GetPagination().Limit)
		})

		t.Run("default pagination has a skip of 0", func(t *testing.T) {
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/hosts", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetHosts)
			handler.ServeHTTP(rr, req)

			assert.NotEqual(t, db.Pagination{}, hostDB.GetPagination())
			assert.Equal(t, 0, hostDB.GetPagination().Skip)
		})

		t.Run("sets custom limit", func(t *testing.T) {
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/hosts?limit=2", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetHosts)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			assert.NotEqual(t, db.Pagination{}, hostDB.GetPagination())
			assert.Equal(t, 2, hostDB.GetPagination().Limit)
		})

		t.Run("sets negative limit", func(t *testing.T) {
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/hosts?limit=-2", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetHosts)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "No negative number allowed for the query param 'limit'\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})

		t.Run("sets negative skip", func(t *testing.T) {
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/hosts?skip=-2", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetHosts)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "No negative number allowed for the query param 'skip'\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})

		t.Run("sets skip to not a number", func(t *testing.T) {
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/hosts?skip=a", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetHosts)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "Query param 'skip' expected to be a number: a is not a number\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})

		t.Run("sets limit to not a number", func(t *testing.T) {
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/hosts?limit=a", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetHosts)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "Query param 'limit' expected to be a number: a is not a number\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})

	})
}

type MockHostDB struct {
	hosts      []db.HostInfo
	hostsError error

	host      db.HostInfo
	hostError error

	stats      []db.Stats
	statsError error

	insertedStat      db.Stats
	insertedStatError error

	pagination db.Pagination
}

// GetHosts
func (m *MockHostDB) SetHosts(hosts []db.HostInfo) {
	m.hosts = hosts
}
func (m *MockHostDB) SetHostsError(err error) {
	m.hostsError = err
}
func (m *MockHostDB) GetHosts(pagination db.Pagination) ([]db.HostInfo, error) {
	m.pagination = pagination
	if m.hostsError != nil {
		return []db.HostInfo{}, m.hostsError
	}
	return m.hosts, nil
}

// GetHost
func (m *MockHostDB) SetHost(host db.HostInfo) {
	m.host = host
}
func (m *MockHostDB) SetHostError(err error) {
	m.hostError = err
}
func (m *MockHostDB) GetHost(hostname string) (db.HostInfo, error) {
	if m.hostError != nil {
		return db.HostInfo{}, m.hostError
	}
	return m.host, nil
}

// GetStatsByHostname
func (m *MockHostDB) SetStatsByHostname(stats []db.Stats) {
	m.stats = stats
}
func (m *MockHostDB) SetStatsByHostnameError(err error) {
	m.statsError = err
}
func (m *MockHostDB) GetStatsByHostname(hostname string, pagination db.Pagination) ([]db.Stats, error) {
	if m.statsError != nil {
		return []db.Stats{}, m.hostError
	}
	return m.stats, nil
}

// InsertStats
func (m *MockHostDB) GetInsertedStats() db.Stats {
	return m.insertedStat
}
func (m *MockHostDB) SetInsertStatsError(err error) {
	m.insertedStatError = err
}
func (m *MockHostDB) InsertStats(hostname string, stats db.Stats) error {
	if m.statsError != nil {
		return m.hostError
	}
	return nil
}

func (m *MockHostDB) GetPagination() db.Pagination {
	return m.pagination
}
