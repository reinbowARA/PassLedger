package app

import (
	"database/sql"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/reinbowARA/PassLedger/db"
)

func ShowLoginWindow(a fyne.App) {
	win := a.NewWindow("Password Book — Вход")
	// Resize will be set later based on isFirstTime
	win.CenterOnScreen()

	settings, _ := LoadSettings()

	var dbPath string = settings.DBPath
	isFirstTime := false
	if stat, e := os.Stat(dbPath); os.IsNotExist(e) || (e == nil && stat.Size() == 0) {
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

	// Database path selection elements (only for first time)
	dbFolderLabel := widget.NewLabel("Выберите папку базы данных:")
	dbFolderLabel.Hidden = !isFirstTime



	dbFolderEntry := widget.NewEntry()
	dbFolderEntry.SetPlaceHolder("Папка базы данных")
	dbFolderEntry.SetText("data")
	dbFolderEntry.Hidden = !isFirstTime

	dbFolderEntryContainer := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(250, 36)),
		dbFolderEntry,
	)

	dbFileEntry := widget.NewEntry()
	dbFileEntry.SetPlaceHolder("Имя файла")
	dbFileEntry.SetText("passwords.db")
	dbFileEntry.Hidden = !isFirstTime

	dbFileEntryContainer := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(250, 36)),
		dbFileEntry,
	)

	browseFolderBtn := widget.NewButton("Обзор...", func() {
		dlg := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			dbFolderEntry.SetText(list.Path())
		}, win)
		dlg.Show()
	})
	browseFolderBtn.Hidden = !isFirstTime

	dbFileLabel := widget.NewLabel("Имя файла базы данных:")
	dbFileLabel.Hidden = !isFirstTime



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
			dbfolder := dbFolderEntry.Text
			if dbfolder == "" {
				dbfolder = "data"
			}
			dbfile := dbFileEntry.Text
			if dbfile == "" {
				dbfile = "passwords.db"
			}
			dbPath = filepath.Join(dbfolder, dbfile)
			dbase, key, err = db.CreateNewDatabase(dbPath, master)
			if err != nil {
				status.SetText("Ошибка создания базы: " + err.Error())
				return
			}
			// Save new dbPath to settings
			settings.DBPath = dbPath
			SaveSettings(settings)
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
		dbFolderLabel,
		container.NewHBox(browseFolderBtn, dbFolderEntryContainer),
		dbFileLabel,
		dbFileEntryContainer,
		status,
		layout.NewSpacer(),
		loginBtn,
	)

	win.SetContent(container.NewPadded(content))

	if isFirstTime {
		win.Resize(fyne.NewSize(450, 300))
	} else {
		win.Resize(fyne.NewSize(350, 0))
	}

	win.ShowAndRun()
}
