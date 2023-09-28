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
	selectGoroutineIndex func(index int)
	cancel               context.CancelFunc
}

type GoroutineStackViewParams struct {
	Application *tview.Application
	PID         int
	OnClose     func()
}

type GoroutineListView struct {
	*tview.Table
}

func (g GoroutineListView) Apply(goroutines []gops.Goroutine) {
	g.SetTitle(" Goroutines (" + strconv.Itoa(len(goroutines)) + ") ")
	for g.GetRowCount() > len(goroutines)+1 {
		g.RemoveRow(g.GetRowCount() - 1)
	}
	for i, goroutine := range goroutines {
		g.SetCell(i+1, 0, tview.NewTableCell(strconv.Itoa(goroutine.Id)))
		g.SetCell(i+1, 1, tview.NewTableCell(goroutine.State))
		g.SetCell(i+1, 2, tview.NewTableCell(goroutine.Wait))
	}
	if row, _ := g.GetSelection(); row <= 0 || row >= len(goroutines)+1 {
		g.Select(1, 0)
		g.ScrollToBeginning()
	}
}

func (v *GoroutineStackView) newGoroutineList() GoroutineListView {
	table := tview.NewTable()
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
		v.selectGoroutineIndex(row - 1)
	})
	table.SetSelectable(true, false)

	return GoroutineListView{
		Table: table,
	}
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
			v.selectGoroutineIndex = func(index int) {
				if flex.GetItemCount() > 1 {
					flex.RemoveItem(flex.GetItem(1))
				}
				if index < 0 || index >= len(goroutines) {
					return
				}
				goroutine := goroutines[index]
				stackListView = newStackList(goroutine.Frames)
				flex.AddItem(stackListView, 0, 3, false)
			}

			goroutineListView.Apply(goroutines)
		})
	})

	return v
}
