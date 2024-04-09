package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
	"github.com/tawesoft/golib/v2/dialog"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/satisfactorymodding/SatisfactoryModManager/backend"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/app"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/autoupdate"
	appCommon "github.com/satisfactorymodding/SatisfactoryModManager/backend/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/ficsitcli"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/installfinders/common"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/logging"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/settings"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/utils"
	"github.com/satisfactorymodding/SatisfactoryModManager/backend/websocket"

	"regexp"
	"strings"

	"github.com/brahma-adshonor/gohook"
	ficsitCliCache "github.com/satisfactorymodding/ficsit-cli/cli/cache"
	ficsitCliUtils "github.com/satisfactorymodding/ficsit-cli/utils"
)

//go:embed all:frontend/build
var assets embed.FS

var (
	version = "0.2.0-zh"
	commit  = "unknown"
	date    = "unknown"

	updateMode = "none"
)

func main() {
	logging.Init()

	autoupdate.Init()

	err := settings.LoadSettings()
	if err != nil {
		slog.Error("failed to load settings", slog.Any("error", err))
		// Cannot use wails message dialogs here yet, because they expect a frontend to exist
		_ = dialog.Error("Failed to load settings: %s", err.Error())
		os.Exit(1)
	}

	zhmod()

	if settings.Settings.CacheDir != "" {
		err = settings.ValidateCacheDir(settings.Settings.CacheDir)
		if err != nil {
			slog.Error("failed to set cache dir", slog.Any("error", err))
		} else {
			viper.Set("cache-dir", settings.Settings.CacheDir)
		}
	}

	err = ficsitcli.Init()
	if err != nil {
		slog.Error("failed to initialize ficsit-cli", slog.Any("error", err))
		_ = dialog.Error("Failed to initialize ficsit-cli: %s", err.Error())
		os.Exit(1)
	}

	windowStartState := options.Normal
	if settings.Settings.Maximized {
		windowStartState = options.Maximised
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:            "SatisfactoryModManager",
		Frameless:        runtime.GOOS == "windows",
		Width:            settings.Settings.UnexpandedSize.Width,
		Height:           settings.Settings.UnexpandedSize.Height,
		MinWidth:         utils.UnexpandedMin.Width,
		MaxWidth:         utils.UnexpandedMax.Width,
		MinHeight:        utils.UnexpandedMin.Height,
		MaxHeight:        utils.UnexpandedMax.Height,
		WindowStartState: windowStartState,
		AssetServer:      &assetserver.Options{Assets: assets},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "SatisfactoryModManager_b04ab4c3-450f-48f4-ab14-af6d7adc5416",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				app.App.Show()
				backend.ProcessArguments(secondInstanceData.Args)
			},
		},
		OnStartup: func(ctx context.Context) {
			appCommon.AppContext = ctx

			// Wails doesn't support setting the window position on init, so we do it here
			loadWindowLocation(ctx)

			app.App.WatchWindow() //nolint:contextcheck
			go websocket.ListenAndServeWebsocket()

			ficsitcli.FicsitCLI.StartGameRunningWatcher() //nolint:contextcheck
		},
		OnDomReady: func(ctx context.Context) {
			backend.ProcessArguments(os.Args[1:]) //nolint:contextcheck
			autoupdate.Updater.CheckInterval(5 * time.Minute)
		},
		OnShutdown: func(ctx context.Context) {
			app.App.StopWindowWatcher()
		},
		Bind: []interface{}{
			app.App,
			ficsitcli.FicsitCLI,
			autoupdate.Updater,
			settings.Settings,
		},
		EnumBind: []interface{}{
			common.AllInstallTypes,
			common.AllBranches,
			common.AllLocationTypes,
			ficsitcli.AllInstallationStates,
		},
		Logger: backend.WailsZeroLogLogger{},
	})

	if err != nil {
		slog.Error("failed to start application", slog.Any("error", err))
		_ = dialog.Error("Failed to start application: %s", err.Error())
	}

	err = autoupdate.Updater.OnExit()
	if err != nil {
		slog.Error("failed to apply update on exit", slog.Any("error", err))
		_ = dialog.Error("Failed to apply update on exit: %s", err.Error())
	}
}

func loadWindowLocation(ctx context.Context) {
	if settings.Settings.WindowPosition != nil {
		// Setting the window location is relative to the current monitor,
		// but we save it as absolute position.

		wailsRuntime.WindowSetPosition(ctx, 0, 0)

		// Get the location the window was actually placed at
		monitorLeft, monitorTop := wailsRuntime.WindowGetPosition(ctx)

		x := settings.Settings.WindowPosition.X - monitorLeft
		y := settings.Settings.WindowPosition.Y - monitorTop

		wailsRuntime.WindowSetPosition(ctx, x, y)
	}
}

func init() {
	// Pass build-time variables to viper
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}
	viper.Set("version", version)
	viper.Set("commit", commit)
	viper.Set("date", date)
	viper.Set("update-mode", updateMode)

	var baseLocalDir string

	switch runtime.GOOS {
	case "windows":
		baseLocalDir = os.Getenv("APPDATA")
	case "linux":
		baseLocalDir = filepath.Join(os.Getenv("HOME"), ".local", "share")
	default:
		slog.Error("unsupported OS", slog.String("os", runtime.GOOS))
		_ = dialog.Error("Unsupported OS: %s", runtime.GOOS)
		os.Exit(1)
	}

	viper.Set("base-local-dir", baseLocalDir)

	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	// ficsit-cli config

	viper.Set("profiles-file", "profiles.json")
	viper.Set("installations-file", "installations.json")
	viper.Set("api-base", "https://api.ficsit.app")
	viper.Set("graphql-api", "/v2/query")
	viper.Set("concurrent-downloads", 5)

	cacheDir := filepath.Clean(filepath.Join(baseCacheDir, "ficsit"))
	_ = utils.EnsureDirExists(cacheDir)
	viper.Set("cache-dir", cacheDir)

	localDir := filepath.Clean(filepath.Join(baseLocalDir, "ficsit"))
	_ = utils.EnsureDirExists(localDir)
	viper.Set("local-dir", localDir)

	// SMM config

	smmCacheDir := filepath.Clean(filepath.Join(baseCacheDir, "SatisfactoryModManager"))
	_ = utils.EnsureDirExists(smmCacheDir)
	viper.Set("smm-cache-dir", smmCacheDir)

	smmLocalDir := filepath.Clean(filepath.Join(baseLocalDir, "SatisfactoryModManager"))
	_ = utils.EnsureDirExists(smmLocalDir)
	viper.Set("smm-local-dir", smmLocalDir)

	viper.Set("default-cache-dir", cacheDir)

	viper.Set("websocket-port", 33642)

	viper.Set("github-release-repo", "satisfactorymodding/SatisfactoryModManager")

	// logging

	viper.Set("log-file", filepath.Join(smmCacheDir, "logs", "SatisfactoryModManager.log"))
}

func zhmod() {
	err := gohook.Hook(ficsitCliCache.DownloadOrCache, PatchDownloadOrCache, TrampDownloadOrCache)
	if err != nil {
		slog.Error("failed to patch", slog.Any("error", err))
	}

	if settings.Settings.Proxy != "" {
		proxy := settings.Settings.Proxy
		os.Unsetenv("no_proxy")
		os.Setenv("http_proxy", proxy)
		os.Setenv("https_proxy", proxy)
		slog.Info("set", slog.String("http_proxy", proxy))
	}

	slog.Info("zhmod done")
}

var reSml = regexp.MustCompile(`^https://github\.com/satisfactorymodding/SatisfactoryModLoader/releases/download\/(?P<version>v[\d\.]+)/SML(?P<arch>-\w+)?\.zip$`)

func PatchDownloadOrCache(cacheKey string, hash string, url string, updates chan<- ficsitCliUtils.GenericProgress, downloadSemaphore chan int) (*os.File, int64, error) {
	match := reSml.FindStringSubmatch(url)
	if match != nil {
		slog.Info("SML caught", slog.String("url", url))
		if settings.Settings.SmlLinkReplacer != "" {
			url = settings.Settings.SmlLinkReplacer // new URL
			for i, name := range reSml.SubexpNames() {
				if i > 0 && name != "" {
					var group string
					if i < len(match) {
						group = match[i]
					} else {
						group = "" // default empty string
					}
					// replace <version> <arch> ...
					url = strings.ReplaceAll(url, "<"+name+">", group)
				}
			}
			slog.Info("replaced SML", slog.String("url", url))
		}
	}
	return TrampDownloadOrCache(cacheKey, hash, url, updates, downloadSemaphore) // original call
}

func TrampDownloadOrCache(cacheKey string, hash string, url string, updates chan<- ficsitCliUtils.GenericProgress, downloadSemaphore chan int) (*os.File, int64, error) {
	slog.Info("DUMMY NOT CALLED0")
	slog.Info("DUMMY NOT CALLED1")
	slog.Info("DUMMY NOT CALLED2")
	slog.Info("DUMMY NOT CALLED3")
	slog.Info("DUMMY NOT CALLED4")
	return nil, 0, nil
}
