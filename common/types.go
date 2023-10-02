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

type CallStack struct {
	Id                int
	Name              string
	State             string
	Wait              string
	Frames            []*Frame
	ParentGoroutineId int
	ParentFrame       *Frame
	Children          []*CallStack
}

type Frame struct {
	Package string
	Func    string
	Params  string
	File    string
}
