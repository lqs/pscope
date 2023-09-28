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
	Id                int
	State             string
	Wait              string
	Frames            []Frame
	ParentGoroutineId int
	ParentFrame       Frame
	Children          []*Goroutine
}

type Frame struct {
	Package string
	Func    string
	Params  string
	File    string
}

var headerRegex = regexp.MustCompile(`^goroutine (\d+) \[(.+?)(?:, (.+))?]:`)
var frameRegex = regexp.MustCompile(`^([a-zA-Z0-9._/\-]+\.)((?:\(\*?[^()]*\)\.)?\w+(?:\[\.\.\.])?)(?:\(([^()]*)\))?$`)
var createdByRegex = regexp.MustCompile(`^created by .+ in goroutine (\d+)$`)

func ParseGoStack(stack []byte) []*Goroutine {
	var goroutines []*Goroutine
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
		goroutine := &Goroutine{
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

			matches := createdByRegex.FindStringSubmatch(line)
			if matches != nil {
				goroutine.ParentGoroutineId, _ = strconv.Atoi(matches[1])
			} else {
				matches := frameRegex.FindStringSubmatch(line)
				if matches != nil {
					frame.Package = matches[1]
					frame.Func = matches[2]
					frame.Params = matches[3]
				} else {
					frame.Func = line
				}
			}

			line, _ = reader.ReadString('\n')
			line = strings.TrimSpace(line)
			frame.File = line

			goroutine.Frames = append(goroutine.Frames, frame)
		}

		goroutines = append(goroutines, goroutine)
	}

	roots := makeTree(goroutines)
	sortGoroutines(roots)
	return roots
}

func makeTree(goroutines []*Goroutine) []*Goroutine {
	goroutineMap := make(map[int]*Goroutine)
	for i := range goroutines {
		goroutineMap[goroutines[i].Id] = goroutines[i]
	}

	for _, goroutine := range goroutines {
		if goroutine.ParentGoroutineId != 0 {
			parent, ok := goroutineMap[goroutine.ParentGoroutineId]
			if !ok {
				// make a dummy parent goroutine
				parent = &Goroutine{
					Id:                goroutine.ParentGoroutineId,
					State:             "terminated",
					ParentGoroutineId: 1,
				}
				goroutineMap[goroutine.ParentGoroutineId] = parent
				goroutines = append(goroutines, parent)
			}
			parent.Children = append(parent.Children, goroutine)
		}
	}

	var roots []*Goroutine
	for i := range goroutines {
		goroutine := goroutines[i]
		if goroutine.ParentGoroutineId == 0 {
			roots = append(roots, goroutine)
		}
	}
	return roots
}

func sortGoroutines(goroutines []*Goroutine) {
	slices.SortFunc(goroutines, func(a, b *Goroutine) int {
		return b.Id - a.Id
	})
	for _, goroutine := range goroutines {
		sortGoroutines(goroutine.Children)
	}
}
