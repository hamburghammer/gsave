package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hamburghammer/gsave/controller"
	"github.com/hamburghammer/gsave/db"
	"github.com/stretchr/testify/assert"
)

func TestHostsRouter_GetHosts_WithoutPagination(t *testing.T) {
	t.Run("should return hosts list", func(t *testing.T) {
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

	// t.Run("", func(t *testing.T) {
	// 	stats := []db.HostInfo{
	// 		{Hostname: "foo"},
	// 		{Hostname: "foo"},
	// 		{Hostname: "bar"},
	// 	}
	// 	hostDB := &MockHostDB{}
	// 	hostDB.SetHosts(stats)
	// 	hostsRouter := controller.NewHostsRouter(hostDB)

	// 	req, err := http.NewRequest("GET", "/entries", nil)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	rr := httptest.NewRecorder()
	// 	handler := http.HandlerFunc(hostsRouter.GetHosts)
	// 	handler.ServeHTTP(rr, req)
	// 	if status := rr.Code; status != http.StatusOK {
	// 		t.Errorf("handler returned wrong status code: got %v want %v",
	// 			status, http.StatusOK)
	// 	}
	// })
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
}

// GetHosts
func (m *MockHostDB) SetHosts(hosts []db.HostInfo) {
	m.hosts = hosts
}
func (m *MockHostDB) SetHostsError(err error) {
	m.hostsError = err
}
func (m *MockHostDB) GetHosts(pagination db.Pagination) ([]db.HostInfo, error) {
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
