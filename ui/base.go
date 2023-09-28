package ui

import "github.com/rivo/tview"

type Widget interface {
	tview.Primitive
	Dispose()
}

type noResource struct{}

func (n noResource) Dispose() {}
