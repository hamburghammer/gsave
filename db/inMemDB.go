package db

import (
	"sync"
	"time"
)

// NewInMemoryDB a constructor to build a new inMemoryDB.
func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{storage: make(map[string]Host), m: sync.Mutex{}}
}

// InMemoryDB a in memory DB implementing the db.HostDB interface.
type InMemoryDB struct {
	storage map[string]Host
	m       sync.Mutex
}

// WithCustomStorage allows to put a custom map as DB storage.
// Returns a new InMemoryDB.
func (db *InMemoryDB) WithCustomStorage(storage map[string]Host) *InMemoryDB {
	db.m.Lock()
	defer db.m.Unlock()

	return &InMemoryDB{storage: storage, m: sync.Mutex{}}
}

// GetHosts returns a paginated result of all hosts.
// It returns an error if no host was found or all entries are beeing skiped.
func (db *InMemoryDB) GetHosts(pagination Pagination) ([]HostInfo, error) {
	db.m.Lock()
	defer db.m.Unlock()

	hosts := make([]HostInfo, 0)
	for _, value := range db.storage {
		hosts = append(hosts, value.HostInfo)
	}

	if len(hosts) == 0 {
		return []HostInfo{}, ErrHostsNotFound
	}

	records := len(hosts)
	if records < pagination.Skip {
		return []HostInfo{}, ErrAllEntriesSkipped
	} else if records < (pagination.Skip + pagination.Limit) {
		hosts = hosts[pagination.Skip:]
		foundHosts := make([]HostInfo, len(hosts))
		copy(foundHosts, hosts)
		return foundHosts, nil
	}

	hosts = hosts[pagination.Skip:(pagination.Skip + pagination.Limit)]
	foundHosts := make([]HostInfo, len(hosts))
	copy(foundHosts, hosts)
	return foundHosts, nil
}

// GetHost returns a host with the matching hostname.
// If no host could be found it will return an error.
func (db *InMemoryDB) GetHost(hostname string) (HostInfo, error) {
	db.m.Lock()
	defer db.m.Unlock()

	host, found := db.storage[hostname]
	if !found {
		return HostInfo{}, ErrHostNotFound
	}

	return host.HostInfo, nil
}

// GetStatsByHostname gets all Stats in a paginated form from a specific host.
// It returns errors if no host is found or if all entries are beeing skiped.
func (db *InMemoryDB) GetStatsByHostname(hostname string, pagination Pagination) ([]Stats, error) {
	db.m.Lock()
	defer db.m.Unlock()

	host, found := db.storage[hostname]
	if !found {
		return []Stats{}, ErrHostNotFound
	}

	records := len(host.Stats)
	if records < pagination.Skip {
		return []Stats{}, ErrAllEntriesSkipped
	} else if records < (pagination.Skip + pagination.Limit) {
		stats := host.Stats[pagination.Skip:]
		foundStats := make([]Stats, len(stats))
		copy(foundStats, stats)
		return foundStats, nil
	}

	stats := host.Stats[pagination.Skip:(pagination.Skip + pagination.Limit)]
	foundStats := make([]Stats, len(stats))
	copy(foundStats, stats)
	return foundStats, nil
}

// InsertStats into the DB.
// To do so it takes the hostname of the Hostname field and creates a new host inside the DB and/or adds the stat to it.
// The HostInfos are also beeing updated.
// This implementation won't return an error but its declared to implement the db.HostDB interface.
func (db *InMemoryDB) InsertStats(hostname string, stats Stats) error {
	db.m.Lock()
	defer db.m.Unlock()

	host, found := db.storage[hostname]
	if !found {
		hostInfo := HostInfo{Hostname: hostname, DataPoints: 1, LastInsert: time.Now()}
		db.storage[hostname] = Host{HostInfo: hostInfo, Stats: []Stats{stats}}
		return nil
	}

	host.Stats = db.insertAtBeginning(host.Stats, stats)
	host.HostInfo.DataPoints++
	host.HostInfo.LastInsert = time.Now()

	db.storage[hostname] = host
	return nil
}

func (db *InMemoryDB) insertAtBeginning(stats []Stats, stat Stats) []Stats {
	stats = append(stats, Stats{})
	copy(stats[1:], stats)
	stats[0] = stat

	return stats
}
