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
	// Resize will be set later based on isFirstTime
	win.CenterOnScreen()

	dbPath := models.DefaultDBPath
	isFirstTime := false
	if _, e := os.Stat(dbPath); os.IsNotExist(e) {
		isFirstTime = true
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите мастер-пароль")

	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.SetPlaceHolder("Повторите мастер-пароль")
	confirmEntry.Hidden = !isFirstTime

	warningLabel := widget.NewLabel("⚠️ Внимание! Мастер-пароль нельзя изменить после создания базы данных.\nЛучше запишите его на бумажку и храните в безопасном месте.")
	warningLabel.Wrapping = fyne.TextWrapWord
	warningLabel.Hidden = !isFirstTime

	status := widget.NewLabel("")

	var dbase *sql.DB
	var key []byte
	var err error

	var loginBtn *widget.Button = widget.NewButton("Войти", func() {
		master := passwordEntry.Text
		if master == "" {
			status.SetText("⚠️ Пароль не может быть пустым!")
			return
		}

		if isFirstTime {
			confirm := confirmEntry.Text
			if confirm != master {
				status.SetText("Пароли не совпадают!")
				return
			}
			dbase, key, err = db.CreateNewDatabase(dbPath, master)
			if err != nil {
				status.SetText("Ошибка создания базы: " + err.Error())
				return
			}
			entries, _ := db.LoadAllEntries(dbase, key)
			ShowMainWindow(a, dbase, key, entries)
			win.Close()
			return
		}

		dbase, key, err = db.OpenAndAuthenticate(dbPath, master)
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
		confirmEntry,
		warningLabel,
		status,
		layout.NewSpacer(),
		loginBtn,
	)

	win.SetContent(container.NewPadded(content))

	if isFirstTime {
		win.Resize(fyne.NewSize(400, 0))
	} else {
		win.Resize(fyne.NewSize(350, 0))
	}

	win.ShowAndRun()
}
