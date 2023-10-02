package ui

import (
	"context"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/google/gops/signal"
	"github.com/lqs/pscope/common"
	"github.com/lqs/pscope/gops"
	"github.com/lqs/pscope/jvmhsperf"
	"github.com/rivo/tview"
)

type GoroutineStackView struct {
	*tview.Flex
	selectGoroutine func(goroutine *common.CallStack)
	cancel          context.CancelFunc
}

type GoroutineStackViewParams struct {
	Application *tview.Application
	Process     common.Process
	OnClose     func()
}

type GoroutineListView struct {
	*tview.Table
	process          common.Process
	indexToGoroutine []*common.CallStack
	currentRow       int
}

func (g *GoroutineListView) add(goroutine *common.CallStack, level int, isLastChild bool) {
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

	switch g.process.Type {
	case common.ProcessTypeGo:
		g.SetCell(g.currentRow, 0, tview.NewTableCell(prefix+strconv.Itoa(goroutine.Id)))
		g.SetCell(g.currentRow, 1, tview.NewTableCell(goroutine.State))
		g.SetCell(g.currentRow, 2, tview.NewTableCell(goroutine.Frames[0].Func))
		g.SetCell(g.currentRow, 3, tview.NewTableCell(goroutine.Wait))
	case common.ProcessTypeJava:
		g.SetCell(g.currentRow, 0, tview.NewTableCell(prefix+goroutine.Name))
		g.SetCell(g.currentRow, 1, tview.NewTableCell(goroutine.State))
	}

	g.indexToGoroutine = append(g.indexToGoroutine, goroutine)
	g.currentRow++

	for i, child := range goroutine.Children {
		g.add(child, level+1, i+1 == len(goroutine.Children))
	}
}

func (g *GoroutineListView) Apply(callStacks []*common.CallStack) {
	g.indexToGoroutine = nil
	g.currentRow = 1
	for i, goroutine := range callStacks {
		g.add(goroutine, 0, i+1 == len(callStacks))
	}
	if g.process.Type == common.ProcessTypeGo {
		g.SetTitle(" Goroutines (" + strconv.Itoa(len(g.indexToGoroutine)) + ") ")
	} else {
		g.SetTitle(" Threads (" + strconv.Itoa(len(g.indexToGoroutine)-1) + ") ")
	}
	for g.GetRowCount() > g.currentRow+1 {
		g.RemoveRow(g.GetRowCount() - 1)
	}
	if row, _ := g.GetSelection(); row <= 0 || row >= g.currentRow+1 {
		g.Select(1, 0)
		g.ScrollToBeginning()
	}
}

func (v *GoroutineStackView) newGoroutineList(process common.Process) *GoroutineListView {
	table := tview.NewTable()
	g := &GoroutineListView{
		Table:   table,
		process: process,
	}

	table.SetBorder(true)
	table.SetBorderPadding(0, 0, 1, 1)
	table.SetFixed(1, 3)
	switch process.Type {
	case common.ProcessTypeGo:
		table.SetCell(0, 0, tview.NewTableCell("ID"))
		table.SetCell(0, 1, tview.NewTableCell("State"))
		table.SetCell(0, 2, tview.NewTableCell("Function"))
		table.SetCell(0, 3, tview.NewTableCell("Wait"))
	case common.ProcessTypeJava:
		table.SetCell(0, 0, tview.NewTableCell("Name"))
		table.SetCell(0, 1, tview.NewTableCell("State"))
	}
	table.SetSelectionChangedFunc(func(row, column int) {
		if row <= 0 {
			row = 1
			table.Select(row, column)
		}
		if row < len(g.indexToGoroutine) {
			v.selectGoroutine(g.indexToGoroutine[row-1])
		}
	})
	table.SetSelectable(true, false)

	return g
}

func newStackList(frames []*common.Frame) tview.Primitive {
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

	goroutineListView := v.newGoroutineList(params.Process)
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
		var goroutines []*common.CallStack
		switch params.Process.Type {
		case common.ProcessTypeGo:
			result, _ := gops.Cmd(ctx, params.Process.PID, signal.StackTrace)
			goroutines = gops.ParseGoStack(result)
		case common.ProcessTypeJava:
			result, _ := jvmhsperf.Execute(params.Process.PID, "threaddump")
			for _, thread := range jvmhsperf.ParseJavaThreadDump(result) {
				goroutines = append(goroutines, &common.CallStack{
					Id:     thread.Id,
					Name:   thread.Name,
					State:  thread.State,
					Frames: thread.Frames,
				})
			}
		}

		params.Application.QueueUpdateDraw(func() {
			v.selectGoroutine = func(goroutine *common.CallStack) {
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
