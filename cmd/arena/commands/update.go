package commands

import (
	"encoding/json"
	"fmt"
	"github.com/mholt/archiver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

var (
	url    = "https://api.github.com/repos/run-ai/arena/releases/latest"
	osName = runtime.GOOS
	arch   = runtime.GOARCH
)

type GithubResponse struct {
	AssetsUrl string  `json:"assets_url"`
	Assets    []Asset `json:"assets"`
}

type Asset struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
}

func NewUpdateCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "update",
		Short: "update cli to latest version.",
		Run: func(cmd *cobra.Command, args []string) {
			if os.Getuid() != 0 {
				log.Error("The command must be run as root")
				os.Exit(1)
			}

			latestRelease := new(GithubResponse)

			// Find latest release from github
			err := getResponseFromGithub(url, &latestRelease)

			if err != nil {
				log.Error(err)
				os.Exit(1)
			}

			var matchingAsset Asset
			// Find matching asset for current OS and ARCH
			for _, asset := range latestRelease.Assets {
				if strings.Contains(asset.Name, osName) && strings.Contains(asset.Name, arch) {
					log.Infof("Found matching asset %s", asset.Name)
					matchingAsset = asset
					break
				}
			}

			if matchingAsset.DownloadUrl == "" {
				log.Errorf("Could not find matching asset for %s-%s", osName, arch)
				os.Exit(1)
			}

			// Download the asset to temp folder
			downloadPath, err := downloadFile(matchingAsset.DownloadUrl, matchingAsset.Name)

			if err != nil {
				log.Errorf("Could not download archive file %s", err)
				os.Exit(1)
			}

			// Unarchive the asset to temp folder
			unarchivePath := path.Join(os.TempDir(), fmt.Sprintf("%s-%s", osName, arch))

			tarArchiver := archiver.Tar{
				OverwriteExisting: true,
				MkdirAll:          true,
			}

			targzArchiver := archiver.TarGz{
				Tar: &tarArchiver,
			}

			err = targzArchiver.Unarchive(downloadPath, unarchivePath)
			if err != nil {
				log.Errorf("Error unarchiving downloaded file %s", err)
				os.Exit(1)
			}

			log.Infof("Unarchived version in %s", unarchivePath)

			// Install using install script
			installScriptPath := path.Join(unarchivePath, "install-runai.sh")
			installCommand := exec.Command(installScriptPath)
			installCommand.Stdout = os.Stdout
			installCommand.Stderr = os.Stderr
			err = installCommand.Run()

			if err != nil {
				log.Errorf("Error executing install script %s", err)
				os.Exit(1)
			}

			log.Infof("Successfully installed new version")
		},
	}

	return command
}

func getResponseFromGithub(url string, output interface{}) error {
	res, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("Could not access github api: %s", err)
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(output)

	if err != nil {
		return fmt.Errorf("Could not read body of github response: %s", err)
	}

	return nil
}

func downloadFile(url string, assetName string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Create the file
	downloadPath := path.Join(os.TempDir(), assetName)

	out, err := os.Create(downloadPath)
	if err != nil {
		return "", err
	}

	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)

	log.Infof("Downloaded arena archive to %s", downloadPath)

	return downloadPath, nil
}
