package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type GithubRelease struct {
	ID          int                  `json:"id"`
	Tag         string               `json:"tag_name"`
	Draft       bool                 `json:"draft"`
	PublishedAt string               `json:"published_at"`
	Assets      []GithubReleaseAsset `json:"assets"`
}

type GithubReleaseAsset struct {
	Url         string `json:"url"`
	ContentType string `json:"content_type"`
}

const GH_RELEASE_URL = "https://api.github.com/repos/blang/gosqm-slotlist/releases"

func AutoUpdate(version string) error {
	resp, err := http.Get(GH_RELEASE_URL)
	if err != nil {
		return err
	}

	var releases []GithubRelease
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return err
	}

	if len(releases) == 0 {
		return errors.New("No releases found")
	}

	recent := releases[0]
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
	log.Printf("File written: %s", file.Name())
	return nil
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

	tmpFile, err := ioutil.TempFile("", "hornet")
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

	tmpExtrFile, err := ioutil.TempFile("", "hornet")
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
