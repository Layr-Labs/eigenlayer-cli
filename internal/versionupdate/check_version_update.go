package versionupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fatih/color"

	"github.com/blang/semver/v4"
)

const (
	organization = "Layr-Labs"
	repository   = "eigenlayer-cli"
)

type release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// Check if there is a new version of the package
// If there is, print a message to the user
// If there isn't, do nothing
// Don't do anything for development version
// If anything fails in this, it will silently pass since this doesn't affect operations
func Check(currentVersion string) {
	if currentVersion == "development" {
		return
	}

	// Get latest version from GitHub releases
	latestReleaseUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", organization, repository)
	response, err := http.Get(latestReleaseUrl)
	if err != nil {
		return
	}
	defer response.Body.Close()
	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	var data release

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return
	}

	// GitHub API returns in vX.X.X format so remove v
	latestVersion := data.TagName[1:]
	latestSemVer, err := semver.Make(latestVersion)
	if err != nil {
		return
	}

	currentSemVer, err := semver.Make(currentVersion)
	if err != nil {
		return
	}

	if latestSemVer.GT(currentSemVer) {
		greenVersion := color.GreenString(latestVersion)
		yellowOldVersion := color.YellowString(currentVersion)
		fmt.Println()
		fmt.Printf("There is a new version (%s) for this library available.\n", greenVersion)
		fmt.Printf("Your current running verison is (%s).\n", yellowOldVersion)
		fmt.Println("Please update (https://github.com/Layr-Labs/eigenlayer-cli#install-eigenlayer-cli-using-a-binary) to get latest features and bug fixes.")
		fmt.Println()
	}
}
