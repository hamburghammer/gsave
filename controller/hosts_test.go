package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
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

func TestGetHost(t *testing.T) {
	t.Run("search with hostname from url", func(t *testing.T) {
		hostname := "foo"
		hostInfo := db.HostInfo{Hostname: hostname, DataPoints: 1}
		hostDB := &MockHostDB{}
		hostDB.SetHost(hostInfo)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname, nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHost)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, hostname, hostDB.GetHostHostname())
	})

	t.Run("db has an item", func(t *testing.T) {
		hostname := "foo"
		hostInfo := db.HostInfo{Hostname: hostname, DataPoints: 1}
		hostDB := &MockHostDB{}
		hostDB.SetHost(hostInfo)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname, nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHost)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var gotBody db.HostInfo
		err = json.NewDecoder(rr.Body).Decode(&gotBody)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, hostInfo, gotBody)
	})

	t.Run("db returns not found error", func(t *testing.T) {
		hostname := "foo"
		hostDB := &MockHostDB{}
		hostDB.SetHostError(db.ErrHostNotFound)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname, nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHost)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		wantErr := fmt.Sprintf("No host with the name '%s' found\n", hostname)
		assert.Equal(t, wantErr, rr.Body.String())
	})

	t.Run("db returns unknown error", func(t *testing.T) {
		unknownErr := errors.New("unknown error")
		hostname := "foo"
		hostDB := &MockHostDB{}
		hostDB.SetHostError(unknownErr)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname, nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetHost)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		wantErr := fmt.Sprintf("%s\n", unknownErr.Error())
		assert.Equal(t, wantErr, rr.Body.String())
	})
}

func TestGetStat(t *testing.T) {
	t.Run("pagination", func(t *testing.T) {
		t.Run("default pagination has a limit of 10", func(t *testing.T) {
			hostname := "foo"
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/"+hostname+"/stats", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetStats)
			handler.ServeHTTP(rr, req)

			assert.NotEqual(t, db.Pagination{}, hostDB.GetPagination())
			assert.Equal(t, 10, hostDB.GetPagination().Limit)
		})

		t.Run("default pagination has a skip of 0", func(t *testing.T) {
			hostname := "foo"
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/"+hostname+"/stats", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetStats)
			handler.ServeHTTP(rr, req)

			assert.NotEqual(t, db.Pagination{}, hostDB.GetPagination())
			assert.Equal(t, 0, hostDB.GetPagination().Skip)
		})

		t.Run("sets custom limit", func(t *testing.T) {
			hostname := "foo"
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/"+hostname+"/stats?limit=2", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetStats)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			assert.NotEqual(t, db.Pagination{}, hostDB.GetPagination())
			assert.Equal(t, 2, hostDB.GetPagination().Limit)
		})

		t.Run("sets negative limit", func(t *testing.T) {
			hostname := "foo"
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/"+hostname+"/stats?limit=-2", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetStats)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "No negative number allowed for the query param 'limit'\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})

		t.Run("sets negative skip", func(t *testing.T) {
			hostname := "foo"
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/"+hostname+"/stats?skip=-2", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetStats)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "No negative number allowed for the query param 'skip'\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})

		t.Run("sets skip to not a number", func(t *testing.T) {
			hostname := "foo"
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/"+hostname+"/stats?skip=a", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetStats)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "Query param 'skip' expected to be a number: a is not a number\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})

		t.Run("sets limit to not a number", func(t *testing.T) {
			hostname := "foo"
			hostDB := &MockHostDB{}
			hostsRouter := controller.NewHostsRouter(hostDB)

			req, err := http.NewRequest("GET", "/"+hostname+"/stats?limit=a", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(hostsRouter.GetStats)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)

			wantErr := "Query param 'limit' expected to be a number: a is not a number\n"
			assert.Equal(t, wantErr, rr.Body.String())
		})
	})

	t.Run("search with hostname from url", func(t *testing.T) {
		hostname := "foo"
		hostInfo := db.HostInfo{Hostname: hostname, DataPoints: 1}
		hostDB := &MockHostDB{}
		hostDB.SetHost(hostInfo)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname+"/stats", nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, hostname, hostDB.GetStatsByHostnameHostname())
	})

	t.Run("db has an item", func(t *testing.T) {
		hostname := "foo"
		stats := []db.Stats{{Hostname: hostname}}
		hostDB := &MockHostDB{}
		hostDB.SetStatsByHostname(stats)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname+"/stats", nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, hostname, hostDB.GetStatsByHostnameHostname())

		var gotBody []db.Stats
		err = json.NewDecoder(rr.Body).Decode(&gotBody)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, stats, gotBody)
	})

	t.Run("db returns not found error", func(t *testing.T) {
		hostname := "foo"
		hostDB := &MockHostDB{}
		hostDB.SetStatsByHostnameError(db.ErrHostNotFound)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname+"/stats", nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		wantErr := fmt.Sprintf("No host with the name '%s' found\n", hostname)
		assert.Equal(t, wantErr, rr.Body.String())
	})

	t.Run("db returns all entries skipped error", func(t *testing.T) {
		hostname := "foo"
		hostDB := &MockHostDB{}
		hostDB.SetStatsByHostnameError(db.ErrAllEntriesSkipped)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname+"/stats", nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		wantErr := "db: All entries skipped\n"
		assert.Equal(t, wantErr, rr.Body.String())
	})

	t.Run("db returns unknown error", func(t *testing.T) {
		unknownErr := errors.New("unknown error")
		hostname := "foo"
		hostDB := &MockHostDB{}
		hostDB.SetStatsByHostnameError(unknownErr)
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("GET", "/"+hostname+"/stats", nil)
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.GetStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		wantErr := fmt.Sprintf("%s\n", unknownErr.Error())
		assert.Equal(t, wantErr, rr.Body.String())
	})
}

func TestInsertStat(t *testing.T) {
	t.Run("insert new stat into the db", func(t *testing.T) {
		hostname := "foo"
		stat := db.Stats{Hostname: hostname}
		hostDB := &MockHostDB{}
		hostsRouter := controller.NewHostsRouter(hostDB)

		requestBody, _ := json.Marshal(stat)
		req, err := http.NewRequest("POST", "/"+hostname+"/stats", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.PostStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, hostname, hostDB.GetInsertStatsHostname())

		assert.Equal(t, stat, hostDB.GetInsertedStats())
	})

	t.Run("missing body", func(t *testing.T) {
		hostname := "foo"
		hostDB := &MockHostDB{}
		hostsRouter := controller.NewHostsRouter(hostDB)

		req, err := http.NewRequest("POST", "/"+hostname+"/stats", bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.PostStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "Could not read the body\n", rr.Body.String())
	})

	t.Run("db returns an unknown error", func(t *testing.T) {
		hostname := "foo"
		unknownErr := errors.New("unknown error")
		stat := db.Stats{Hostname: hostname}
		hostDB := &MockHostDB{}
		hostDB.SetInsertStatsError(unknownErr)
		hostsRouter := controller.NewHostsRouter(hostDB)

		requestBody, _ := json.Marshal(stat)
		req, err := http.NewRequest("POST", "/"+hostname+"/stats", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		req = mux.SetURLVars(req, map[string]string{"hostname": hostname})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(hostsRouter.PostStats)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "Something with the DB went wrong.\n", rr.Body.String())
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
	hostname   string
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
func (m *MockHostDB) GetHostHostname() string {
	return m.hostname
}
func (m *MockHostDB) GetHost(hostname string) (db.HostInfo, error) {
	m.hostname = hostname
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
func (m *MockHostDB) GetStatsByHostnameHostname() string {
	return m.hostname
}
func (m *MockHostDB) GetStatsByHostname(hostname string, pagination db.Pagination) ([]db.Stats, error) {
	m.hostname = hostname
	m.pagination = pagination
	if m.statsError != nil {
		return []db.Stats{}, m.statsError
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
func (m *MockHostDB) GetInsertStatsHostname() string {
	return m.hostname
}
func (m *MockHostDB) InsertStats(hostname string, stats db.Stats) error {
	m.hostname = hostname
	m.insertedStat = stats
	if m.insertedStatError != nil {
		return m.insertedStatError
	}
	return nil
}

func (m *MockHostDB) GetPagination() db.Pagination {
	return m.pagination
}
