package common

import "time"

type ProcessType int

const (
	ProcessTypeGo   = ProcessType(1)
	ProcessTypeJava = ProcessType(2)
)

type Process struct {
	Type         ProcessType
	PID          int
	BuildVersion string
	Path         string
	Agent        bool
	StartTime    time.Time
}
