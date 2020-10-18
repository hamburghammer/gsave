package db

import "errors"

var (
	// ErrHostNotFound error if the host could not be found.
	ErrHostNotFound = errors.New("db: Host not found")
	// ErrHostsNotFound if no hosts where found.
	ErrHostsNotFound = errors.New("db: No hosts found")
	// ErrAllEntriesSkipped if all entries are beeing skiped.
	ErrAllEntriesSkipped = errors.New("db: All entries skipped")
)

// HostDB is an interface to acquire information of the hosts saved inside the DB and to update them.
type HostDB interface {
	// GetHosts returns all hosts respecting to the pagination.
	// Returns an ErrHostsNotFound if no hosts could be found or ErrAllEntriesSkipped if the skip values is to high.
	GetHosts(pagination Pagination) ([]HostInfo, error)

	// GetHost by the hostname.
	// Returns ErrHostNotFound if no host with the host name could be found.
	GetHost(hostname string) (HostInfo, error)

	// GetStatsByHostname get all stats entries for a hostname respecting the pagination.
	// Returns ErrHostNotFound if no host with the host name could be found or ErrAllEntriesSkipped if the skip values is to high.
	GetStatsByHostname(hostname string, pagination Pagination) ([]Stats, error)

	// InsertStats insert a new stats dataset into the db.
	InsertStats(hostname string, stats Stats) error
}

type Pagination struct {
	Skip  int
	Limit int
}
