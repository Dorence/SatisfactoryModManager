package github

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

type Provider struct {
	repo             string
	checksumArtifact string
}

func MakeGithubProvider(repo string, checksumArtifact string) *Provider {
	return &Provider{
		repo:             repo,
		checksumArtifact: checksumArtifact,
	}
}

func (g *Provider) GetLatestVersion(includePrerelease bool) (string, error) {
	if !includePrerelease {
		release, err := g.getLatestReleaseData()
		if err != nil {
			return "", errors.Wrap(err, "failed to get latest release")
		}
		return release.TagName, nil
	}

	// GitHub does not return pre-releases on the /latest endpoint
	allReleases, err := g.getReleasesData()
	var latest *semver.Version
	var latestTagName string
	if err != nil {
		return "", errors.Wrap(err, "failed to get releases")
	}
	for _, release := range allReleases {
		version, err := semver.NewVersion(release.TagName)
		if err != nil {
			continue
		}
		if !includePrerelease && version.Prerelease() != "" {
			continue
		}
		if latest == nil || version.GreaterThan(latest) {
			latest = version
			latestTagName = release.TagName
		}
	}
	if latest == nil {
		return "", errors.New("no releases found")
	}
	return latestTagName, nil
}

func (g *Provider) GetFile(version string, filename string) (io.ReadCloser, int64, []byte, error) {
	release, err := g.getReleaseData(version)
	if err != nil {
		return nil, 0, nil, errors.Wrap(err, "failed to get latest release")
	}
	fileURL := getAssetURL(release, filename)
	if fileURL == "" {
		return nil, 0, nil, errors.Errorf("failed to find asset")
	}
	checksum, err := g.getFileChecksum(release, filename)
	if err != nil {
		return nil, 0, nil, errors.Wrap(err, "failed to get checksum")
	}
	response, err := http.Get(fileURL)
	if err != nil {
		return nil, 0, nil, errors.Wrapf(err, "failed to download asset")
	}
	return response.Body, response.ContentLength, checksum, nil
}

func getAssetURL(release *Release, assetName string) string {
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

func (g *Provider) GetChangelogs() (map[string]string, error) {
	releases, err := g.getReleasesData()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest release")
	}
	changelogs := make(map[string]string)
	for _, release := range releases {
		changelogs[release.TagName] = release.Body
	}
	return changelogs, nil
}

func (g *Provider) getLatestReleaseData() (*Release, error) {
	response, err := http.Get("https://api.github.com/repos/" + g.repo + "/releases/latest")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest release")
	}
	defer response.Body.Close()
	var release Release
	err = json.NewDecoder(response.Body).Decode(&release)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode latest release")
	}
	return &release, nil
}

func (g *Provider) getReleasesData() ([]Release, error) {
	response, err := http.Get("https://api.github.com/repos/" + g.repo + "/releases")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get releases")
	}
	defer response.Body.Close()
	var releases []Release
	err = json.NewDecoder(response.Body).Decode(&releases)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode releases")
	}
	return releases, nil
}

func (g *Provider) getReleaseData(tagName string) (*Release, error) {
	response, err := http.Get("https://api.github.com/repos/" + g.repo + "/releases/tags/" + tagName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get releases")
	}
	defer response.Body.Close()
	var release Release
	err = json.NewDecoder(response.Body).Decode(&release)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode releases")
	}
	return &release, nil
}

func (g *Provider) getFileChecksum(release *Release, filename string) ([]byte, error) {
	if g.checksumArtifact == "" {
		return nil, nil
	}
	url := getAssetURL(release, g.checksumArtifact)
	if url == "" {
		return nil, errors.Errorf("failed to find checksum asset")
	}
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download checksum asset")
	}
	defer response.Body.Close()
	checksum, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read checksum")
	}
	checksums := parseChecksumFile(checksum)
	if sum, ok := checksums[filename]; ok {
		return sum, nil
	}
	return nil, errors.Errorf("failed to find checksum for file")
}

func parseChecksumFile(checksumFile []byte) map[string][]byte {
	checksums := make(map[string][]byte)
	lines := strings.Split(string(checksumFile), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			// Skip empty lines
			continue
		}
		parts := strings.Split(line, "  ")
		if len(parts) != 2 {
			slog.Debug("invalid checksum entry", slog.String("entry", line))
			continue
		}
		hexSum := parts[0]
		filename := parts[1]
		sum, err := hex.DecodeString(hexSum)
		if err != nil {
			slog.Debug("failed to decode checksum", slog.String("checksum", hexSum), slog.String("filename", filename), slog.Any("error", err))
			continue
		}
		checksums[parts[1]] = sum
	}
	return checksums
}
