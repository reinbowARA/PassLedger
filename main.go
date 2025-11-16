package main

import (
	application "github.com/reinbowARA/PassLedger/app"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	var icon fyne.Resource
	var err error
	a := app.NewWithID("PassLedger")
	theme := a.Settings().ThemeVariant()
	if theme == 0 {
		icon, err = fyne.LoadResourceFromPath("src/icon-white.svg")
	} else {
		icon, err = fyne.LoadResourceFromPath("src/icon.svg")
	}
	if err == nil {
		a.SetIcon(icon)
	}
	application.ShowLoginWindow(a)
}
