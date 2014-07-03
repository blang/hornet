package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blang/semver"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const gh_release_url = "https://api.github.com/repos/%s/releases"

type githubRelease struct {
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
	Tag         string               `json:"tag_name"`
	Draft       bool                 `json:"draft"`
	PublishedAt string               `json:"published_at"`
	Assets      []githubReleaseAsset `json:"assets"`
}

type githubReleaseAsset struct {
	Url         string `json:"url"`
	ContentType string `json:"content_type"`
}

func AutoUpdate(version string, repository string) error {
	ghReleaseURL := fmt.Sprintf(gh_release_url, repository)
	resp, err := http.Get(ghReleaseURL)
	if err != nil {
		return err
	}

	var releases []githubRelease
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return err
	}

	if len(releases) == 0 {
		return errors.New("No releases found")
	}

	recent := releases[0]
	if newer, err := isNewerVersion(version, recent.Name); err != nil {
		return err
	} else if !newer {
		return nil
	}

	if len(recent.Assets) == 0 {
		return errors.New("Most recent release has no assets")
	}

	recentAsset := recent.Assets[0]
	if recentAsset.ContentType != "application/zip" {
		return errors.New("Most recent release asset is not a zip file")
	}

	file, err := downloadToTmp(recentAsset.Url)
	if err != nil {
		return err
	}
	err = swapExecutable(file.Name())
	if err != nil {
		return err
	}

	return nil
}

func isNewerVersion(currentversion string, rawVersion string) (bool, error) {
	curV, err := semver.New(currentversion)
	if err != nil {
		return false, err
	}
	rawVersion = strings.Trim(rawVersion, " \t")
	rawVersion = strings.TrimPrefix(rawVersion, "v")
	rawVersion = (strings.SplitN(rawVersion, " ", 2))[0]
	newV, err := semver.New(rawVersion)
	if err != nil {
		return false, err
	}

	// Don't allow prereleases
	if len(newV.Pre) > 0 {
		return false, nil
	}
	return newV.GT(curV), nil
}

func swapExecutable(newFileName string) error {
	const oldSuffix = ".old"
	exePath := os.Args[0]
	oldExePath := exePath + oldSuffix
	if _, err := os.Stat(oldExePath); err == nil {
		os.Remove(oldExePath)
	}
	err := os.Rename(exePath, oldExePath)
	if err != nil {
		return err
	}
	err = copyFile(newFileName, exePath)
	if err != nil {
		//Try to restore
		os.Rename(oldExePath, exePath)
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func downloadToTmp(assetUrl string) (*os.File, error) {
	req, err := http.NewRequest("GET", assetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	client := &http.Client{}
	binresp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(binresp.Body)

	tmpFile, err := ioutil.TempFile("", "autoupdatear")
	if err != nil {
		return nil, err
	}
	// TODO: Don't forget to remove the tmp file
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	w := bufio.NewWriter(tmpFile)
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err // TODO: Log this error?
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			return nil, err // TODO: Log this error?
		}
	}

	if err = w.Flush(); err != nil {
		return nil, err // TODO: Log this error?
	}

	//reset tmpFile
	tmpFile.Seek(0, 0)

	zr, err := zip.NewReader(tmpFile, binresp.ContentLength)
	if err != nil {
		return nil, err
	}
	var zFile *zip.File
	for _, file := range zr.File {
		ext := filepath.Ext(file.Name)
		if strings.ToLower(ext) == ".exe" {
			zFile = file
			break
		}
	}

	if zFile == nil {
		return nil, errors.New("No .exe file found in asset zip")
	}
	zFileHandle, err := zFile.Open()
	if err != nil {
		return nil, err
	}
	zFileReader := bufio.NewReader(zFileHandle)

	tmpExtrFile, err := ioutil.TempFile("", "autoupdate")
	if err != nil {
		return nil, err
	}
	// TODO: Don't forget to remove the tmp file
	defer tmpExtrFile.Close()
	w = bufio.NewWriter(tmpExtrFile)
	buf = make([]byte, 1024)
	for {
		// read a chunk
		n, err := zFileReader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err // TODO: Log this error?
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			return nil, err // TODO: Log this error?
		}
	}

	if err = w.Flush(); err != nil {
		return nil, err // TODO: Log this error?
	}
	return tmpExtrFile, nil
}
