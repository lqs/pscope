package gops

import (
	"bufio"
	"bytes"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Goroutine struct {
	Id     int
	State  string
	Wait   string
	Frames []Frame
}

type Frame struct {
	Package string
	Func    string
	Params  string
	File    string
}

var headerRegex = regexp.MustCompile(`^goroutine (\d+) \[(.+?)(?:, (.+))?]:`)
var frameRegex = regexp.MustCompile(`^([a-zA-Z0-9._/\-]+\.)((?:\(\*?[^()]*\)\.)?\w+(?:\[\.\.\.])?)(?:\(([^()]*)\))?$`)

func ParseGoStack(stack []byte) (goroutines []Goroutine) {
	reader := bufio.NewReader(bytes.NewReader(stack))
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\n")
		matches := headerRegex.FindStringSubmatch(line)
		if matches == nil {
			break
		}

		goroutineId, _ := strconv.Atoi(matches[1])
		goroutine := Goroutine{
			Id:    goroutineId,
			State: matches[2],
			Wait:  matches[3],
		}
		for {
			frame := Frame{}
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			matches := frameRegex.FindStringSubmatch(line)
			if matches == nil {
				frame.Func = line
			} else {
				frame.Package = matches[1]
				frame.Func = matches[2]
				frame.Params = matches[3]
			}

			line, _ = reader.ReadString('\n')
			line = strings.TrimSpace(line)
			frame.File = line

			goroutine.Frames = append(goroutine.Frames, frame)
		}

		goroutines = append(goroutines, goroutine)
	}

	slices.SortFunc(goroutines, func(a, b Goroutine) int {
		return b.Id - a.Id
	})
	return
}
