package app

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/reinbowARA/PassLedger/db"
	"github.com/reinbowARA/PassLedger/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowInfo — простой попап
func ShowInfo(win fyne.Window, title, message string) {
	content := container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(message),
		widget.NewButton("OK", func() { win.Close() /* не закрываем окно приложения! */ }),
	)
	// используем ModalPopUp, если есть canvas
	pop := widget.NewModalPopUp(content, win.Canvas())
	pop.Show()
}

// getUniqueGroupsFromDB грузит свежие группы из DB
func getUniqueGroupsFromDB(database *sql.DB, key []byte) []string {
	groups, err := db.GetGroup(database)
	if err != nil {
		// если ошибка — возвращаем пустой набор кроме models.DefaultNameAllGroups
		return []string{models.DefaultNameAllGroups}
	}
	out := []string{models.DefaultNameAllGroups}
	for _, g := range groups {
		out = append(out, g.Name)
	}
	return out
}

func maskPassword(p string) string {
	if len(p) == 0 {
		return ""
	}
	return "********"
}

func refreshListFiltered(database *sql.DB, key []byte, entries *[]models.PasswordEntry, win fyne.Window, group, query string, filters models.SearchFilters, detail *widget.RichText) {
	all, err := db.LoadAllEntries(database, key)
	if err != nil {
		ShowInfo(win, "Ошибка", "Не удалось загрузить записи: "+err.Error())
		return
	}

	filtered := []models.PasswordEntry{}
	for _, e := range all {
		if group != "" && group != models.DefaultNameAllGroups && e.Group != group {
			continue
		}
		if query != "" {
			q := strings.ToLower(query)
			matches := false
			if filters.Title && strings.Contains(strings.ToLower(e.Title), q) {
				matches = true
			}
			if filters.Username && strings.Contains(strings.ToLower(e.Username), q) {
				matches = true
			}
			if filters.URL && strings.Contains(strings.ToLower(e.URL), q) {
				matches = true
			}
			if filters.Group && strings.Contains(strings.ToLower(e.Group), q) {
				matches = true
			}
			if filters.Notes && strings.Contains(strings.ToLower(e.Notes), q) {
				matches = true
			}
			if !matches {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	*entries = filtered
	if len(filtered) == 0 && query != "" {
		detail.ParseMarkdown("# Ничего не найдено\n\nПо запросу: `" + query + "`")
	} else {
		detail.ParseMarkdown("") // очищаем сообщение
	}

	win.Content().Refresh()
}

func showFilterDialog(win fyne.Window, filters *models.SearchFilters, onChange func()) {
	titleCb := widget.NewCheck(models.TITLE, nil)
	titleCb.SetChecked(filters.Title)

	usernameCb := widget.NewCheck(models.LOGIN, nil)
	usernameCb.SetChecked(filters.Username)

	urlCb := widget.NewCheck(models.URL, nil)
	urlCb.SetChecked(filters.URL)

	groupCb := widget.NewCheck(models.GROUP, nil)
	groupCb.SetChecked(filters.Group)

	notesCb := widget.NewCheck(models.NOTES, nil)
	notesCb.SetChecked(filters.Notes)

	content := container.NewVBox(
		titleCb,
		usernameCb,
		urlCb,
		groupCb,
		notesCb,
	)

	dialog.ShowCustomConfirm("Выберите поля для поиска", models.CONFIRM, models.CANCEL, content, func(ok bool) {
		if ok {
			filters.Title = titleCb.Checked
			filters.Username = usernameCb.Checked
			filters.URL = urlCb.Checked
			filters.Group = groupCb.Checked
			filters.Notes = notesCb.Checked
			onChange()
		}
	}, win)
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
