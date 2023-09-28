package ui

import (
	"context"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/google/gops/signal"
	"github.com/lqs/pscope/gops"
	"github.com/rivo/tview"
)

type GoroutineStackView struct {
	*tview.Flex
	selectGoroutine func(goroutine *gops.Goroutine)
	cancel          context.CancelFunc
}

type GoroutineStackViewParams struct {
	Application *tview.Application
	PID         int
	OnClose     func()
}

type GoroutineListView struct {
	*tview.Table
	indexToGoroutine map[int]*gops.Goroutine
	currentRow       int
}

func (g *GoroutineListView) add(goroutine *gops.Goroutine, level int, isLastChild bool) {
	prefix := ""
	if level > 0 {
		for i := 0; i < level-1; i++ {
			prefix += "│ "
		}
		if isLastChild {
			prefix += "└ "
		} else {
			prefix += "├ "
		}
	}
	g.SetCell(g.currentRow, 0, tview.NewTableCell(prefix+strconv.Itoa(goroutine.Id)))
	g.SetCell(g.currentRow, 1, tview.NewTableCell(goroutine.State))
	g.SetCell(g.currentRow, 2, tview.NewTableCell(goroutine.Frames[0].Func))
	g.SetCell(g.currentRow, 3, tview.NewTableCell(goroutine.Wait))
	g.indexToGoroutine[g.currentRow] = goroutine
	g.currentRow++

	for i, child := range goroutine.Children {
		g.add(child, level+1, i+1 == len(goroutine.Children))
	}
}

func (g *GoroutineListView) Apply(goroutines []*gops.Goroutine) {
	//g.SetTitle(" Goroutines (" + strconv.Itoa(len(goroutines)) + ") ")
	//for g.GetRowCount() > len(goroutines)+1 {
	//	g.RemoveRow(g.GetRowCount() - 1)
	//}
	g.indexToGoroutine = make(map[int]*gops.Goroutine)
	g.currentRow = 1
	for i, goroutine := range goroutines {
		g.add(goroutine, 0, i+1 == len(goroutines))
	}
	for g.GetRowCount() > g.currentRow+1 {
		g.RemoveRow(g.GetRowCount() - 1)
	}
	if row, _ := g.GetSelection(); row <= 0 || row >= g.currentRow+1 {
		g.Select(1, 0)
		g.ScrollToBeginning()
	}
}

func (v *GoroutineStackView) newGoroutineList() *GoroutineListView {
	table := tview.NewTable()
	g := &GoroutineListView{
		Table: table,
	}

	table.SetBorder(true)
	table.SetTitle(" Goroutines ")
	table.SetBorderPadding(0, 0, 1, 1)
	table.SetFixed(1, 3)
	table.SetCell(0, 0, tview.NewTableCell("ID"))
	table.SetCell(0, 1, tview.NewTableCell("State"))
	table.SetCell(0, 2, tview.NewTableCell("Wait"))
	table.SetSelectionChangedFunc(func(row, column int) {
		if row <= 0 {
			row = 1
			table.Select(row, column)
		}
		v.selectGoroutine(g.indexToGoroutine[row])
	})
	table.SetSelectable(true, false)

	return g
}

func newStackList(frames []gops.Frame) tview.Primitive {
	table := tview.NewTable()
	table.SetBorder(true).SetTitle(" Stack Frames ")
	table.SetBorderPadding(0, 0, 1, 1)
	table.SetFixed(1, 3)
	table.SetCell(0, 0, tview.NewTableCell("Function").SetTextColor(tview.Styles.SecondaryTextColor))
	table.SetCell(0, 1, tview.NewTableCell("File").SetTextColor(tview.Styles.SecondaryTextColor))

	for i, frame := range frames {
		table.SetCell(i+1, 0, tview.NewTableCell(frame.Func))
		table.SetCell(i+1, 1, tview.NewTableCell(frame.File))
	}
	table.Select(1, 0)
	table.SetSelectionChangedFunc(func(row, column int) {
		if row == 0 {
			table.Select(1, column)
		}
	})
	table.SetSelectable(true, false)

	return table
}

func (v *GoroutineStackView) Dispose() {
	v.cancel()
}

func NewGoroutineStackView(params GoroutineStackViewParams) Widget {
	ctx, cancel := context.WithCancel(context.Background())

	flex := tview.NewFlex()

	v := &GoroutineStackView{
		Flex:   flex,
		cancel: cancel,
	}

	goroutineListView := v.newGoroutineList()
	var stackListView tview.Primitive
	flex.AddItem(goroutineListView, 0, 1, true)
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			params.OnClose()
		case tcell.KeyTab:
			if stackListView == nil {
				goroutineListView.Focus(nil)
			} else if !goroutineListView.HasFocus() {
				goroutineListView.Focus(nil)
				stackListView.Blur()
			} else {
				stackListView.Focus(nil)
				goroutineListView.Blur()
			}
		}
		return event
	})

	NewReloader(ctx, func() {
		result, _ := gops.Cmd(ctx, params.PID, signal.StackTrace)
		goroutines := gops.ParseGoStack(result)

		params.Application.QueueUpdateDraw(func() {
			v.selectGoroutine = func(goroutine *gops.Goroutine) {
				if flex.GetItemCount() > 1 {
					flex.RemoveItem(flex.GetItem(1))
				}
				if goroutine == nil {
					return
				}
				stackListView = newStackList(goroutine.Frames)
				flex.AddItem(stackListView, 0, 3, false)
			}

			goroutineListView.Apply(goroutines)
		})
	})

	return v
}
