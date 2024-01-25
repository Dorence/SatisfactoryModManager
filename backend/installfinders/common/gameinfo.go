package common

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type installType struct {
	versionPath string
	executable  string
	installType InstallType
}

var gameInfo = []installType{
	{
		executable:  "FactoryServer.sh",
		versionPath: filepath.Join("Engine", "Binaries", "Linux", "UnrealServer-Linux-Shipping.version"),
		installType: InstallTypeLinuxServer,
	},
	{
		executable:  "FactoryServer.exe",
		versionPath: filepath.Join("Engine", "Binaries", "Win64", "UnrealServer-Win64-Shipping.version"),
		installType: InstallTypeWindowsServer,
	},
	{
		executable:  "FactoryGame.exe",
		versionPath: filepath.Join("Engine", "Binaries", "Win64", "FactoryGame-Win64-Shipping.version"),
		installType: InstallTypeWindowsClient,
	},
	// Update 9 stuff below
	{
		executable:  "FactoryServer.sh",
		versionPath: filepath.Join("Engine", "Binaries", "Linux", "FactoryServer-Linux-Shipping.version"),
		installType: InstallTypeLinuxServer,
	},
	{
		executable:  "FactoryServer.exe",
		versionPath: filepath.Join("Engine", "Binaries", "Win64", "FactoryServer-Win64-Shipping.version"),
		installType: InstallTypeWindowsServer,
	},
}

type GameVersionFile struct {
	MajorVersion         int    `json:"MajorVersion"`
	MinorVersion         int    `json:"MinorVersion"`
	PatchVersion         int    `json:"PatchVersion"`
	Changelist           int    `json:"Changelist"`
	CompatibleChangelist int    `json:"CompatibleChangelist"`
	IsLicenseeVersion    int    `json:"IsLicenseeVersion"`
	IsPromotedBuild      int    `json:"IsPromotedBuild"`
	BranchName           string `json:"BranchName"`
	BuildID              string `json:"BuildId"`
}

func GetGameInfo(path string) (InstallType, int, error) {
	for _, info := range gameInfo {
		executablePath := filepath.Join(path, info.executable)
		if _, err := os.Stat(executablePath); os.IsNotExist(err) {
			slog.Debug("game not of type", slog.String("path", executablePath), slog.String("type", string(info.installType)))
			continue
		}

		versionFilePath := filepath.Join(path, info.versionPath)
		if _, err := os.Stat(versionFilePath); os.IsNotExist(err) {
			return InstallTypeWindowsClient, 0, errors.Errorf("failed to get game info")
		}

		versionFile, err := os.ReadFile(versionFilePath)
		if err != nil {
			return InstallTypeWindowsClient, 0, errors.Wrapf(err, "failed to read version file %s", versionFilePath)
		}

		var versionData GameVersionFile
		if err := json.Unmarshal(versionFile, &versionData); err != nil {
			return InstallTypeWindowsClient, 0, errors.Wrapf(err, "failed to parse version file %s", versionFilePath)
		}

		return info.installType, versionData.Changelist, nil
	}
	return InstallTypeWindowsClient, 0, errors.New("failed to get game info")
}
