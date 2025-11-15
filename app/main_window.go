package app

import (
	"database/sql"
	"fmt"
	"strings"

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

	// делаем pointer-like переменные которые используются в замыканиях:
	groupsSlice := getUniqueGroups(entries)
	var groupList *widget.List // обязательно объявить до использования
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
		layout.NewGridWrapLayout(fyne.NewSize(250, 36)), // фиксируем ширину и высоту
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
					// добавляем новую группу в db (и обновляем список)
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
					showRenameGroup(win, name, &entries, groupList, database, key)
				}
				delBtn.OnTapped = func() {
					dialog.ShowConfirm("Удаление группы", "Удалить группу '"+name+"' и все её записи?", func(ok bool) {
						if ok {
							// получаем id группы по имени
							var id int
							err := database.QueryRow(`SELECT id FROM groups WHERE name = ?`, name).Scan(&id)
							if err != nil {
								if err == sql.ErrNoRows {
									dialog.ShowError(fmt.Errorf("Группа '%s' не найдена", name), win)
								} else {
									dialog.ShowError(err, win)
								}
								return
							}
							// удаляем все записи в группе
							_, err = database.Exec(`DELETE FROM entries WHERE group_id = ?`, id)
							if err != nil {
								dialog.ShowError(err, win)
								return
							}
							// удаляем группу
							err = db.DeleteGroup(database, id)
							if err != nil {
								dialog.ShowError(err, win)
								return
							}
							// пересобираем groupsSlice и список
							groupsSlice = getUniqueGroupsFromDB(database, key)
							groupList.Refresh()
							// обновляем записи, показывая "Все"
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
				detail.ParseMarkdown("") // сбрасываем детальную панель
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
		e := entries[id]
		var text string = fmt.Sprintf(`
	Название: %s 
	Логин: %s 
	Пароль: %s 
	URL: %s 
	Заметки: %s `, 
		e.Title, e.Username, maskPassword(e.Password), e.URL, e.Notes)
		detail.ParseMarkdown(text) //TODO переделать показ пароля

		showPassBtn.OnTapped = func() {
			var text string = fmt.Sprintf(`
	Название: %s 
	Логин: %s 
	Пароль: %s 
	URL: %s 
	Заметки: %s `, 
		e.Title, e.Username, e.Password, e.URL, e.Notes)
			detail.ParseMarkdown(text)
		}

		copyBtn.OnTapped = func() {
			win.Clipboard().SetContent(e.Password)
		}
	}

	//TODO
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

// --- вспомогательные функции UI ---

func maskPassword(p string) string {
	if len(p) == 0 {
		return ""
	}
	return "********"
}

func refreshListFiltered(database *sql.DB, key []byte, entries *[]models.PasswordEntry, win fyne.Window, group, query string) {
	all, err := db.LoadAllEntries(database, key)
	if err != nil {
		ShowInfo(win, "Ошибка", "Не удалось загрузить записи: "+err.Error())
		return
	}

	filtered := []models.PasswordEntry{}
	for _, e := range all {
		if group != "" && group != "Все" && e.Group != group {
			continue
		}
		if query != "" {
			q := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(e.Title), q) &&
				!strings.Contains(strings.ToLower(e.Username), q) &&
				!strings.Contains(strings.ToLower(e.URL), q) {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	*entries = filtered
	win.Content().Refresh()
}
