package db

import "time"

// Stats struct represents a stats datapoint information of a host to be stored.
type Stats struct {
	Hostname string
	Date     time.Time
	CPU      float64
}
