package db

import "time"

// Stats struct represents a stats datapoint information of a host to be stored.
type Stats struct {
	Hostname  string    `json:"hostname"`
	Date      time.Time `json:"date"`
	CPU       float64   `json:"cpu"`
	Processes []Process `json:"processes"`
	Disk      Memory
	Mem       Memory
}

// Process is the representation of a UNIX process with some of its information.
type Process struct {
	Name string  `json:"name"`
	Pid  int     `json:"pid"`
	CPU  float64 `json:"cpu"`
}

// Memory represents the usage of disk or RAM space.
// It shows the used and the total available space.
type Memory struct {
	Used  int `json:"used"`
	Total int `json:"total"`
}
