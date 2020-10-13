package db

import "time"

// Host representation of a Host inside the DB.
type Host struct {
	HostInfo HostInfo
	Stats    []Stats
}

// HostInfo a small object to represent the Host object.
type HostInfo struct {
	Hostname    string
	StatsAmount int
	LastInsert  time.Time
}
