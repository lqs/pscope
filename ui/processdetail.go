package ui

import (
	"fmt"
	"strconv"

	"github.com/lqs/pscope/common"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessDetailView struct {
	*tview.Form
	noResource
}

type ProcessDetailViewParams struct {
	Process       common.Process
	OnClose       func()
	OnShowStack   func()
	OnCPUProfile  func()
	OnHeapProfile func()
}

func NewProcessDetailView(params ProcessDetailViewParams) ProcessDetailView {
	form := tview.NewForm()

	p, err := process.NewProcess(int32(params.Process.PID))
	if err != nil {
		panic(err)
	}
	name, _ := p.Name()
	cpuPercent, _ := p.CPUPercent()

	//form.SetHorizontal(true)
	form.SetItemPadding(0)
	form.SetBorder(true)
	form.SetTitle(" " + name + " ")
	form.SetBorderPadding(0, 0, 1, 1)

	form.AddTextView("PID", strconv.Itoa(params.Process.PID), 0, 1, true, false)
	form.AddTextView("Name", name, 0, 1, true, false)
	form.AddTextView("%CPU (average)", fmt.Sprintf("%.1f%%", cpuPercent*100), 0, 1, true, false)

	if params.Process.Agent {
		form.AddButton("Stack Dump", params.OnShowStack)
		if params.Process.Type == common.ProcessTypeGo {
			form.AddButton("CPU Profile", params.OnCPUProfile)
			form.AddButton("Heap Profile", params.OnHeapProfile)
		}
	} else {
		form.AddButton("gops agent not started. Press Esc to go back", nil)
	}
	form.SetButtonStyle(buttonStyle)
	form.SetButtonActivatedStyle(buttonActivatedStyle)
	form.SetCancelFunc(func() {
		params.OnClose()
	})

	return ProcessDetailView{
		Form: form,
	}
}
