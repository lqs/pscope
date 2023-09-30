package common

import "time"

type ProcessType int

const (
	ProcessTypeGo ProcessType = iota
)

type Process struct {
	PID          int
	BuildVersion string
	Path         string
	Agent        bool
	StartTime    time.Time
}
