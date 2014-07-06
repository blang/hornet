package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"github.com/blang/pushr"
	"github.com/blang/semver"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func AutoUpdate(currentVersion string, pushrHost string, pushrRelease string, pushrChannel string, pushrReadToken string) error {
	client := pushr.NewClient(pushrHost, pushrReadToken, "")
	v, versionStr, err := client.LatestVersion(pushrRelease, pushrChannel)
	if err != nil {
		log.Printf("Latest version error: %s", err)
		return err
	}
	if v.ContentType != "application/zip" {
		return fmt.Errorf("Content-Type %s not supported", v.ContentType)
	}
	newer, err := isNewerVersion(currentVersion, versionStr)
	if err != nil {
		return err
	}
	if !newer {
		log.Printf("Already on the newest version: %s\n", currentVersion)
		return nil
	}

	tmpFile, err := ioutil.TempFile("", "autoupdatear")
	if err != nil {
		return err
	}
	tmpFile.Close()

	defer os.Remove(tmpFile.Name())

	err = client.Download(pushrRelease, versionStr, tmpFile.Name())
	if err != nil {
		return err
	}

	exFile, err := extractZip(tmpFile.Name())
	if err != nil {
		log.Printf("Error while extracting: %s", err)
		return err
	}
	defer os.Remove(exFile.Name())

	err = swapExecutable(exFile.Name())
	if err != nil {
		return err
	}
	log.Printf("Successfully updated to: %s\n", versionStr)
	return nil
}

func isNewerVersion(currentversion string, rawVersion string) (bool, error) {
	curV, err := semver.New(currentversion)
	if err != nil {
		return false, err
	}
	rawVersion = strings.Trim(rawVersion, " \t")
	rawVersion = strings.TrimPrefix(rawVersion, "v")
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

func extractZip(zipfilename string) (*os.File, error) {
	fin, err := os.OpenFile(zipfilename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	finInfo, err := fin.Stat()
	if err != nil {
		return nil, err
	}
	//reset tmpFile
	fin.Seek(0, 0)
	zr, err := zip.NewReader(fin, finInfo.Size())
	if err != nil {
		log.Printf("Error on new zip reader: %s", err)
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
	w := bufio.NewWriter(tmpExtrFile)
	buf := make([]byte, 1024)
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
