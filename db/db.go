package db

import "errors"

var (
	// ErrHostNotFound error if the host could not be found.
	ErrHostNotFound = errors.New("Host not found")
	// ErrHostsNotFound if no hosts where found.
	ErrHostsNotFound = errors.New("No hosts found")
	// ErrAllEntriesSkipped if all entries are beeing skiped.
	ErrAllEntriesSkipped = errors.New("All entries skipped")
)

// HostDB is an interface to acquire information of the hosts saved inside the DB and to update them.
type HostDB interface {
	// GetHosts returns all hosts respecting to the pagination.
	GetHosts(skip, limit int) ([]HostInfo, error)
	// GetHost by the hostname.
	GetHost(hostname string) (HostInfo, error)
	// GetStatsByHostname get all stats entries for a hostname respecting the pagination.
	GetStatsByHostname(hostname string, skip, limit int) ([]Stats, error)

	// InsertStats insert a new stats dataset into the db.
	InsertStats(stats Stats) error
}
