package app

import (
	"database/sql"

	"github.com/reinbowARA/PassLedger/db"
	"github.com/reinbowARA/PassLedger/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showAddForm(win fyne.Window, database *sql.DB, key []byte, onSave func(), editEntry ...*models.PasswordEntry) {
	var e models.PasswordEntry
	editMode := len(editEntry) > 0
	if editMode {
		e = *editEntry[0]
	}

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название")
	titleEntry.SetText(e.Title)

	loginEntry := widget.NewEntry()
	loginEntry.SetPlaceHolder("Логин")
	loginEntry.SetText(e.Username)

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Пароль")
	passEntry.SetText(e.Password)

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("URL")
	urlEntry.SetText(e.URL)

	groupEntry := widget.NewEntry()
	groupEntry.SetPlaceHolder("Группа")
	groupEntry.SetText(e.Group)

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Заметки")
	notesEntry.SetText(e.Notes)

	form := widget.NewForm(
		widget.NewFormItem("Название", titleEntry),
		widget.NewFormItem("Логин", loginEntry),
		widget.NewFormItem("Пароль", passEntry),
		widget.NewFormItem("URL", urlEntry),
		widget.NewFormItem("Группа", groupEntry),
		widget.NewFormItem("Заметки", notesEntry),
	)

	saveBtn := func() {
		newEntry := models.PasswordEntry{
			Title:    titleEntry.Text,
			Username: loginEntry.Text,
			Password: passEntry.Text,
			URL:      urlEntry.Text,
			Group:    groupEntry.Text,
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
		onSave()
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

// showAddGroup теперь принимает pointer на slice groups и саму группу list
func showAddGroup(win fyne.Window, groupsSlice *[]string, groupList *widget.List) {
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
				*groupsSlice = append(*groupsSlice, name)
				groupList.Refresh()
			}
		},
		win,
	)
}

func showRenameGroup(win fyne.Window, oldName string, entries *[]models.PasswordEntry, groupList *widget.List, database *sql.DB, key []byte) {
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
					_, _ = database.Exec(`UPDATE entries SET "group"=? WHERE "group"=?`, newName, oldName)
					refreshListFiltered(database, key, entries, win, "Все", "")
					groupList.Refresh()
				}
			}
		},
		win,
	)
}
