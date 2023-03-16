package settings

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/satisfactorymodding/SatisfactoryModManager/utils"
	"github.com/spf13/viper"
)

type SavedModFilters struct {
	Order  string `json:"order"`
	Filter string `json:"filter"`
}

type View string

var (
	VIEW_COMPACT  View = "compact"
	VIEW_EXPANDED View = "expanded"
)

type SettingsData struct {
	FavoriteMods    []string          `json:"favoriteMods"`
	ModFilters      SavedModFilters   `json:"modFilters"`
	UnexpandedSize  utils.Size        `json:"unexpandedSize"`
	ExpandedSize    utils.Size        `json:"expandedSize"`
	StartView       View              `json:"startView"`
	QueueAutoStart  bool              `json:"queueAutoStart"`
	SelectedInstall string            `json:"selectedInstall"`
	SelectedProfile map[string]string `json:"selectedProfile"`
	ModsEnabled     map[string]bool   `json:"modsEnabled"`
	Offline         bool              `json:"offline"`
	Konami          bool              `json:"konami"`
	LaunchButton    string            `json:"launchButton"`
}

var Settings SettingsData = SettingsData{
	FavoriteMods: []string{},
	ModFilters: SavedModFilters{
		Order:  "Last updated",
		Filter: "Compatible",
	},
	UnexpandedSize:  utils.UnexpandedDefault,
	ExpandedSize:    utils.ExpandedDefault,
	StartView:       VIEW_COMPACT,
	QueueAutoStart:  true,
	SelectedInstall: "",
	SelectedProfile: map[string]string{},
	ModsEnabled:     map[string]bool{},
	Offline:         false,
	Konami:          false,
	LaunchButton:    "normal",
}
var settingsFileName = "settings.json"

func LoadSettings() error {
	settingsFilePath := filepath.Join(viper.GetString("local-dir"), settingsFileName)

	_, err := os.Stat(settingsFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to stat settings file")
		}

		err = SaveSettings()
		if err != nil {
			return errors.Wrap(err, "failed to save default settings")
		}
	}

	settingsFile, err := os.ReadFile(filepath.Join(viper.GetString("local-dir"), settingsFileName))
	if err != nil {
		return errors.Wrap(err, "failed to read settings")
	}

	if err := json.Unmarshal(settingsFile, &Settings); err != nil {
		return errors.Wrap(err, "failed to unmarshal settings")
	}

	return nil
}

func SaveSettings() error {
	settingsFile, err := json.MarshalIndent(Settings, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal settings")
	}
	err = os.WriteFile(filepath.Join(viper.GetString("local-dir"), settingsFileName), settingsFile, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to write settings")
	}

	return nil
}
