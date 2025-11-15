package app

import (
	"database/sql"

	"github.com/reinbowARA/PassLedger/db"
	"github.com/reinbowARA/PassLedger/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showAddForm(win fyne.Window, database *sql.DB, key []byte, onSave func(filters models.SearchFilters), editEntry ...*models.PasswordEntry) {
	var e models.PasswordEntry
	editMode := len(editEntry) > 0
	if editMode {
		e = *editEntry[0]
	}

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder(models.TITLE)
	titleEntry.SetText(e.Title)

	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder(models.LOGIN)
	loginEntry.SetText(e.Username)

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder(models.PASSWD)
	passEntry.SetText(e.Password)

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder(models.URL)
	urlEntry.SetText(e.URL)

	existingGroups := getUniqueGroupsFromDB(database, key)
	groupOptions := []string{}
	for _, g := range existingGroups {
		if g != "Все" {
			groupOptions = append(groupOptions, g)
		}
	}
	groupOptions = append(groupOptions, "") 
	// Создаем выпадающий список для существующих групп
	groupSelect := widget.NewSelect(groupOptions, nil)
	groupSelect.PlaceHolder = "Выберите существующую группу"

	if e.Group != "" {
		// Проверяем, есть ли текущая группа в списке
		found := false
		for _, g := range groupOptions {
			if g == e.Group {
				found = true
				groupSelect.SetSelected(e.Group)
				break
			}
		}
		// Если не нашли в списке, значит это новая группа
		if !found {
			groupSelect.Hide()
		}
	}

	// Создаем поле для ввода новой группы
	groupEntry := widget.NewEntry()
	groupEntry.SetPlaceHolder("Или введите новую группу")
	if e.Group != "" {
		// Проверяем, есть ли текущая группа в списке
		found := false
		for _, g := range groupOptions {
			if g == e.Group {
				found = true
				break
			}
		}
		// Если не нашли в списке, показываем её в поле ввода
		if !found {
			groupEntry.SetText(e.Group)
		} else {
			groupEntry.Hide()
		}
	}

	// Логика взаимного скрытия
	groupSelect.OnChanged = func(selected string) {
		if selected != "" {
			groupEntry.Hide()
			groupEntry.SetText("")
		} else {
			groupEntry.Show()
		}
	}

	groupEntry.OnChanged = func(text string) {
		if text != "" {
			groupSelect.Hide()
			groupSelect.SetSelected("")
		} else {
			groupSelect.Show()
		}
	}

	groupContainer := container.NewVBox(groupSelect, groupEntry)

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder(models.NOTES)
	notesEntry.SetText(e.Notes)

	form := widget.NewForm(
		widget.NewFormItem(models.TITLE, titleEntry),
		widget.NewFormItem(models.LOGIN, loginEntry),
		widget.NewFormItem(models.PASSWD, passEntry),
		widget.NewFormItem(models.URL, urlEntry),
		widget.NewFormItem(models.GROUP, groupContainer),
		widget.NewFormItem(models.NOTES, notesEntry),
	)

	saveBtn := func() {
		// Определяем группу: если выбрана из списка - используем её, иначе - из поля ввода
		selectedGroup := groupSelect.Selected
		if selectedGroup == "" {
			selectedGroup = groupEntry.Text
		}

		newEntry := models.PasswordEntry{
			Title:    titleEntry.Text,
			Username: loginEntry.Text,
			Password: passEntry.Text,
			URL:      urlEntry.Text,
			Group:    selectedGroup,
			Notes:    notesEntry.Text,
		}

		var err error
		if editMode {
			newEntry.ID = e.ID
			err = db.UpdateEntry(database, key, newEntry)
		} else {
			err = db.SaveEntry(database, key, newEntry)
		}

		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		onSave(models.SearchFilters{Title: true, Username: true, URL: true})
	}

	dialog.ShowCustomConfirm(
		"Добавить учётку",
		"Сохранить",
		"Отмена",
		form,
		func(ok bool) {
			if ok {
				saveBtn()
			}
		},
		win,
	)
}

func showAddGroup(win fyne.Window, database *sql.DB, key []byte, groupsSlice *[]string, groupList *widget.List) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Название новой группы")

	dialog.ShowCustomConfirm(
		"Добавить группу",
		"Создать",
		"Отмена",
		entry,
		func(ok bool) {
			if ok {
				name := entry.Text
				if name == "" {
					return
				}
				// не допускаем дубли
				for _, g := range *groupsSlice {
					if g == name {
						return
					}
				}
				// добавляем в db
				err := db.AddGroup(database, name)
				if err != nil {
					dialog.ShowError(err, win)
					return
				}
				// обновляем список групп из db
				*groupsSlice = getUniqueGroupsFromDB(database, key)
				groupList.Refresh()
			}
		},
		win,
	)
}

func showRenameGroup(win fyne.Window, oldName string, entries *[]models.PasswordEntry, groupsSlice *[]string, groupList *widget.List, database *sql.DB, key []byte, filters models.SearchFilters, onRefresh func()) {
	entry := widget.NewEntry()
	entry.SetText(oldName)
	dialog.ShowCustomConfirm(
		"Переименовать группу",
		"Сохранить",
		"Отмена",
		entry,
		func(ok bool) {
			if ok {
				newName := entry.Text
				if newName != "" && newName != oldName {
					db.UpdateGroup(database, oldName, newName)
					*groupsSlice = getUniqueGroupsFromDB(database, key)
					onRefresh()
					groupList.Refresh()
				}
			}
		},
		win,
	)
}
