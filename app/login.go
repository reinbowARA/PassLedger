package app

import (
	"database/sql"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/reinbowARA/PassLedger/db"
	"github.com/reinbowARA/PassLedger/models"
)

func ShowLoginWindow(a fyne.App) {
	win := a.NewWindow("Password Book — Вход")
	win.Resize(fyne.NewSize(350, 250))
	win.CenterOnScreen()

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите мастер-пароль")

	status := widget.NewLabel("")

	loginBtn := widget.NewButton("Войти", func() {
		master := passwordEntry.Text
		if master == "" {
			status.SetText("⚠️ Пароль не может быть пустым!")
			return
		}

		dbPath := models.DefaultDBPath
		var dbase *sql.DB
		var key []byte
		var err error

		if _, e := os.Stat(dbPath); os.IsNotExist(e) {
			dbase, key, err = db.CreateNewDatabase(dbPath, master)
		} else {
			dbase, key, err = db.OpenAndAuthenticate(dbPath, master)
		}

		if err != nil {
			status.SetText("Ошибка: " + err.Error())
			return
		}

		entries, _ := db.LoadAllEntries(dbase, key)
		ShowMainWindow(a, dbase, key, entries)
		win.Close()
	})

	content := container.NewVBox(
		widget.NewLabel("Введите мастер-пароль"),
		passwordEntry,
		status,
		layout.NewSpacer(),
		loginBtn,
	)

	win.SetContent(container.NewPadded(content))
	win.ShowAndRun()
}
