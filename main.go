package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lqs/pscope/ui"
	"github.com/rivo/tview"
)

type Tui struct {
	application     *tview.Application
	navigationStack []ui.Widget
}

func (t *Tui) Push(w ui.Widget) {
	t.navigationStack = append(t.navigationStack, w)
	t.application.SetRoot(w, true)
}

func (t *Tui) Pop() {
	last := t.navigationStack[len(t.navigationStack)-1]
	last.Dispose()
	t.navigationStack = t.navigationStack[:len(t.navigationStack)-1]
	if len(t.navigationStack) == 0 {
		t.application.Stop()
		return
	}
	t.application.SetRoot(t.navigationStack[len(t.navigationStack)-1], true)
}

func main() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDarkBlue
	tview.Styles.PrimaryTextColor = tcell.ColorSilver
	tview.Styles.ContrastBackgroundColor = tcell.ColorGreen

	application := tview.NewApplication()

	t := &Tui{
		application: application,
	}

	processListView := ui.NewProcessListView(ui.ProcessListViewParams{
		Application: application,
		OnSelect: func(pid int) {
			t.Push(ui.NewProcessDetailView(ui.ProcessDetailViewParams{
				PID: pid,
				OnShowStack: func() {
					t.Push(ui.NewGoroutineStackView(ui.GoroutineStackViewParams{
						Application: application,
						PID:         pid,
						OnClose:     t.Pop,
					}))
				},
				OnCPUProfile: func() {
					t.Push(ui.NewProfileView(ui.ProfileViewParams{
						Application: application,
						PID:         pid,
						Type:        "cpu",
						OnClose:     t.Pop,
					}))
				},
				OnHeapProfile: func() {
					t.Push(ui.NewProfileView(ui.ProfileViewParams{
						Application: application,
						PID:         pid,
						Type:        "heap",
						OnClose:     t.Pop,
					}))
				},
				OnClose: t.Pop,
			}))
		},
		OnClose: t.Pop,
	})

	t.Push(ui.NewHomeView(ui.HomeViewParams{
		MainView: processListView,
	}))
	t.application.EnableMouse(true)

	if err := t.application.Run(); err != nil {
		panic(err)
	}
}
