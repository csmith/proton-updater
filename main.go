package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/csmith/proton-updater/steamclient"
	"github.com/google/go-github/v28/github"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	release, url, err := getProtonVersion()
	if err != nil {
		log.Fatalf("Unable to find latest release: %s", err.Error())
		return
	}

	targetDir := fmt.Sprintf("Proton-%s", *release)

	if steamclient.HasCompatibilityTool(targetDir) {
		log.Printf("Steam already has compatibility tool '%s' available", targetDir)
		return
	}

	if steamclient.IsRunning() {
		log.Printf("Stopping steam")
		steamclient.Shutdown()
	}

	downloadAndInstall(url)

	log.Printf("Finished extracting")
}

func downloadAndInstall(url *string) {
	res, err := http.Get(*url)
	if err != nil {
		log.Fatalf("Unable to download %s: %s", *url, err.Error())
		return
	}

	defer res.Body.Close()

	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		log.Fatalf("Unable to create gzip reader: %s", err.Error())
	}

	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		th, err := tr.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Unable to read tar stream: %s", err.Error())
		}

		if th.Typeflag == tar.TypeSymlink {
			createSymbolicLink(th)
		} else if th.Typeflag != tar.TypeDir {
			createRegularFile(th, tr)
		}
	}
}

func createRegularFile(th *tar.Header, tr *tar.Reader) {
	log.Printf("Extracting %s", th.Name)
	target, err := steamclient.CreateCompatibilityToolFile(th.Name, th.Mode)
	if err != nil {
		log.Fatalf("Unable to create file %s: %s", th.Name, err.Error())
	}

	defer target.Close()

	_, err = io.Copy(target, tr)
	if err != nil {
		log.Fatalf("Unable to write file %s: %s", th.Name, err.Error())
	}
}

func createSymbolicLink(th *tar.Header) {
	log.Printf("Creating link from %s to %s", th.Name, th.Linkname)
	base := steamclient.CompatibilityToolPath()
	p := filepath.Join(base, th.Name)

	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		log.Fatalf("Unable to create directory %s: %s", filepath.Dir(p), err.Error())
	}

	if err := os.Symlink(th.Linkname, p); err != nil {
		log.Fatalf("Unable to create link: %s", err.Error())
	}
}

func getProtonVersion() (*string, *string, error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "GloriousEggroll", "proton-ge-custom")
	if err != nil {
		return nil, nil, err
	}

	log.Printf("Found latest release of tag %s", release.GetTagName())

	expectedName := fmt.Sprintf("Proton-%s.tar.gz", release.GetTagName())
	for _, asset := range release.Assets {
		log.Printf("Found asset named '%s'", asset.GetName())
		if asset.GetName() == expectedName {
			log.Printf("Selected asset. Download URL: %s", asset.GetBrowserDownloadURL())
			return release.TagName, asset.BrowserDownloadURL, nil
		}
	}

	return nil, nil, fmt.Errorf("no matching download for release %s", release.GetTagName())
}
