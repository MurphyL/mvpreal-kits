package gui

import (
	"mvpreal/gui/window"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

const (
	desktopMinWidth  float32 = 1366
	desktopMinHeight float32 = 768
)

func RunApp() {
	appInst := app.New()
	featWin := NewAppWindow()
	mainWin := appInst.NewWindow(featWin.Title())
	mainWin.SetMaster()
	mainWin.SetIcon(featWin.AppIcon())
	c := container.New(featWin, featWin.Contents()...)
	mainWin.SetContent(c)
	mainWin.CenterOnScreen()
	mainWin.ShowAndRun()
}

type AppWindow interface {
	fyne.Layout
	AppIcon() fyne.Resource
	Title() string
	Contents() []fyne.CanvasObject
}

func NewAppWindow() AppWindow {
	return window.NewFeatureWindow(desktopMinWidth, desktopMinHeight)
}
