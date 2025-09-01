package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PasswordEntry struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Notes    string `json:"notes"`
	Group    string `json:"group"`
}

type FilterSettings struct {
	Field string
	Query string
}

const addGroupNodeID = "__add_group__"
const AllEntryGroupName = "Все"

func main() {
	myApp := app.New()
	loginWindow := myApp.NewWindow("Password Book - Вход")
	loginWindow.Resize(fyne.NewSize(350, 250))
	loginWindow.CenterOnScreen()
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите мастер-пароль")

	loginButton := widget.NewButton("Войти", func() {
		if passwordEntry.Text == "" {
			showInfoPopup(loginWindow, "Ошибка", "Пароль не может быть пустым!")
			return
		}
		showMainWindow(myApp)
		loginWindow.Close()
	})

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

func showMainWindow(a fyne.App) {
	mainWindow := a.NewWindow("Password Book - Главное окно")
	mainWindow.Resize(fyne.NewSize(1200, 700))
	mainWindow.CenterOnScreen()
	allEntries := loadMockData()
	filterSettings := FilterSettings{Field: "all", Query: ""}
	currentGroup := AllEntryGroupName
	groups := buildGroupsStable(allEntries)
	visibleEntries := filterEntriesWithGroup(allEntries, filterSettings, currentGroup)

	// ---------- Список записей ----------
	var selectedEntry *PasswordEntry
	var list *widget.List
	list = widget.NewList(
		func() int { return len(visibleEntries) },
		func() fyne.CanvasObject {
			title := widget.NewLabel("")
			login := widget.NewLabel("")
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			editBtn.Importance = widget.LowImportance
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			deleteBtn.Importance = widget.LowImportance
			return container.NewBorder(nil, nil, nil,
				container.NewHBox(editBtn, deleteBtn),
				container.NewVBox(title, login),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			entry := visibleEntries[i]
			c := o.(*fyne.Container)
			vbox := c.Objects[0].(*fyne.Container)
			title := vbox.Objects[0].(*widget.Label)
			login := vbox.Objects[1].(*widget.Label)
			title.SetText(entry.Title)
			login.SetText(fmt.Sprintf("👤 %s", entry.Username))

			// Кнопки редактирования/удаления
			btns := c.Objects[1].(*fyne.Container)
			editBtn := btns.Objects[0].(*widget.Button)
			deleteBtn := btns.Objects[1].(*widget.Button)

			editBtn.OnTapped = func() {
				showInfoPopup(mainWindow, "Редактировать", "Редактирование записи пока не реализовано.")
			}
			deleteBtn.OnTapped = func() {
				visibleEntries = append(visibleEntries[:i], visibleEntries[i+1:]...)
				updateList(list, visibleEntries)
				list.UnselectAll()
			}
		},
	)

	// ---------- Детали записи ----------
	detailsTitle := widget.NewLabel("")
	detailsTitle.TextStyle = fyne.TextStyle{Bold: true}
	detailsContent := widget.NewLabel("")
	detailsContent.Wrapping = fyne.TextWrapWord
	showPassword := false

	btnToggle := widget.NewButton("Показать пароль", nil)
	btnCopy := widget.NewButton("Копировать пароль", nil)

	btns := container.NewHBox(btnToggle, btnCopy)
	detailsBox := container.NewVBox(detailsContent, btns)
	detailsCard := widget.NewCard("", "", container.NewVScroll(detailsBox))
	detailsCard.Hide()

	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(visibleEntries) {
			return
		}
		selectedEntry = &visibleEntries[id]
		showPassword = false
		refreshDetails(detailsTitle, detailsContent, selectedEntry, showPassword)
		detailsCard.SetTitle("Детали записи")
		detailsCard.Show()
	}

	btnToggle.OnTapped = func() {
		if selectedEntry == nil {
			return
		}
		showPassword = !showPassword
		refreshDetails(detailsTitle, detailsContent, selectedEntry, showPassword)
		if showPassword {
			btnToggle.SetText("Скрыть пароль")
		} else {
			btnToggle.SetText("Показать пароль")
		}
	}
	btnCopy.OnTapped = func() {
		if selectedEntry != nil {
			a.Clipboard().SetContent(selectedEntry.Password)
		}
	}

	// ---------- Дерево групп ----------
	tree := widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			if uid == "" {
				nodes := append([]string{}, groups...)
				nodes = append(nodes, addGroupNodeID)
				return nodes
			}
			return nil
		},
		func(uid widget.TreeNodeID) bool { return uid == "" },
		func(branch bool) fyne.CanvasObject {
			lbl := widget.NewLabel("")
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			editBtn.Importance = widget.LowImportance
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			deleteBtn.Importance = widget.LowImportance
			return container.NewBorder(nil, nil, nil,
				container.NewHBox(editBtn, deleteBtn),
				lbl,
			)
		},
		func(uid string, branch bool, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			lbl := c.Objects[0].(*widget.Label)
			btns := c.Objects[1].(*fyne.Container)
			editBtn := btns.Objects[0].(*widget.Button)
			deleteBtn := btns.Objects[1].(*widget.Button)

			if uid == addGroupNodeID {
				lbl.SetText("+ Добавить группу")
				editBtn.Hide()
				deleteBtn.Hide()
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				lbl.Refresh()
			} else if uid == AllEntryGroupName {
				lbl.SetText(AllEntryGroupName)
				editBtn.Hide()
				deleteBtn.Hide()
			} else {
				lbl.SetText(uid)
				editBtn.Show()
				deleteBtn.Show()
				editBtn.OnTapped = func() {
					showInfoPopup(mainWindow, "Редактировать", "Редактирование группы пока не реализовано.")
				}
				deleteBtn.OnTapped = func() {
					newGroups := []string{}
					for _, g := range groups {
						if g != uid {
							newGroups = append(newGroups, g)
						}
					}
					groups = newGroups
					list.Refresh()
				}
			}
		},
	)

	tree.OnSelected = func(uid string) {
		if uid == "" {
			return
		}
		if uid == addGroupNodeID {
			tree.Unselect(uid)
			showAddGroupForm(mainWindow, &groups, tree)
			return
		}
		currentGroup = uid
		visibleEntries = filterEntriesWithGroup(allEntries, filterSettings, currentGroup)
		updateList(list, visibleEntries)
		list.UnselectAll()
		detailsCard.Hide()
		detailsTitle.SetText("")
		detailsContent.SetText("")
	}

	// ---------- Toolbar ----------
	searchInput := widget.NewEntry()
	searchInput.SetPlaceHolder("Введите текст для поиска...")
	searchInput.OnChanged = func(query string) {
		filterSettings.Query = query
		visibleEntries = filterEntriesWithGroup(allEntries, filterSettings, currentGroup)
		updateList(list, visibleEntries)
		list.UnselectAll()
		detailsCard.Hide()
	}

	// выпадающий список для выбора поля фильтрации
	fieldSelect := widget.NewSelect([]string{
		"Все поля", "По названию", "По логину", "По URL", "По заметкам",
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
		visibleEntries = filterEntriesWithGroup(allEntries, filterSettings, currentGroup)
		updateList(list, visibleEntries)
		list.UnselectAll()
		detailsCard.Hide()
	})
	fieldSelect.SetSelected("Все поля")

	addButton := widget.NewButtonWithIcon("Добавить", theme.ContentAddIcon(), func() {
		showAddForm(mainWindow, &allEntries, list, &groups, &visibleEntries, filterSettings, &currentGroup, tree)
	})
	addButton.Importance = widget.HighImportance

	exitButton := widget.NewButtonWithIcon("Выход", theme.LogoutIcon(), func() {
		mainWindow.Close()
	})

	// теперь в тулбаре есть кнопка добавления, фильтр и поиск
	toolbar := container.NewBorder(nil, nil,
		container.NewHBox(addButton, fieldSelect),
		exitButton,
		searchInput,
	)

	// ---------- Макет ----------
	paddedToolbar := container.NewPadded(toolbar)
	splitRight := container.NewHSplit(list,
		container.NewBorder(detailsTitle, nil, nil, nil, detailsCard))
	splitRight.Offset = 0.35
	splitMain := container.NewHSplit(tree, splitRight)
	splitMain.Offset = 0.2

	mainContent := container.NewBorder(paddedToolbar, nil, nil, nil, splitMain)
	statusBar := widget.NewLabel(fmt.Sprintf("Загружено записей: %d | Версия: 1.0", len(allEntries)))
	fullContent := container.NewBorder(nil, statusBar, nil, nil, mainContent)

	mainWindow.SetContent(fullContent)
	mainWindow.Show()
}

func refreshDetails(title, content *widget.Label, entry *PasswordEntry, showPass bool) {
	title.SetText(entry.Title)
	pass := strings.Repeat("•", len(entry.Password))
	if showPass {
		pass = entry.Password
	}
	content.SetText(fmt.Sprintf(
		"Сайт: %s\nURL: %s\nГруппа: %s\nЛогин: %s\nПароль: %s\n\nЗаметки:\n%s",
		entry.Title, entry.URL, displayGroup(*entry),
		entry.Username, pass, entry.Notes,
	))
}

func filterEntriesWithGroup(all []PasswordEntry, settings FilterSettings, group string) []PasswordEntry {
	pool := all
	if group != "" && group != AllEntryGroupName {
		var buf []PasswordEntry
		for _, e := range all {
			if displayGroup(e) == group {
				buf = append(buf, e)
			}
		}
		pool = buf
	}
	return filterEntries(pool, settings)
}

func filterEntries(entries []PasswordEntry, settings FilterSettings) []PasswordEntry {
	// если запрос пуст – показываем всё (с учётом выбранной группы, это делает внешняя обёртка)
	if strings.TrimSpace(settings.Query) == "" {
		return entries
	}

	query := strings.ToLower(settings.Query)
	out := make([]PasswordEntry, 0, len(entries))

	for _, e := range entries {
		switch settings.Field {
		case "all":
			if strings.Contains(strings.ToLower(e.Title), query) ||
				strings.Contains(strings.ToLower(e.Username), query) ||
				strings.Contains(strings.ToLower(e.URL), query) ||
				strings.Contains(strings.ToLower(e.Notes), query) {
				out = append(out, e)
			}
		case "title":
			if strings.Contains(strings.ToLower(e.Title), query) {
				out = append(out, e)
			}
		case "username":
			if strings.Contains(strings.ToLower(e.Username), query) {
				out = append(out, e)
			}
		case "url":
			if strings.Contains(strings.ToLower(e.URL), query) {
				out = append(out, e)
			}
		case "notes":
			if strings.Contains(strings.ToLower(e.Notes), query) {
				out = append(out, e)
			}
		}
	}

	return out
}


func loadMockData() []PasswordEntry {
	jsonData := `[
		{"id":1,"title":"Google","username":"user@gmail.com","password":"MySecret123!","url":"https://google.com","notes":"Основной аккаунт","group":"Интернет"},
		{"id":2,"title":"Яндекс","username":"yandex_user","password":"Y@ndexP@ss","url":"https://yandex.ru","notes":"Рабочий аккаунт","group":"Работа"},
		{"id":3,"title":"GitHub","username":"dev_user","password":"G1tHubs3cur3","url":"https://github.com","notes":"Opensource","group":"Интернет"}
	]`
	var entries []PasswordEntry
	_ = json.Unmarshal([]byte(jsonData), &entries)
	return entries
}

func buildGroupsStable(entries []PasswordEntry) []string {
	seen := make(map[string]bool)
	out := []string{AllEntryGroupName}
	for _, e := range entries {
		g := displayGroup(e)
		if !seen[g] {
			seen[g] = true
			out = append(out, g)
		}
	}
	return out
}

func displayGroup(e PasswordEntry) string {
	if strings.TrimSpace(e.Group) == "" {
		return "Без группы"
	}
	return e.Group
}

func updateList(list *widget.List, entries []PasswordEntry) {
	list.Length = func() int { return len(entries) }
	list.Refresh()
}

func showAddForm(parent fyne.Window, entries *[]PasswordEntry, list *widget.List,
	groups *[]string, visibleEntries *[]PasswordEntry,
	filterSettings FilterSettings, currentGroup *string, tree *widget.Tree) {

	titleEntry := widget.NewEntry()
	urlEntry := widget.NewEntry()
	userEntry := widget.NewEntry()
	passEntry := widget.NewPasswordEntry()
	notesEntry := widget.NewMultiLineEntry()

	groupSelect := widget.NewSelect(*groups, nil)
	groupSelect.SetSelected(AllEntryGroupName)
	newGroupEntry := widget.NewEntry()
	newGroupEntry.SetPlaceHolder("Или введите новую группу")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Название", Widget: titleEntry},
			{Text: "URL", Widget: urlEntry},
			{Text: "Логин", Widget: userEntry},
			{Text: "Пароль", Widget: passEntry},
			{Text: "Заметки", Widget: notesEntry},
			{Text: "Группа", Widget: container.NewVBox(groupSelect, newGroupEntry)},
		},
		OnSubmit: func() {
			group := strings.TrimSpace(newGroupEntry.Text)
			if group == "" {
				group = groupSelect.Selected
			}
			if group == "" || group == AllEntryGroupName {
				group = "Без группы"
			}

			newEntry := PasswordEntry{
				ID: len(*entries) + 1, Title: titleEntry.Text, URL: urlEntry.Text,
				Username: userEntry.Text, Password: passEntry.Text,
				Notes: notesEntry.Text, Group: group,
			}
			*entries = append(*entries, newEntry)
			*groups = buildGroupsStable(*entries)
			tree.Refresh()
			*visibleEntries = filterEntriesWithGroup(*entries, filterSettings, *currentGroup)
			updateList(list, *visibleEntries)
			parent.Canvas().Content().Refresh()
		},
		SubmitText: "Сохранить", CancelText: "Отмена",
	}

	popup := widget.NewModalPopUp(container.NewVBox(
		widget.NewLabelWithStyle("Добавить новую запись", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		form,
	), parent.Canvas())
	form.OnCancel = popup.Hide
	popup.Show()
}

func showAddGroupForm(parent fyne.Window, groups *[]string, tree *widget.Tree) {
	groupEntry := widget.NewEntry()
	groupEntry.SetPlaceHolder("Название группы")
	form := &widget.Form{
		Items: []*widget.FormItem{{Text: "Группа", Widget: groupEntry}},
		OnSubmit: func() {
			g := strings.TrimSpace(groupEntry.Text)
			if g != "" && g != AllEntryGroupName {
				*groups = append(*groups, g)
				tree.Refresh()
			}
		},
		SubmitText: "Создать", CancelText: "Отмена",
	}
	popup := widget.NewModalPopUp(container.NewVBox(
		widget.NewLabelWithStyle("Добавить новую группу", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		form,
	), parent.Canvas())
	form.OnCancel = popup.Hide
	popup.Show()
}

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
