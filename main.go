package main

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"image/color"
	"strings"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Структура для хранения данных пароля
type PasswordEntry struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Notes    string `json:"notes"`
}

// FilterSettings настройки фильтрации
type FilterSettings struct {
	Field string
	Query string
}

func main() {
	myApp := app.New()
	loginWindow := myApp.NewWindow("Password Book - Вход")
	loginWindow.Resize(fyne.NewSize(350, 250))

	var masterPassword string
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите мастер-пароль")

	loginButton := widget.NewButton("Войти", func() {
		masterPassword = passwordEntry.Text
		if masterPassword == "" {
			showInfoPopup(loginWindow, "Ошибка", "Пароль не может быть пустым!")
			return
		}
		showMainWindow(myApp, masterPassword)
		loginWindow.Close()
	})

	// Красивое оформление
	title := canvas.NewText("Добро пожаловать в Password Book", color.NRGBA{R: 30, G: 100, B: 180, A: 255})
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		title,
		widget.NewLabel(""),
		widget.NewLabel("Для продолжения введите ваш мастер-пароль:"),
		passwordEntry,
		loginButton,
	)

	loginWindow.SetContent(content)
	loginWindow.ShowAndRun()
}

func showMainWindow(a fyne.App, password string) {
	mainWindow := a.NewWindow("Password Book - Главное окно")
	mainWindow.Resize(fyne.NewSize(1000, 700))

	// Загрузка фиктивных данных
	entries := loadMockData()
	// Привязки для настроек фильтра
	filterSettings := FilterSettings{
		Field: "all",
		Query: "",
	}

	// Создание списка записей
	list := widget.NewList(
		func() int { return len(entries) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil,
				widget.NewIcon(theme.DocumentIcon()),
				nil,
				container.NewVBox(
					widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabel(""),
				),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			entry := entries[i]
			mainContainer := o.(*fyne.Container)
			vbox := mainContainer.Objects[0].(*fyne.Container)

			titleLabel := vbox.Objects[0].(*widget.Label)
			titleLabel.SetText(entry.Title)

			loginLabel := vbox.Objects[1].(*widget.Label)
			loginLabel.SetText(fmt.Sprintf("👤 %s", entry.Username))
			loginLabel.Wrapping = fyne.TextWrapBreak
		},
	)
	// Выделенная область для деталей записи
	detailsTitle := widget.NewLabel("")
	detailsTitle.TextStyle = fyne.TextStyle{Bold: true}
	detailsTitle.Alignment = fyne.TextAlignCenter

	detailsContent := widget.NewLabel("")
	detailsContent.Wrapping = fyne.TextWrapWord

	detailsCard := widget.NewCard("", "", container.NewVScroll(detailsContent))
	detailsCard.Hide()

	// Обработчик выбора записи
	list.OnSelected = func(id widget.ListItemID) {
		var entry PasswordEntry
		for _, creds := range entries {
			if creds.ID-1 == id{
				entry = creds
			}
		}
		//entry := entries[id]

		detailsTitle.SetText(entry.Title)
		detailsContent.SetText(fmt.Sprintf(
			"Сайт: %s\nURL: %s\nЛогин: %s\nПароль: %s\n\nЗаметки:\n%s",
			entry.Title,
			entry.URL,
			entry.Username,
			strings.Repeat("•", len(entry.Password)), // Маскируем пароль
			entry.Notes,
		))

		detailsCard.SetTitle("Детали записи")
		detailsCard.Show()
	}

	// Улучшенная панель инструментов
	// Поле поиска
	searchInput := widget.NewEntry()
	searchInput.SetPlaceHolder("Введите текст для поиска...")
	searchInput.OnChanged = func(query string) {
		filterSettings.Query = query
		filtered := filterEntries(entries, filterSettings)
		updateList(list, filtered)
	}

	// Выпадающий список для выбора поля
	fieldSelect := widget.NewSelect([]string{
		"Все поля",
		"По названию",
		"По логину",
		"По URL",
		"По заметкам",
	}, func(selected string) {
		switch selected {
		case "Все поля":
			filterSettings.Field = "all"
		case "По названию":
			filterSettings.Field = "title"
		case "По логину":
			filterSettings.Field = "username"
		case "По URL":
			filterSettings.Field = "url"
		case "По заметкам":
			filterSettings.Field = "notes"
		}
		filtered := filterEntries(entries, filterSettings)
		updateList(list, filtered)
	})
	fieldSelect.SetSelected("Все поля") // Значение по умолчанию
	fieldSelect.PlaceHolder = "Фильтр по..."

	addButton := widget.NewButtonWithIcon("Добавить", theme.ContentAddIcon(), func() {
		showAddForm(mainWindow, &entries, list)
	})
	addButton.Importance = widget.HighImportance

	exitButton := widget.NewButtonWithIcon("Выход", theme.LogoutIcon(), func() {
		mainWindow.Close()
	})

	// Панель инструментов с улучшенным поиском
	toolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(addButton, fieldSelect), // Слева
		exitButton,  // Справа
		searchInput, // Центр (растягивается)
	)
	// Добавляем отступы
	paddedToolbar := container.NewPadded(toolbar)
	paddedList := container.NewPadded(list)
	paddedDetails := container.NewPadded(detailsCard)

	// Основной макет
	split := container.NewHSplit(
		paddedList,
		container.NewBorder(
			container.NewPadded(detailsTitle),
			nil, nil, nil,
			paddedDetails,
		),
	)
	split.Offset = 0.35 // Немного больше места для списка

	mainContent := container.NewBorder(
		paddedToolbar,
		nil,
		nil,
		nil,
		split,
	)
	// Статус бар
	statusBar := widget.NewLabel(fmt.Sprintf("Загружено записей: %d | Версия: 1.0", len(entries)))

	// Итоговый макет
	fullContent := container.NewBorder(
		nil,
		statusBar,
		nil,
		nil,
		mainContent,
	)

	mainWindow.SetContent(fullContent)
	mainWindow.Show()
}

// Фильтрация записей с учетом выбранного поля
func filterEntries(entries []PasswordEntry, settings FilterSettings) []PasswordEntry {
	if settings.Query == "" {
		return entries
	}

	query := strings.ToLower(settings.Query)
	filtered := []PasswordEntry{}

	for _, entry := range entries {
		found := false

		switch settings.Field {
		case "all":
			found = strings.Contains(strings.ToLower(entry.Title), query) ||
				strings.Contains(strings.ToLower(entry.Username), query) ||
				strings.Contains(strings.ToLower(entry.URL), query) ||
				strings.Contains(strings.ToLower(entry.Notes), query)

		case "title":
			found = strings.Contains(strings.ToLower(entry.Title), query)

		case "username":
			found = strings.Contains(strings.ToLower(entry.Username), query)

		case "url":
			found = strings.Contains(strings.ToLower(entry.URL), query)

		case "notes":
			found = strings.Contains(strings.ToLower(entry.Notes), query)
		}

		if found {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Загрузка фиктивных данных
func loadMockData() []PasswordEntry {
	// В реальном приложении будем загружать из файла/БД
	// Сейчас используем хардкодированный JSON
	jsonData := `[
		{
			"id": 1,
			"title": "Google",
			"username": "user@gmail.com",
			"password": "MySecret123!",
			"url": "https://google.com",
			"notes": "Основной аккаунт, включена 2FA"
		},
		{
			"id": 2,
			"title": "Яндекс",
			"username": "yandex_user",
			"password": "Y@ndexP@ss",
			"url": "https://yandex.ru",
			"notes": "Рабочий аккаунт"
		},
		{
			"id": 3,
			"title": "GitHub",
			"username": "dev_user",
			"password": "G1tHubs3cur3",
			"url": "https://github.com",
			"notes": "Для opensource проектов"
		},
		{
			"id": 4,
			"title": "Банк ВТБ",
			"username": "79101234567",
			"password": "Vtb2023Secure",
			"url": "https://vtb.ru",
			"notes": "Основной банковский счет"
		},
		{
			"id": 5,
			"title": "Steam",
			"username": "gamer123",
			"password": "SteamSummerSale",
			"url": "https://steampowered.com",
			"notes": "Аккаунт с играми"
		}
	]`

	var entries []PasswordEntry
	if err := json.Unmarshal([]byte(jsonData), &entries); err != nil {
		fmt.Println("Ошибка загрузки данных:", err)
		return []PasswordEntry{}
	}
	return entries
}

// Обновление списка
func updateList(list *widget.List, entries []PasswordEntry) {
	list.Length = func() int { return len(entries) }
	list.CreateItem = func() fyne.CanvasObject {
		return container.NewBorder(
			nil, nil,
			widget.NewIcon(theme.DocumentIcon()),
			nil,
			container.NewVBox(
				widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(""),
			),
		)
	}
	list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		entry := entries[id]
		mainContainer := item.(*fyne.Container)
		vbox := mainContainer.Objects[0].(*fyne.Container)

		titleLabel := vbox.Objects[0].(*widget.Label)
		titleLabel.SetText(entry.Title)

		loginLabel := vbox.Objects[1].(*widget.Label)
		loginLabel.SetText(fmt.Sprintf("👤 %s", entry.Username))
		loginLabel.Wrapping = fyne.TextWrapBreak
	}
	list.Refresh()
}

// Форма добавления новой записи
func showAddForm(parent fyne.Window, entries *[]PasswordEntry, list *widget.List) {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название сервиса")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("URL сайта")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Логин/Email")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Пароль")

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Заметки")
	notesEntry.Wrapping = fyne.TextWrapWord

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Название", Widget: titleEntry, HintText: "Обязательное поле"},
			{Text: "URL", Widget: urlEntry},
			{Text: "Логин", Widget: userEntry},
			{Text: "Пароль", Widget: passEntry},
			{Text: "Заметки", Widget: notesEntry},
		},
		//OnCancel: func() {},
		OnSubmit: func() {
			newEntry := PasswordEntry{
				ID:       len(*entries) + 1,
				Title:    titleEntry.Text,
				URL:      urlEntry.Text,
				Username: userEntry.Text,
				Password: passEntry.Text,
				Notes:    notesEntry.Text,
			}

			*entries = append(*entries, newEntry)
			list.Refresh()
			parent.Canvas().Content().Refresh()
		},
		SubmitText: "Сохранить",
		CancelText: "Отмена",
	}

	popup := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabelWithStyle("Добавить новую запись", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			form,
		),
		parent.Canvas(),
	)

	form.OnCancel = popup.Hide
	popup.Show()
}

// Вспомогательные функции для всплывающих окон
func showInfoPopup(parent fyne.Window, title, message string) {
	content := container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(message),
		widget.NewButton("OK", nil),
	)

	popup := widget.NewPopUp(content, parent.Canvas())
	content.Objects[2].(*widget.Button).OnTapped = popup.Hide
	popup.Show()
}
