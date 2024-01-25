package steam

import (
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/andygrunwald/vdf"
	"github.com/pkg/errors"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
)

var manifests = []string{"appmanifest_526870.acf", "appmanifest_1690800.acf"}

func findInstallationsSteam(steamPath string, launcher string, executable []string) ([]*common.Installation, []error) {
	steamAppsPath := filepath.Join(steamPath, "steamapps")
	libraryFoldersManifestPath := filepath.Join(steamAppsPath, "libraryfolders.vdf")

	libraryFoldersF, err := os.Open(libraryFoldersManifestPath)
	if err != nil {
		return nil, []error{errors.Wrap(err, "failed to open library folders manifest")}
	}

	parser := vdf.NewParser(libraryFoldersF)
	libraryFoldersManifest, err := parser.Parse()
	if err != nil {
		return nil, []error{errors.Wrap(err, "failed to parse library folders manifest")}
	}

	var libraryFoldersList map[string]interface{}

	if _, ok := libraryFoldersManifest["LibraryFolders"]; ok {
		libraryFoldersList = libraryFoldersManifest["LibraryFolders"].(map[string]interface{})
	} else if _, ok := libraryFoldersManifest["libraryfolders"]; ok {
		libraryFoldersList = libraryFoldersManifest["libraryfolders"].(map[string]interface{})
	} else {
		return nil, []error{errors.New("failed to find library folders in manifest")}
	}

	libraryFolders := []string{
		path.Clean(steamPath),
	}

	for key, val := range libraryFoldersList {
		if _, err := strconv.Atoi(key); err != nil {
			continue
		}

		libraryFolderData := val.(map[string]interface{})
		libraryFolder := libraryFolderData["path"].(string)

		found := false
		for _, existingLibraryFolder := range libraryFolders {
			if common.OsPathEqual(existingLibraryFolder, libraryFolder) {
				found = true
				break
			}
		}
		if !found {
			libraryFolders = append(libraryFolders, libraryFolder)
		}
	}

	installs := make([]*common.Installation, 0)
	var findErrors []error

	for _, libraryFolder := range libraryFolders {
		for _, manifest := range manifests {
			manifestPath := filepath.Join(libraryFolder, "steamapps", manifest)

			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				continue
			}

			manifestF, err := os.Open(manifestPath)
			if err != nil {
				findErrors = append(findErrors, errors.Wrapf(err, "failed to open manifest file %s", manifestPath))
				continue
			}

			parser := vdf.NewParser(manifestF)
			manifest, err := parser.Parse()
			if err != nil {
				findErrors = append(findErrors, errors.Wrapf(err, "failed to parse manifest file %s", manifestPath))
				continue
			}

			if _, ok := manifest["AppState"]; !ok {
				findErrors = append(findErrors, errors.Errorf("Failed to find AppState in manifest %s", manifestPath))
				continue
			}

			fullInstallationPath := filepath.Join(libraryFolder, "steamapps", "common", manifest["AppState"].(map[string]interface{})["installdir"].(string))

			installType, version, err := common.GetGameInfo(fullInstallationPath)
			if err != nil {
				findErrors = append(findErrors, common.InstallFindError{
					Path:  fullInstallationPath,
					Inner: err,
				})
				continue
			}

			var branch common.GameBranch
			userConfig := manifest["AppState"].(map[string]interface{})["UserConfig"].(map[string]interface{})
			betakey, ok := userConfig["betakey"]
			if !ok {
				branch = common.BranchEarlyAccess
			} else {
				if betakey == "experimental" {
					branch = common.BranchExperimental
				} else {
					findErrors = append(findErrors, errors.Errorf("Unknown beta key %s", betakey))
				}
			}

			installs = append(installs, &common.Installation{
				Path:     filepath.Clean(fullInstallationPath),
				Version:  version,
				Type:     installType,
				Location: common.LocationTypeLocal,
				Branch:   branch,
				Launcher: launcher,
				LaunchPath: append(
					executable,
					`steam://rungameid/526870`,
				),
			})
		}
	}

	return installs, findErrors
}
