package app

import (
	"database/sql"
	"strings"

	"github.com/reinbowARA/PassLedger/db"
	"github.com/reinbowARA/PassLedger/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

// getUniqueGroups берёт группы из текущего слайса entries и возвращает с "Все" первым
func getUniqueGroups(entries []models.PasswordEntry) []string {
	seen := map[string]bool{}
	out := []string{"Все"}
	for _, e := range entries {
		g := e.Group
		if g == "" || g == "Все" {
			continue
		}
		if !seen[g] {
			seen[g] = true
			out = append(out, g)
		}
	}
	return out
}

// getUniqueGroupsFromDB грузит свежие группы из DB
func getUniqueGroupsFromDB(database *sql.DB, key []byte) []string {
	groups, err := db.GetGroup(database)
	if err != nil {
		// если ошибка — возвращаем пустой набор кроме "Все"
		return []string{"Все"}
	}
	out := []string{"Все"}
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
