package db

import "time"

// Stats struct represents a stats datapoint information of a host to be stored.
type Stats struct {
	Hostname string    `json:"hostname"`
	Date     time.Time `json:"date"`
	CPU      float64   `json:"cpu"`
}
