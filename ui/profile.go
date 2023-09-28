package ui

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/gops/signal"
	"github.com/google/pprof/profile"
	"github.com/lqs/pscope/gops"
	"github.com/rivo/tview"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ProfileView struct {
	*tview.Pages
	params ProfileViewParams
	flex   *tview.Flex
	cancel context.CancelFunc
}

type ProfileViewParams struct {
	Application *tview.Application
	PID         int
	Type        string
	OnClose     func()
}

func (v ProfileView) Dispose() {
	v.cancel()
}

func (v ProfileView) start(ctx context.Context) {

	progressDialog := tview.NewModal()
	progressDialog.SetBackgroundColor(tcell.ColorSilver)
	progressDialog.SetTextColor(tcell.ColorBlack)
	//progressDialog.AddButtons([]string{"Cancel"})
	//progressDialog.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
	//	params.OnClose()
	//})
	v.Pages.AddPage("progressDialog", progressDialog, true, true)

	done := make(chan struct{})

	var p byte
	if v.params.Type == "cpu" {
		p = signal.CPUProfile
		go func() {
			const seconds = 30
			const width = 50
			bar := make([]rune, width)
			for i := 0; i < width; i++ {
				bar[i] = tview.BlockLightShade
			}

			for i := 0; i < width; i++ {
				v.params.Application.QueueUpdateDraw(func() {
					progressDialog.SetText("Profiling CPU now, will take 30 secs... \n\n" + string(bar))
				})
				select {
				case <-done:
					return
				case <-time.After(seconds * time.Second / width):
					bar[i] = tview.BlockDarkShade
				}
			}
		}()
	} else if v.params.Type == "heap" {
		p = signal.HeapProfile
	} else {
		panic("unknown profile type")
	}

	go func() {
		defer close(done)
		result, err := gops.Cmd(ctx, v.params.PID, p)
		if err != nil {
			// TODO: show error
			return
		}

		profile, err := profile.ParseData(result)
		if err != nil {
			// TODO: show error
			return
		}
		v.params.Application.QueueUpdateDraw(func() {
			sampleTypeTable := tview.NewTable()
			sampleTypeTable.SetBorder(true)
			sampleTypeTable.SetTitle(" Sample Types ")
			sampleTypeTable.SetBorderPadding(0, 0, 1, 1)
			for i, sampleType := range profile.SampleType {
				sampleTypeTable.SetCellSimple(i, 0, sampleType.Type)
			}
			sampleTypeTable.SetSelectable(true, false)

			sampleTypeTable.SetBorders(false)

			v.flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEsc:
					v.params.OnClose()
				}
				return event
			})

			v.flex.AddItem(sampleTypeTable, 0, 1, true)

			dataTable := tview.NewTable()
			dataTable.SetBorder(true)
			dataTable.SetBorderPadding(0, 0, 1, 1)
			dataTable.SetFixed(1, 0)
			dataTable.SetSelectable(true, false)
			dataTable.SetSelectionChangedFunc(func(row int, column int) {
				if row <= 0 {
					dataTable.Select(1, 0)
				}
			})

			v.flex.AddItem(dataTable, 0, 3, false)

			sampleTypeTable.SetSelectionChangedFunc(func(row int, column int) {
				if row < 0 {
					sampleTypeTable.Select(0, 0)
					return
				}
				dataTable.SetTitle(" Top " + profile.SampleType[row].Type + " ")
				dataTable.SetCell(0,
					0,
					tview.NewTableCell(cases.Title(language.English).String(profile.SampleType[row].Unit)).
						SetStyle(tableHeaderStyle).
						SetAlign(tview.AlignRight),
				)
				dataTable.SetCell(0, 1, tview.NewTableCell("Function").SetStyle(tableHeaderStyle))
				sortSamples(profile.Sample, row)
				for i, sample := range profile.Sample {
					var v string
					if profile.SampleType[row].Unit == "bytes" {
						v = friendlySize(sample.Value[row])
					} else {
						v = fmt.Sprint(sample.Value[row])
					}
					dataTable.SetCell(1+i, 0, tview.NewTableCell(v).SetAlign(tview.AlignRight))
					dataTable.SetCell(1+i, 1, tview.NewTableCell(sample.Location[0].Line[0].Function.Name))
				}
				dataTable.Select(1, 0)
				dataTable.ScrollToBeginning()
			})

			sampleTypeTable.Select(0, 0)

			v.Pages.RemovePage("progressDialog")
			_ = result
		})

	}()
}

func NewProfileView(params ProfileViewParams) ProfileView {
	ctx, cancel := context.WithCancel(context.Background())

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.SetTitle(" Profile ")

	pages := tview.NewPages()
	pages.AddPage("table", flex, true, true)

	view := ProfileView{
		Pages:  pages,
		params: params,
		flex:   flex,
		cancel: cancel,
	}

	if params.Type == "cpu" {
		confirmDialog := tview.NewModal()
		confirmDialog.SetBackgroundColor(tcell.ColorSilver)
		confirmDialog.SetTextColor(tcell.ColorBlack)
		confirmDialog.AddButtons([]string{
			"  OK  ",
			"Cancel",
		})
		confirmDialog.SetButtonStyle(buttonStyle)
		confirmDialog.SetButtonActivatedStyle(buttonActivatedStyle)
		confirmDialog.SetText("Profiling CPU will take 30 secs, continue?")
		confirmDialog.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0:
				pages.RemovePage("confirmDialog")
				view.start(ctx)
			case 1:
				params.OnClose()
			}
		})
		pages.AddPage("confirmDialog", confirmDialog, true, true)
	} else {
		view.start(ctx)
	}

	return view
}

func friendlySize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1fK", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1fM", float64(size)/(1024*1024))
	} else if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.1fG", float64(size)/(1024*1024*1024))
	} else {
		return fmt.Sprintf("%.1fT", float64(size)/(1024*1024*1024*1024))
	}
}

func sortSamples(samples []*profile.Sample, valueIndex int) {
	slices.SortFunc(samples, func(a, b *profile.Sample) int {
		return -cmp.Compare(a.Value[valueIndex], b.Value[valueIndex])
	})
}
