package db_test

import (
	"testing"
	"time"

	"github.com/hamburghammer/gsave/db"
	"github.com/stretchr/testify/assert"
)

func TestGetHosts(t *testing.T) {
	t.Run("should return all 2 hosts", func(t *testing.T) {
		storage := make(map[string]db.Host)
		storage["foo"] = db.Host{HostInfo: db.HostInfo{Hostname: "foo"}}
		storage["bar"] = db.Host{HostInfo: db.HostInfo{Hostname: "bar"}}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		got, err := memDB.GetHosts(db.Pagination{0, 2})
		want := []db.HostInfo{{Hostname: "foo"}, {Hostname: "bar"}}

		assert.NoError(t, err)
		assert.EqualValues(t, want, got)
	})

	t.Run("should not find hosts on empty db", func(t *testing.T) {
		memDB := db.NewInMemoryDB()
		_, gotErr := memDB.GetHosts(db.Pagination{Skip: 0, Limit: 0})
		want := db.ErrHostsNotFound.Error()

		assert.EqualError(t, gotErr, want)
	})

	t.Run("should return error if all entries are beeing skiped", func(t *testing.T) {
		storage := make(map[string]db.Host)
		storage["foo"] = db.Host{HostInfo: db.HostInfo{Hostname: "foo"}}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		_, gotErr := memDB.GetHosts(db.Pagination{Skip: 2, Limit: 0})
		want := db.ErrAllEntriesSkipped.Error()

		assert.EqualError(t, gotErr, want)
	})

	t.Run("should return all entries left from skiping if limit is to high", func(t *testing.T) {
		storage := make(map[string]db.Host)
		storage["foo"] = db.Host{HostInfo: db.HostInfo{Hostname: "foo"}}
		storage["bar"] = db.Host{HostInfo: db.HostInfo{Hostname: "bar"}}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		got, err := memDB.GetHosts(db.Pagination{Skip: 0, Limit: 3})
		want := 2

		assert.NoError(t, err)
		assert.Equal(t, want, len(got))
	})
}

func TestGetHost(t *testing.T) {
	t.Run("should return all hosts", func(t *testing.T) {
		hostname := "foo"

		storage := make(map[string]db.Host)
		storage[hostname] = db.Host{HostInfo: db.HostInfo{Hostname: hostname}}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		got, err := memDB.GetHost(hostname)
		want := db.HostInfo{Hostname: hostname}

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("should not find host on empty db", func(t *testing.T) {
		memDB := db.NewInMemoryDB()
		_, gotErr := memDB.GetHost("")
		want := db.ErrHostNotFound.Error()

		assert.EqualError(t, gotErr, want)
	})
}

func TestGetStatsByHostname(t *testing.T) {
	t.Run("should return error if no host matching the name was found", func(t *testing.T) {
		storage := make(map[string]db.Host)

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		_, gotErr := memDB.GetStatsByHostname("foo", db.Pagination{Skip: 0, Limit: 0})
		want := db.ErrHostNotFound.Error()

		assert.EqualError(t, gotErr, want)
	})

	t.Run("should return error if all entries are beeing skiped", func(t *testing.T) {
		hostname := "foo"

		storage := make(map[string]db.Host)
		storage[hostname] = db.Host{HostInfo: db.HostInfo{Hostname: hostname}}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		_, gotErr := memDB.GetStatsByHostname(hostname, db.Pagination{Skip: 2, Limit: 0})
		want := db.ErrAllEntriesSkipped.Error()

		assert.EqualError(t, gotErr, want)
	})

	t.Run("should return all entries left from skiping if limit is to high", func(t *testing.T) {
		hostname := "foo"
		stats := []db.Stats{{Hostname: hostname}, {Hostname: hostname}}

		storage := make(map[string]db.Host)
		storage[hostname] = db.Host{HostInfo: db.HostInfo{Hostname: hostname}, Stats: stats}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		got, err := memDB.GetStatsByHostname(hostname, db.Pagination{Skip: 0, Limit: 3})
		want := 2

		assert.NoError(t, err)
		assert.Equal(t, want, len(got))
	})

	t.Run("should return all entries respecting skip and limit", func(t *testing.T) {
		hostname := "foo"
		stats := []db.Stats{{Hostname: hostname}, {Hostname: "target"}}

		storage := make(map[string]db.Host)
		storage[hostname] = db.Host{HostInfo: db.HostInfo{Hostname: hostname}, Stats: stats}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		got, err := memDB.GetStatsByHostname(hostname, db.Pagination{Skip: 1, Limit: 1})
		want := stats[1:]

		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestInsertStats_UpdateHostInfos(t *testing.T) {
	t.Run("should add new host in storage if not existing", func(t *testing.T) {
		hostname := "foo"
		stats := db.Stats{Hostname: hostname}

		memDB := db.NewInMemoryDB()

		_, err := memDB.GetHost(hostname)
		assert.Error(t, err, "It should not find the host")

		err = memDB.InsertStats(hostname, stats)
		assert.NoError(t, err)

		got, err := memDB.GetHost(hostname)
		want := 1

		assert.NoError(t, err)
		assert.Equal(t, want, got.DataPoints)
	})

	t.Run("should increment stats count and update time", func(t *testing.T) {
		hostname := "foo"
		stats := db.Stats{Hostname: hostname}

		storage := make(map[string]db.Host)
		storage[hostname] = db.Host{HostInfo: db.HostInfo{Hostname: hostname}, Stats: []db.Stats{stats}}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		oldHost, err := memDB.GetHost(hostname)
		assert.NoError(t, err)

		err = memDB.InsertStats(hostname, stats)
		assert.NoError(t, err)

		newHost, err := memDB.GetHost(hostname)

		assert.NoError(t, err)
		assert.Greater(t, newHost.DataPoints, oldHost.DataPoints)
		assert.Greater(t, time.Now().Nanosecond(), oldHost.LastInsert.Nanosecond())
	})
}

func TestInsertStats_AddStatsToHost(t *testing.T) {
	t.Run("should add the stats to a new host", func(t *testing.T) {
		hostname := "foo"
		stats := db.Stats{Hostname: hostname}

		memDB := db.NewInMemoryDB()

		_, err := memDB.GetHost(hostname)
		assert.Error(t, err, "It should not find the host")

		err = memDB.InsertStats(hostname, stats)
		assert.NoError(t, err)

		got, err := memDB.GetStatsByHostname(hostname, db.Pagination{Skip: 0, Limit: 1})

		assert.NoError(t, err)
		assert.Equal(t, stats, got[0])
	})

	t.Run("should increment stats count and update time", func(t *testing.T) {
		hostname := "foo"
		stats := db.Stats{Hostname: hostname}

		storage := make(map[string]db.Host)
		storage[hostname] = db.Host{HostInfo: db.HostInfo{Hostname: hostname}, Stats: []db.Stats{stats}}

		memDB := db.NewInMemoryDB().WithCustomStorage(storage)

		oldHost, err := memDB.GetStatsByHostname(hostname, db.Pagination{Skip: 0, Limit: 1})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(oldHost))

		err = memDB.InsertStats(hostname, stats)
		assert.NoError(t, err)

		newHost, err := memDB.GetStatsByHostname(hostname, db.Pagination{Skip: 0, Limit: 2})

		assert.NoError(t, err)
		assert.Equal(t, 2, len(newHost))

	})
}
