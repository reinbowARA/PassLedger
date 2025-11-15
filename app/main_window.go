package app

import (
	"database/sql"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/reinbowARA/PassLedger/db"
	"github.com/reinbowARA/PassLedger/models"
)

func ShowMainWindow(a fyne.App, database *sql.DB, key []byte, entries []models.PasswordEntry) {
	win := a.NewWindow("Password Book")
	win.Resize(fyne.NewSize(1000, 600))
	win.CenterOnScreen()

	groupsSlice := getUniqueGroupsFromDB(database, key)
	var groupList *widget.List
	var list *widget.List
	detail := widget.NewRichText()
	detail.Wrapping = fyne.TextWrapWord
	currentGroup := "Все"
	searchText := ""

	// === Toolbar ===

	addBtn := widget.NewButtonWithIcon("Добавить", theme.ContentAddIcon(), func() {
		showAddForm(win, database, key, func() {
			refreshListFiltered(database, key, &entries, win, currentGroup, searchText)
			groupsSlice = getUniqueGroupsFromDB(database, key)
			groupList.Refresh()
		})
	})

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск...")
	searchBox := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(250, 36)),
		searchEntry,
	)

	searchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText = searchEntry.Text
		refreshListFiltered(database, key, &entries, win, currentGroup, searchText)
	})

	exitBtn := widget.NewButtonWithIcon("Выйти", theme.LogoutIcon(), func() {
		a.Quit()
	})
	toolbar := container.NewHBox(
		addBtn,
		layout.NewSpacer(),
		container.NewHBox(searchBox, searchBtn), // не сжимается
		layout.NewSpacer(),
		exitBtn,
	)

	// === Группы ===

	groupList = widget.NewList(
		func() int { return len(groupsSlice) + 1 }, // +1 для "+ Добавить группу"
		func() fyne.CanvasObject {
			// левая "кликабельная" часть — Button, справа — кнопки редактирования/удаления
			rowBtn := widget.NewButton("", nil)
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			return container.NewBorder(nil, nil, nil, container.NewHBox(editBtn, delBtn), rowBtn)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			// формируем список: groupsSlice + последняя нода как "+ Добавить группу"
			display := append(groupsSlice, "+ Добавить группу")
			name := display[i]

			// структура: Border( content=rowBtn, south=HBox(edit,del) )
			rowBtn := o.(*fyne.Container).Objects[0].(*widget.Button)
			btns := o.(*fyne.Container).Objects[1].(*fyne.Container)
			editBtn := btns.Objects[0].(*widget.Button)
			delBtn := btns.Objects[1].(*widget.Button)

			// Устанавливаем текст и поведение
			rowBtn.SetText(name)

			// Сценарии:
			if name == "+ Добавить группу" {
				// Сделать видной кнопку как Add (без иконок справа)
				editBtn.Hide()
				delBtn.Hide()
				rowBtn.Importance = widget.HighImportance
				rowBtn.OnTapped = func() {
					showAddGroup(win, database, key, &groupsSlice, groupList)
				}
				return
			}

			// Для группы "Все" запрещаем редактировать/удалять
			if name == "Все" {
				editBtn.Hide()
				delBtn.Hide()
			} else {
				editBtn.Show()
				delBtn.Show()
				editBtn.OnTapped = func() {
					showRenameGroup(win, name, &entries, &groupsSlice, groupList, database, key)
				}
				delBtn.OnTapped = func() {
					dialog.ShowConfirm("Удаление группы", "Удалить группу '"+name+"' и все её записи?", func(ok bool) {
						if ok {
							var id int
							id, err := db.DeleteEntriesInGroup(database, name)
							if err != nil {
								dialog.ShowError(err, win)
								return
							}
							err = db.DeleteGroup(database, id)
							if err != nil {
								dialog.ShowError(err, win)
								return
							}
							groupsSlice = getUniqueGroupsFromDB(database, key)
							groupList.Refresh()
							refreshListFiltered(database, key, &entries, win, "Все", "")
						}
					}, win)
				}
			}
			list.UnselectAll()

			// Нажатие на саму группу — фильтрация списка
			rowBtn.OnTapped = func() {
				currentGroup = name
				refreshListFiltered(database, key, &entries, win, currentGroup, searchText)
				list.Refresh()
				win.Content().Refresh()
				detail.ParseMarkdown("")
			}

		},
	)

	// === Учётки ===
	list = widget.NewList(
		func() int { return len(entries) },
		func() fyne.CanvasObject {
			title := widget.NewLabel("")
			login := widget.NewLabel("")
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			return container.NewBorder(nil, nil, nil,
				container.NewHBox(editBtn, deleteBtn),
				container.NewVBox(title, login))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(entries) {
				return
			}
			entry := entries[i]
			c := o.(*fyne.Container)
			vbox := c.Objects[0].(*fyne.Container)
			title := vbox.Objects[0].(*widget.Label)
			login := vbox.Objects[1].(*widget.Label)
			title.SetText(entry.Title)
			login.SetText(fmt.Sprintf("👤 %s", entry.Username))

			btns := c.Objects[1].(*fyne.Container)
			editBtn := btns.Objects[0].(*widget.Button)
			deleteBtn := btns.Objects[1].(*widget.Button)

			editBtn.OnTapped = func() {
				showAddForm(win, database, key, func() {
					refreshListFiltered(database, key, &entries, win, currentGroup, searchText)
				}, &entry)
			}
			deleteBtn.OnTapped = func() {
				dialog.ShowConfirm("Удаление", "Удалить запись?", func(ok bool) {
					if ok {
						db.DeleteEntry(database, entry.ID)
						refreshListFiltered(database, key, &entries, win, currentGroup, searchText)
					}
				}, win)
			}
		},
	)

	// === Панель деталей ====
	showPassBtn := widget.NewButtonWithIcon("Показать пароль", theme.VisibilityIcon(), nil)
	copyBtn := widget.NewButtonWithIcon("Скопировать", theme.ContentCopyIcon(), nil)
	passwordVisible := false
	selectedEntry := models.PasswordEntry{}

	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(entries) {
			detail.ParseMarkdown("") // очищаем
			return
		}
	}
	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(entries) {
			return
		}
		selectedEntry = entries[id]
		passwordVisible = false
		var text string = ShowEntry(selectedEntry, true)
		detail.ParseMarkdown(text)
		showPassBtn.SetText("Показать пароль")
		showPassBtn.Show() // ensure button is visible

		showPassBtn.OnTapped = func() {
			if passwordVisible {
				// Скрыть пароль
				var text string = ShowEntry(selectedEntry, true)
				detail.ParseMarkdown(text)
				showPassBtn.SetText("Показать пароль")
				passwordVisible = false
			} else {
				// Показать пароль
				var text string = ShowEntry(selectedEntry, false)
				detail.ParseMarkdown(text)
				showPassBtn.SetText("Скрыть пароль")
				passwordVisible = true
			}
		}

		copyBtn.OnTapped = func() {
			win.Clipboard().SetContent(selectedEntry.Password)
		}
	}

	detailPanel := container.New(
		layout.NewVBoxLayout(),
		container.NewPadded(detail),
		layout.NewSpacer(),
		container.NewHBox(
			container.NewPadded(showPassBtn),
			container.NewPadded(copyBtn),
		),
	)

	// === Макет ===
	vs := container.NewVSplit(list, detailPanel)
	//vs.SetOffset(0.2)
	mainContent := container.NewHSplit(groupList, vs)
	mainContent.SetOffset(0.2)

	content := container.NewBorder(toolbar, nil, nil, nil, mainContent)
	win.SetContent(content)
	win.Show()
}

func ShowEntry(entry models.PasswordEntry, hidePasswd bool) (text string) {
	if hidePasswd {
		entry.Password = maskPassword(entry.Password)
	}
	text = fmt.Sprintf(`
	Название: %s
	Логин: %s
	Пароль: %s
	URL: %s
	Заметки: %s `,
		entry.Title, entry.Username, entry.Password, entry.URL, entry.Notes)
	return
}
