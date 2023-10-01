package ui

import (
	"context"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/gops/goprocess"
	"github.com/lqs/pscope/common"
	"github.com/lqs/pscope/jvmhsperf"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessListView struct {
	tview.Primitive
	cancel context.CancelFunc
}

type ProcessListViewParams struct {
	Application *tview.Application
	OnSelect    func(process common.Process)
	OnClose     func()
}

func findProcesses() []common.Process {
	var processes []common.Process
	for _, p := range goprocess.FindAll() {
		if p.PID == os.Getpid() {
			continue
		}
		proc, err := process.NewProcess(int32(p.PID))
		if err != nil {
			continue
		}
		createTime, _ := proc.CreateTime()
		processes = append(processes, common.Process{
			Type:         common.ProcessTypeGo,
			PID:          p.PID,
			BuildVersion: p.BuildVersion,
			Path:         p.Path,
			Agent:        p.Agent,
			StartTime:    time.UnixMilli(createTime),
		})
	}
	processes = append(processes, jvmhsperf.ListProcesses()...)
	slices.SortFunc(processes, func(a, b common.Process) int {
		if a.Agent != b.Agent {
			if a.Agent {
				return -1
			} else {
				return 1
			}
		}
		if r := b.StartTime.Compare(a.StartTime); r != 0 {
			return r
		}
		return b.PID - a.PID
	})
	return processes
}

func (v ProcessListView) Dispose() {
	v.cancel()
}

func NewProcessListView(params ProcessListViewParams) ProcessListView {
	table := tview.NewTable()

	columns := []string{"PID", "Start Time", "Version", "Command"}
	for c, column := range columns {
		table.SetCell(0, c, tview.NewTableCell(column).
			SetStyle(tableHeaderStyle).
			SetAlign(tview.AlignLeft))
	}

	table.SetBorderPadding(0, 0, 1, 1)
	table.SetFixed(1, len(columns))

	table.SetBorder(true).SetTitle(" Processes ")
	table.SetBorders(false)

	table.SetSelectionChangedFunc(func(row int, column int) {
		if row <= 0 {
			table.Select(1, 0)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	NewReloader(ctx, func() {
		processes := findProcesses()
		rowCount := table.GetRowCount()
		for i := 1; i < rowCount; i++ {
			table.RemoveRow(i)
		}
		for c := 0; c < len(columns); c++ {
			for i, process := range processes {
				cols := []string{
					strconv.Itoa(process.PID),
					process.StartTime.Format("01/02 15:04:05"),
					process.BuildVersion,
					process.Path,
				}
				alignment := tview.AlignLeft
				value := ""
				if c < len(cols) {
					value = cols[c]
				}

				expansion := 0
				if c == 3 {
					expansion = 1
				}

				textColor := tcell.ColorSilver
				if process.Agent {
					textColor = tcell.ColorWhite
				}

				params.Application.QueueUpdate(func() {
					table.SetCell(i+1, c, tview.NewTableCell(value).
						SetTextColor(textColor).
						SetAlign(alignment).
						SetExpansion(expansion),
					)
					table.SetSelectable(true, false)
					if row, _ := table.GetSelection(); row <= 0 {
						table.Select(1, 0)
					}
				})
			}
		}

		params.Application.QueueUpdateDraw(func() {
			table.ScrollToBeginning()
			table.SetSelectedFunc(func(row int, column int) {
				params.OnSelect(processes[row-1])
			})
		})
	})

	table.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			params.OnClose()
		}
	})
	return ProcessListView{
		Primitive: table,
		cancel:    cancel,
	}
}
