package app

import (
	"encoding/json"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/reinbowARA/PassLedger/models"
)

const settingsFile = "settings.json"

func LoadSettings() (models.Settings, error) {
	settings := models.Settings{
		DBPath:       models.DefaultDBPath,
		ThemeVariant: 1,
		TimerSeconds: models.TIME_CLEAR_PASSWD,
	}

	file, err := os.Open(settingsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil // Default settings
		}
		return settings, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&settings)
	return settings, err
}

func SaveSettings(settings models.Settings) error {
	file, err := os.Create(settingsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(settings)
}

func showSettingsForm(parent fyne.Window, currentSettings *models.Settings, a fyne.App, overlay *widget.PopUp, settingsWindowOpen *bool, onSave func(models.Settings)) {
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
			ThemeVariant: tempSettings.ThemeVariant,
			TimerSeconds: tempSettings.TimerSeconds,
		}
		onSave(newSettings)
		overlay.Hide()
		settingsWin.Close()
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
		overlay.Hide()
		*settingsWindowOpen = false
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
		overlay.Hide()
		*settingsWindowOpen = false
		settingsWin.Close()
	})
	settingsWin.Show()
}
