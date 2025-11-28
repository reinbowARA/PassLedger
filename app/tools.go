package app

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/reinbowARA/PassLedger/crypto"
	"github.com/reinbowARA/PassLedger/db"
	"github.com/reinbowARA/PassLedger/models"
)

func showPasswordGeneratorPopup(win fyne.Window) {
	lengthEntry := widget.NewEntry()
	lengthEntry.SetText("16")
	lengthEntry.SetPlaceHolder("Длина пароля")

	passwordEntry := widget.NewEntry()
	passwordEntry.SetPlaceHolder("Сгенерированный пароль")
	passwordEntry.Disable()

	uppercaseCheck := widget.NewCheck("Использовать верхний регистр (ABCDEFGHIJKLMNOPQRSTUVWXYZ)", nil)
	uppercaseCheck.SetChecked(true)
	lowercaseCheck := widget.NewCheck("Использовать нижний регистр (abcdefghijklmnopqrstuvwxyz)", nil)
	lowercaseCheck.SetChecked(true)
	digitsCheck := widget.NewCheck("Использовать цифры (0123456789)", nil)
	digitsCheck.SetChecked(true)
	specialCheck := widget.NewCheck("Использовать спец-символы (!@#$%^&*-_=+;:,.?/~`)", nil)
	specialCheck.SetChecked(true)
	spaceCheck := widget.NewCheck("Использовать пробел", nil)
	bracketsCheck := widget.NewCheck("Использовать скобки ('[',']','{','}','(',')','<','>')", nil)

	generateBtn := widget.NewButton("Сгенерировать", func() {
		length, err := strconv.Atoi(lengthEntry.Text)
		if err != nil || length < 1 {
			dialog.ShowError(fmt.Errorf("Неверная длина"), win)
			return
		}
		options := models.PasswordGeneratorOptions{
			Length:       length,
			UseUppercase: uppercaseCheck.Checked,
			UseLowercase: lowercaseCheck.Checked,
			UseDigits:    digitsCheck.Checked,
			UseSpecial:   specialCheck.Checked,
			UseSpace:     spaceCheck.Checked,
			UseBrackets:  bracketsCheck.Checked,
		}
		pass, err := crypto.GeneratePassword(options)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		passwordEntry.SetText(pass)
	})

	copyBtn := widget.NewButtonWithIcon("Копировать", theme.ContentCopyIcon(), func() {
		if passwordEntry.Text != "" {
			fyne.CurrentApp().Clipboard().SetContent(passwordEntry.Text)
			dialog.ShowInformation("Копирование", "Пароль скопирован в буфер", win)
		}
	})

	form := container.NewVBox(
		widget.NewLabel("Длина пароля:"),
		lengthEntry,
		widget.NewLabel("Опции:"),
		uppercaseCheck,
		lowercaseCheck,
		digitsCheck,
		specialCheck,
		spaceCheck,
		bracketsCheck,
		widget.NewLabel("Пароль:"),
		passwordEntry,
		container.NewHBox(generateBtn, copyBtn),
		layout.NewSpacer(),
		widget.NewButton("Закрыть", func() {
			// Close popup
		}),
	)

	popup := widget.NewModalPopUp(container.NewPadded(form), win.Canvas())
	popup.Resize(fyne.NewSize(500, 400))

	// Set close action for the button
	form.Objects[len(form.Objects)-1].(*widget.Button).OnTapped = func() {
		popup.Hide()
	}

	popup.Show()
}

func showExportPopup(win fyne.Window, database *sql.DB, key []byte) {
	entries, err := db.LoadAllEntries(database, key)
	if err != nil {
		dialog.ShowError(err, win)
		return
	}

	fd := dialog.NewFileSave(func(uc fyne.URIWriteCloser, e error) {
		if uc != nil {
			defer uc.Close()
			writer := csv.NewWriter(uc)
			defer writer.Flush()

			// Заголовки
			headers := []string{"Title", "Username", "Password", "URL", "Notes", "Group"}
			if err := writer.Write(headers); err != nil {
				dialog.ShowError(err, win)
				return
			}

			// Данные
			for _, entry := range entries {
				record := []string{
					entry.Title,
					entry.Username,
					entry.Password,
					entry.URL,
					entry.Notes,
					entry.Group,
				}
				if err := writer.Write(record); err != nil {
					dialog.ShowError(err, win)
					return
				}
			}
			dialog.ShowInformation("Экспорт", "Пароли успешно экспортированы", win)
		}
	}, win)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
	fd.Resize(fyne.NewSize(800, 600))
	fd.Show()
}

func showImportPopup(win fyne.Window, database *sql.DB, key []byte, onImport func()) {
	fd := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if uc != nil {
			defer uc.Close()
			reader := csv.NewReader(uc)
			records, err := reader.ReadAll()
			if err != nil {
				dialog.ShowError(fmt.Errorf("Ошибка чтения CSV файла, проверьте его на корректность"), win)
				return
			}
			if len(records) < 2 {
				dialog.ShowError(fmt.Errorf("CSV файл должен содержать заголовки и хотя бы одну строку данных"), win)
				return
			}

			headers := records[0]
			data := records[1:]

			// Найти индексы колонок
			var titleIdx, usernameIdx, passwordIdx, urlIdx, notesIdx, groupIdx = -1, -1, -1, -1, -1, -1
			for i, h := range headers {
				switch strings.ToLower(h) {
				case "title":
					titleIdx = i
				case "username":
					usernameIdx = i
				case "password":
					passwordIdx = i
				case "url":
					urlIdx = i
				case "notes":
					notesIdx = i
				case "group":
					groupIdx = i
				}
			}

			// Проверка обязательных полей
			if usernameIdx == -1 || passwordIdx == -1 || urlIdx == -1 {
				dialog.ShowError(fmt.Errorf("Обязательные колонки: username, password, url"), win)
				return
			}

			// Найти максимальный индекс для проверки корректности строк
			maxIdx := usernameIdx
			if passwordIdx > maxIdx {
				maxIdx = passwordIdx
			}
			if urlIdx > maxIdx {
				maxIdx = urlIdx
			}
			if titleIdx > maxIdx {
				maxIdx = titleIdx
			}
			if notesIdx > maxIdx {
				maxIdx = notesIdx
			}
			if groupIdx > maxIdx {
				maxIdx = groupIdx
			}

			imported := 0
			for _, row := range data {
				if len(row) <= maxIdx {
					dialog.ShowError(fmt.Errorf("Некорректная строка: недостаточно колонок"), win)
					continue
				}

				username := strings.TrimSpace(row[usernameIdx])
				password := strings.TrimSpace(row[passwordIdx])
				url := strings.TrimSpace(row[urlIdx])

				if username == "" || password == "" || url == "" {
					continue // Пропустить неполные строки
				}

				title := ""
				if titleIdx != -1 && titleIdx < len(row) {
					title = strings.TrimSpace(row[titleIdx])
				}
				if title == "" {
					// Взять из URL без http
					title = extractTitleFromURL(url)
				}

				notes := ""
				if notesIdx != -1 && notesIdx < len(row) {
					notes = strings.TrimSpace(row[notesIdx])
				}

				group := ""
				if groupIdx != -1 && groupIdx < len(row) {
					group = strings.TrimSpace(row[groupIdx])
				}

				entry := models.PasswordEntry{
					Title:    title,
					Username: username,
					Password: password,
					URL:      url,
					Notes:    notes,
					Group:    group,
				}

				err := db.SaveEntry(database, key, entry)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Ошибка импорта записи: %v", err), win)
					return
				}
				imported++
			}

			dialog.ShowInformation("Импорт", fmt.Sprintf("Успешно импортировано %d записей", imported), win)
			if onImport != nil {
				onImport()
			}
		}
	}, win)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
	fd.Resize(fyne.NewSize(800, 600))
	fd.Show()
}

func extractTitleFromURL(url string) string {
	// Убрать http:// или https://
	if strings.HasPrefix(url, "https://") {
		url = url[8:]
	} else if strings.HasPrefix(url, "http://") {
		url = url[7:]
	}
	// Взять до первого /
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	return url
}
