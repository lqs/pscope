package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type HomeView struct {
	*tview.Flex
	mainView Widget
}

type HomeViewParams struct {
	MainView Widget
}

func (v HomeView) Dispose() {
	v.mainView.Dispose()
}

func NewHomeView(params HomeViewParams) HomeView {
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)

	titleView := tview.NewTextView()
	titleView.SetText(" pscope ")
	titleView.SetTextColor(tcell.ColorWhite)
	titleView.SetBackgroundColor(tcell.ColorGrey)

	flex.AddItem(titleView, 1, 1, false)
	flex.AddItem(params.MainView, 0, 1, true)

	footerView := tview.NewTextView()
	footerView.SetDynamicColors(true)
	footerView.SetText(" Press [white]↑↓[silver] to navigate, [white]Tab[silver] to switch focus, [white]Enter[silver] to select, [white]Esc[sliver] to go back. ")
	footerView.SetTextColor(tcell.ColorSilver)
	footerView.SetBackgroundColor(tcell.ColorGrey)
	flex.AddItem(footerView, 1, 1, false)

	return HomeView{
		Flex:     flex,
		mainView: params.MainView,
	}
}
