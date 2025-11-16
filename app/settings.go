package app

import (
	"encoding/json"
	"os"

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
