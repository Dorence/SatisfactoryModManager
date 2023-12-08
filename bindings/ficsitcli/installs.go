package ficsitcli

import (
	"os/exec"
	"sort"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/cli"

	"github.com/satisfactorymodding/SatisfactoryModManager/installfinders"
	"github.com/satisfactorymodding/SatisfactoryModManager/settings"
)

func (f *FicsitCLI) initInstallations() error {
	installs, findErrors := installfinders.FindInstallations()

	f.installFindErrors = findErrors
	f.installations = []*InstallationInfo{}
	f.ficsitCli.Installations.Installations = []*cli.Installation{}

	fallbackProfile := "Default"
	if f.ficsitCli.Profiles.GetProfile(fallbackProfile) == nil {
		// Pick first profile found
		for name := range f.ficsitCli.Profiles.Profiles {
			fallbackProfile = name
			break
		}
	}

	for _, install := range installs {
		ficsitCliInstall, err := f.ficsitCli.Installations.AddInstallation(f.ficsitCli, install.Path, fallbackProfile)
		if err != nil {
			return errors.Wrap(err, "failed to add installation")
		}
		f.installations = append(f.installations, &InstallationInfo{
			Installation: ficsitCliInstall,
			Info:         install,
		})
	}

	sort.Slice(f.installations, func(i, j int) bool {
		if f.installations[i].Info.Launcher != f.installations[j].Info.Launcher {
			return f.installations[i].Info.Launcher < f.installations[j].Info.Launcher
		}
		return f.installations[i].Info.Branch < f.installations[j].Info.Branch
	})

	for _, install := range f.installations {
		if savedSelectedProfile, ok := settings.Settings.SelectedProfile[install.Info.Path]; ok {
			if f.ficsitCli.Profiles.GetProfile(savedSelectedProfile) == nil {
				log.Warn().Str("profile", savedSelectedProfile).Str("install", install.Info.Path).Msg("Saved profile not found")
				continue
			}
			err := install.Installation.SetProfile(f.ficsitCli, savedSelectedProfile)
			if err != nil {
				return errors.Wrap(err, "failed to set profile")
			}
		}
		if modsEnabled, ok := settings.Settings.ModsEnabled[install.Info.Path]; ok {
			install.Installation.Vanilla = !modsEnabled
		}
	}

	if len(f.installations) > 0 {
		f.selectedInstallation = f.installations[0]
	}

	savedSelectedInstall := settings.Settings.SelectedInstall
	if savedSelectedInstall != "" {
		for _, install := range f.installations {
			if install.Info.Path == savedSelectedInstall {
				f.selectedInstallation = install
				break
			}
		}
	}

	return nil
}

func (f *FicsitCLI) GetInstallations() []*InstallationInfo {
	return f.installations
}

func (f *FicsitCLI) GetInstallationsInfo() []*installfinders.Installation {
	result := []*installfinders.Installation{}
	for _, install := range f.installations {
		result = append(result, install.Info)
	}
	return result
}

func (f *FicsitCLI) GetInvalidInstalls() []string {
	result := []string{}
	for _, err := range f.installFindErrors {
		var installFindErr installfinders.InstallFindError
		if ok := errors.As(err, &installFindErr); ok {
			result = append(result, installFindErr.Path)
		}
	}
	return result
}

func (f *FicsitCLI) GetInstallation(path string) *InstallationInfo {
	for _, install := range f.installations {
		if install.Info.Path == path {
			return install
		}
	}

	return nil
}

func (f *FicsitCLI) SelectInstall(path string) error {
	l := log.With().Str("task", "selectInstall").Str("path", path).Logger()
	if f.selectedInstallation != nil && f.selectedInstallation.Info.Path == path {
		return nil
	}
	installation := f.GetInstallation(path)
	if installation == nil {
		l.Error().Str("path", path).Msg("Failed to find installation")
		return errors.New("Installation \"" + path + "\" not found")
	}
	f.selectedInstallation = installation

	f.EmitGlobals()

	settings.Settings.SelectedInstall = installation.Info.Path
	_ = settings.SaveSettings()

	f.progress = &Progress{
		Item:     "__select_install__",
		Message:  "Validating install",
		Progress: -1,
	}

	f.setProgress(f.progress)

	defer f.setProgress(nil)

	installErr := f.validateInstall(installation, "__select_install__")

	if installErr != nil {
		l.Error().Err(installErr).Str("install", installation.Info.Path).Msg("Failed to validate install")
		return errors.Wrap(installErr, "Failed to validate install")
	}
	return nil
}

func (f *FicsitCLI) GetSelectedInstall() *installfinders.Installation {
	if f.selectedInstallation == nil {
		return nil
	}
	return f.selectedInstallation.Info
}

func (f *FicsitCLI) SetModsEnabled(enabled bool) error {
	l := log.With().Str("task", "setModsEnabled").Bool("enabled", enabled).Logger()

	var message string
	if enabled {
		message = "Enabling mods"
	} else {
		message = "Disabling mods"
	}

	f.selectedInstallation.Installation.Vanilla = !enabled

	f.EmitGlobals()

	f.progress = &Progress{
		Item:     "__toggle_mods__",
		Message:  message,
		Progress: -1,
	}

	f.setProgress(f.progress)

	defer f.setProgress(nil)

	installErr := f.validateInstall(f.selectedInstallation, "__toggle_mods__")

	if installErr != nil {
		l.Error().Err(installErr).Str("install", f.selectedInstallation.Info.Path).Msg("Failed to validate install")
		return errors.Wrap(installErr, "Failed to validate install")
	}

	settings.Settings.ModsEnabled[f.selectedInstallation.Info.Path] = enabled
	_ = settings.SaveSettings()
	return nil
}

func (f *FicsitCLI) GetModsEnabled() bool {
	return !f.selectedInstallation.Installation.Vanilla
}

func (f *FicsitCLI) GetSelectedInstallProfileMods() map[string]cli.ProfileMod {
	profile := f.GetProfile(f.selectedInstallation.Installation.Profile)
	return profile.Mods
}

func (f *FicsitCLI) GetSelectedInstallLockfileMods() (map[string]cli.LockedMod, error) {
	lockfile, err := f.selectedInstallation.Installation.LockFile(f.ficsitCli)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get lockfile")
	}
	if lockfile == nil {
		return make(map[string]cli.LockedMod), nil
	}
	return lockfile.Mods, nil
}

func (f *FicsitCLI) GetSelectedInstallLockfile() (*cli.LockFile, error) {
	lockfile, err := f.selectedInstallation.Installation.LockFile(f.ficsitCli)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get lockfile")
	}
	return lockfile, nil
}

func (f *FicsitCLI) LaunchGame() {
	if f.selectedInstallation == nil {
		log.Error().Msg("No installation selected")
		return
	}
	cmd := exec.Command(f.selectedInstallation.Info.LaunchPath[0], f.selectedInstallation.Info.LaunchPath[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("cmd", cmd.String()).Str("output", string(out)).Msg("Failed to launch game")
		return
	}
}
