package engine

import (
	"archive/zip"
	"fmt"
	"github.com/phayes/permbits"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func listFilesInZip(src string) ([]string, error) {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		result = append(result, file.Name)
	}

	return result, nil
}

func unzip(archive, targetDir, prefix string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(targetDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()
		path = filepath.Join(targetDir, prefix + file.Name)
		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}


func findReleaseCandidate(myTag string) (string, string, error) {
	releases, err := ListReleases()
	if err != nil {
		return "", myTag, err
	}
	maxTag := ""
	maxRelease := ""
	for _, v := range (releases) {
		fmt.Println(v)
		if myTag < v.Tag {
			if (maxTag < v.Tag) && (len(v.Assets) == 1) {
				maxTag = v.Tag
				maxRelease = v.Assets[0].Url
			}
		}
	}

	if len(maxRelease) > 0 {
		return maxRelease, maxTag, nil
	}
	return "", myTag, nil
}

func FetchRelease(myTag string) (string, error) {
	var (
		err error
		downloadUrl string
		newTag string
		tmpFile* os.File
	)
	tmpFile, err = ioutil.TempFile(os.TempDir(), "prefix-")
    if err != nil {
        log.Printf("Cannot create temporary file %v", err)
        return "", err
    }
    defer os.Remove(tmpFile.Name())

	downloadUrl, newTag, err = findReleaseCandidate(myTag)
	if err != nil {
		log.Printf("Can not find release candidate %v", err.Error())
		return "", err
	}
	if len(downloadUrl) == 0 {
		return "", nil
	}
	err = DownloadFile(tmpFile.Name(), downloadUrl)
	if err != nil {
		log.Printf("Can not download release candidate from %s , %v", downloadUrl, err.Error())
		return "", nil
	}

	files, err := listFilesInZip(tmpFile.Name())
	if err != nil {
		log.Printf("Corrupted release candidate %s , %v", downloadUrl, err.Error())
		return "", err
	}
	if len(files) != 1 {
		log.Printf("Corrupted release candidate %s ", downloadUrl)
		return "", err
	}
	err = unzip(tmpFile.Name(), GetUpdaterDir(), newTag + "-")
	if err != nil {
		log.Printf("Failed to unzip release candidate %s : %v", downloadUrl, err.Error())
		return "", err
	}
	oName := path.Join(GetUpdaterDir(), newTag + "-" + files[0])
	permissions, err := permbits.Stat(oName)
	permissions.SetGroupExecute(true)
	permissions.SetOtherExecute(true)
	permissions.SetUserExecute(true)
	permbits.Chmod(oName, permissions)
	return oName, nil
}