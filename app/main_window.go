package app

import (
	"database/sql"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/storage"
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
	var table *widget.Table
	var popup *widget.PopUp
	detail := widget.NewRichText()
	detail.Wrapping = fyne.TextWrapWord
	currentGroup := models.DefaultNameAllGroups
	searchText := ""
	currentFilters := models.SearchFilters{Title: true, Username: true, URL: true}
	var selectedRow = -1

	// === Toolbar ===

	addBtn := widget.NewButtonWithIcon("Добавить", theme.ContentAddIcon(), func() {
		showAddForm(win, database, key, func(filters models.SearchFilters) {
			currentFilters = filters
			refreshListFiltered(database, key, &entries, win, currentGroup, searchText, currentFilters, detail)
			groupsSlice = getUniqueGroupsFromDB(database, key)
			groupList.Refresh()
		})
	})

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск...")
	searchEntry.OnChanged = func(text string) {
		searchText = text
		refreshListFiltered(database, key, &entries, win, currentGroup, searchText, currentFilters, detail)
	}
	searchBox := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(250, 36)),
		searchEntry,
	)

	// Кнопка для настройки фильтров
	filterBtn := widget.NewButtonWithIcon("Фильтры", theme.MenuIcon(), func() {
		showFilterDialog(win, &currentFilters, func() {
			refreshListFiltered(database, key, &entries, win, currentGroup, searchText, currentFilters, detail)
		})
	})

	settings, err := LoadSettings()
	if err != nil {
		dialog.ShowError(err, win)
		return
	}

	SettingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		showSettingsForm(win, &settings, a, func(newSettings models.Settings) {
			settings = newSettings
			err := SaveSettings(settings)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			// themeStr := "dark"
			// if newSettings.ThemeVariant == 0 {
			// 	themeStr = "light"
			// }
			// a.Preferences().SetString("Theme", themeStr)
			//dialog.ShowInformation("Настройки", "Настройки сохранены. Для применения новой темы перезапустите приложение.", win)
		})
	})

	exitBtn := widget.NewButtonWithIcon("Выйти", theme.LogoutIcon(), func() {
		a.Quit()
	})
	toolbar := container.NewHBox(
		addBtn,
		layout.NewSpacer(),
		container.NewHBox(searchBox, filterBtn),
		layout.NewSpacer(),
		SettingsBtn,
		exitBtn,
	)

	copyBtn := widget.NewButtonWithIcon("Скопировать пароль", theme.ContentCopyIcon(), nil)
	timerProgress := widget.NewProgressBar()
	timerProgress.TextFormatter = func() string {
		return ""
	}
	timerProgress.Hide()
	timerProgress.Max = 1
	timerProgress.Min = 0
	timerProgress.Value = 0
	timerLabel := widget.NewLabel("")
	timerLabel.Hide()
	var cancel chan struct{}

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

			// Для группы models.DefaultNameAllGroups запрещаем редактировать/удалять
			if name == models.DefaultNameAllGroups {
				editBtn.Hide()
				delBtn.Hide()
			} else {
				editBtn.Show()
				delBtn.Show()
				editBtn.OnTapped = func() {
					showRenameGroup(win, name, &entries, &groupsSlice, groupList, database, key, currentFilters, func() {
						refreshListFiltered(database, key, &entries, win, models.DefaultNameAllGroups, "", currentFilters, detail)
					})
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
							refreshListFiltered(database, key, &entries, win, models.DefaultNameAllGroups, "", currentFilters, detail)
						}
					}, win)
				}
			}
			// Нажатие на саму группу — фильтрация списка
			rowBtn.OnTapped = func() {
				selectedRow = -1
				currentGroup = name
				refreshListFiltered(database, key, &entries, win, currentGroup, searchText, currentFilters, detail)
				table.Refresh()
				win.Content().Refresh()
				detail.ParseMarkdown("")
			}

		},
	)

	// === Учётки ===
	table = widget.NewTableWithHeaders(
		func() (int, int) { return len(entries), 5 }, // 5 колонок: Title, Username, URL, Group, Actions
		func() fyne.CanvasObject {
			return widget.NewButton("", nil)
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			button, ok := o.(*widget.Button)
			if !ok {
				return
			}
			if i.Row < 0 || i.Row >= len(entries) {
				button.SetText("")
				return
			}
			entry := entries[i.Row]
			setOnTapped := func() {
				selectedRow = i.Row
				table.Refresh()
				var text string = ShowEntry(entry, true)
				detail.ParseMarkdown(text)
				copyBtn.OnTapped = func() {
					if cancel != nil {
						close(cancel)
					}
					cancel = make(chan struct{})
					a.Clipboard().SetContent(entry.Password)
					go runTimer(a, timerProgress, timerLabel, win, cancel, settings.TimerSeconds)
				}
			}
			// Установка выделения строки
			if i.Row == selectedRow {
				if i.Col < 4 {
					button.Importance = widget.WarningImportance // Выделение выбранной строки
				} else {
					button.Importance = widget.HighImportance
				}
			} else {
				if i.Col < 4 {
					button.Importance = widget.LowImportance // Нормальный стиль
				} else {
					button.Importance = widget.HighImportance
				}
			}
			switch i.Col {
			case 0:
				button.SetIcon(nil)
				button.SetText(entry.Title)
				button.OnTapped = setOnTapped
			case 1:
				button.SetIcon(nil)
				button.SetText(entry.Username)
				button.OnTapped = setOnTapped
			case 2:
				button.SetIcon(nil)
				button.SetText(entry.URL)
				button.OnTapped = setOnTapped
			case 3:
				button.SetIcon(nil)
				button.SetText(entry.Group)
				button.OnTapped = setOnTapped
			case 4:
				button.SetIcon(theme.SettingsIcon())
				button.SetText("")
				button.OnTapped = func() {
					selectedRow = i.Row
					table.Refresh()
					buttonEdit := widget.NewButton("Редактировать", func() {
						showAddForm(win, database, key, func(filters models.SearchFilters) {
							currentFilters = filters
							refreshListFiltered(database, key, &entries, win, currentGroup, searchText, currentFilters, detail)
							groupsSlice = getUniqueGroupsFromDB(database, key)
							groupList.Refresh()
							popup.Hide()
						}, &entry)
					})
					buttonDelete := widget.NewButton("Удалить", func() {
						dialog.ShowConfirm("Удаление", "Удалить запись?", func(ok bool) {
							if ok {
								db.DeleteEntry(database, entry.ID)
								refreshListFiltered(database, key, &entries, win, currentGroup, searchText, currentFilters, detail)
								groupsSlice = getUniqueGroupsFromDB(database, key)
								groupList.Refresh()
								popup.Hide()
							}
						}, win)
					})

					closeBtn := widget.NewButton("Отмена", func() {
						popup.Hide()
					})

					content := container.NewVBox(buttonEdit, buttonDelete, closeBtn)
					popup = widget.NewModalPopUp(content, win.Canvas())
					popup.Show()
				}
			}
		},
	)
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		label, ok := template.(*widget.Label)
		if !ok {
			return
		}
		if id.Row < 0 {
			// Заголовки колонок
			switch id.Col {
			case 0:
				label.SetText(models.TITLE)
			case 1:
				label.SetText(models.LOGIN)
			case 2:
				label.SetText(models.URL)
			case 3:
				label.SetText(models.GROUP)
			case 4:
				label.SetText("")
			}
		} else if id.Col < 0 {
			// Заголовки строк
			label.SetText(fmt.Sprintf("%d", id.Row+1))
		}
	}
	table.SetColumnWidth(0, 150) // Title
	table.SetColumnWidth(1, 150) // Username
	table.SetColumnWidth(2, 175) // URL
	table.SetColumnWidth(3, 100) // Group
	table.SetColumnWidth(4, 50)  // Actions

	// === Панель деталей ====
	copyBtn = widget.NewButtonWithIcon("Скопировать пароль", theme.ContentCopyIcon(), nil)
	timerProgress = widget.NewProgressBar()
	timerProgress.TextFormatter = func() string {
		return ""
	}
	timerProgress.Hide()
	timerProgress.Max = 1
	timerProgress.Min = 0
	timerProgress.Value = 0
	timerLabel = widget.NewLabel("")
	timerLabel.Hide()

	detailPanel := container.New(
		layout.NewVBoxLayout(),
		container.NewPadded(detail),
		layout.NewSpacer(),
		container.NewHBox(
			container.NewPadded(copyBtn),
			container.NewPadded(timerLabel),
			container.NewPadded(timerProgress),
		),
	)

	// === Макет ===
	vs := container.NewVSplit(table, detailPanel)
	//vs.SetOffset(0.2)
	mainContent := container.NewHSplit(groupList, vs)
	mainContent.SetOffset(0.2)

	content := container.NewBorder(toolbar, nil, nil, nil, mainContent)
	win.SetContent(content)
	win.Show()
}

func runTimer(a fyne.App, progress *widget.ProgressBar, timerLabel *widget.Label, win fyne.Window, cancel <-chan struct{}, timerSeconds int) {
	fyne.DoAndWait(func() {
		progress.SetValue(1.0)
		progress.TextFormatter = func() string {
			return fmt.Sprintf("%d сек", timerSeconds)
		}
		timerLabel.SetText("До очистки буфера осталось: ")
		timerLabel.Show()
		progress.Show()
	})
	for i := timerSeconds; i >= 0; i-- {
		select {
		case <-cancel:
			return
		case <-time.After(time.Second):
		}
		secLeft := i
		fyne.DoAndWait(func() {
			if secLeft == 0 {
				a.Clipboard().SetContent("")
				timerLabel.Hide()
				progress.Hide()
			} else {
				progress.TextFormatter = func() string {
					return fmt.Sprintf("%d сек", secLeft)
				}
				progress.SetValue(float64(secLeft) / float64(timerSeconds))
				timerLabel.SetText("До очистки буфера осталось: ")
			}
		})
		if secLeft == 0 {
			return
		}
	}
}

func ShowEntry(entry models.PasswordEntry, hidePasswd bool) (text string) {
	if hidePasswd {
		entry.Password = maskPassword(entry.Password)
	}
	text = fmt.Sprintf(`
**Название:** %s
**Группа:** %s
**Логин:** %s
**Пароль:** %s
**URL:** %s
**Заметки:** %s `,
		entry.Title, entry.Group, entry.Username, entry.Password, entry.URL, entry.Notes)
	return
}

func showSettingsForm(parent fyne.Window, currentSettings *models.Settings, a fyne.App, onSave func(models.Settings)) {
	settingsWin := fyne.CurrentApp().NewWindow("Настройки")
	settingsWin.Resize(fyne.NewSize(600, 400))
	settingsWin.CenterOnScreen()

	applied := false
	originalSettings := *currentSettings
	tempSettings := *currentSettings // desired changes

	var lightBtn, darkBtn *widget.Button

	dbPathEntry := widget.NewEntry()
	dbPathEntry.SetText(currentSettings.DBPath)
	dbPathEntry.SetPlaceHolder("Путь к файлу базы данных")
	dbPathEntry.Disable() // Делаем его не редактируемым, только через выбор

	selectBtn := widget.NewButtonWithIcon("Выбрать файл", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
			if uc != nil {
				path := uc.URI().Path()
				dbPathEntry.SetText(path)
			}
		}, settingsWin)
		filter := storage.NewExtensionFileFilter([]string{".db"})
		fd.SetFilter(filter)
		fd.Resize(fyne.NewSize(800, 600))
		fd.Show()
	})

	createBtn := widget.NewButtonWithIcon("Создать файл", theme.ContentAddIcon(), func() {
		fd := dialog.NewFileSave(func(uc fyne.URIWriteCloser, e error) {
			if uc != nil {
				path := uc.URI().Path()
				dbPathEntry.SetText(path)
			}
		}, settingsWin)
		filter := storage.NewExtensionFileFilter([]string{".db"})
		fd.SetFilter(filter)
		fd.Resize(fyne.NewSize(800, 600))
		fd.Show()
	})

	dbPathContainer := container.NewBorder(container.NewHBox(selectBtn, createBtn), nil, nil, nil, dbPathEntry)

	lightBtn = widget.NewButton("Светлая", func() {
		tempSettings.ThemeVariant = 0
		lightBtn.Importance = widget.HighImportance
		darkBtn.Importance = widget.LowImportance
		settingsWin.Content().Refresh()
	})
	darkBtn = widget.NewButton("Темная", func() {
		tempSettings.ThemeVariant = 1
		darkBtn.Importance = widget.HighImportance
		lightBtn.Importance = widget.LowImportance
		settingsWin.Content().Refresh()
	})
	if tempSettings.ThemeVariant == 0 {
		lightBtn.Importance = widget.HighImportance
		darkBtn.Importance = widget.LowImportance
	} else {
		lightBtn.Importance = widget.LowImportance
		darkBtn.Importance = widget.HighImportance
	}
	themeContainer := container.NewGridWithColumns(2, lightBtn, darkBtn)

	timerSlider := widget.NewSlider(1, 60)
	timerSlider.SetValue(float64(tempSettings.TimerSeconds))
	timerLabel := widget.NewLabel(fmt.Sprintf("%d сек", tempSettings.TimerSeconds))
	timerSlider.OnChanged = func(value float64) {
		tempSettings.TimerSeconds = int(value)
		timerLabel.SetText(fmt.Sprintf("%.0f сек", value))
	}

	timerContainer := container.NewVBox(timerSlider, timerLabel)

	form := widget.NewForm(
		widget.NewFormItem("Путь к БД*", dbPathContainer),
		widget.NewFormItem("Тема", themeContainer),
		widget.NewFormItem("Таймер очистки буфера (сек)", timerContainer),
	)

	saveBtn := widget.NewButtonWithIcon("Сохранить", theme.ConfirmIcon(), func() {
		if !applied {
			return
		}
		newSettings := models.Settings{
			DBPath:       dbPathEntry.Text,
			ThemeVariant: currentSettings.ThemeVariant,
			TimerSeconds: currentSettings.TimerSeconds,
		}
		if dbPathEntry.Text != originalSettings.DBPath {
			dialog.ShowConfirm("Информация",
				"Приложение будет закрыто для применения изменений пути к БД. Перезапустите приложение вручную.",
				func(b bool) {
					if b {
						onSave(newSettings)
						settingsWin.Close()
						a.Quit()
					} 
				},
				settingsWin)
		} else {
			onSave(newSettings)
			settingsWin.Close()
		}
	})
	saveBtn.Disable()
	saveBtn.Importance = widget.SuccessImportance

	applyBtn := widget.NewButtonWithIcon("Применить", theme.ConfirmIcon(), func() {
		applied = true
		*currentSettings = tempSettings
		if tempSettings.ThemeVariant == 0 {
			a.Settings().SetTheme(theme.LightTheme())
		} else {
			a.Settings().SetTheme(theme.DarkTheme())
		}
		saveBtn.Enable()
		settingsWin.Content().Refresh()
	})

	cancelBtn := widget.NewButtonWithIcon("Отмена", theme.CancelIcon(), func() {
		if applied {
			*currentSettings = originalSettings
			if originalSettings.ThemeVariant == 0 {
				a.Settings().SetTheme(theme.LightTheme())
			} else {
				a.Settings().SetTheme(theme.DarkTheme())
			}
		}
		settingsWin.Close()
	})

	content := container.NewVBox(
		form,
		layout.NewSpacer(),
		widget.NewRichTextWithText("* - изменения вступят в силу после перезапуска!"),
		container.NewHBox(
			layout.NewSpacer(),
			applyBtn,
			saveBtn,
			cancelBtn,
		),
	)

	settingsWin.SetContent(container.NewPadded(content))
	settingsWin.SetCloseIntercept(func() {
		if applied {
			*currentSettings = originalSettings
			if originalSettings.ThemeVariant == 0 {
				a.Settings().SetTheme(theme.LightTheme())
			} else {
				a.Settings().SetTheme(theme.DarkTheme())
			}
		}
		settingsWin.Close()
	})
	settingsWin.Show()
}
