package engine

import (
	"github.com/phayes/permbits"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func findReleaseCandidate(myTag string) (string, string, error) {
	releases, err := ListReleases()
	if err != nil {
		return "", myTag, err
	}
	maxTag := ""
	maxRelease := ""
	for _, v := range releases {
		if myTag < v.Tag {
			if maxTag < v.Tag  {
				maxTag = v.Tag
				for _, link := range v.Assets {
					if link.Name == FILE_TO_DOWNLOAD {
						maxRelease = v.Assets[0].Url
						break
					}
				}
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

	files, err := ListFilesInZip(tmpFile.Name())
	if err != nil {
		log.Printf("Corrupted release candidate %s , %v", downloadUrl, err.Error())
		return "", err
	}
	if len(files) != 1 {
		log.Printf("Corrupted release candidate %s ", downloadUrl)
		return "", err
	}
	err = Unzip(tmpFile.Name(), GetUpdaterDir(), newTag + "-")
	if err != nil {
		log.Printf("Failed to unzip release candidate %s : %v", downloadUrl, err.Error())
		return "", err
	}
	oName := path.Join(GetUpdaterDir(), newTag + "-" + files[0].Name)
	err = VerifyFile(oName, files[0].Comment)
	if err != nil {
		return "", err
	}
	permissions, err := permbits.Stat(oName)
	permissions.SetGroupExecute(true)
	permissions.SetOtherExecute(true)
	permissions.SetUserExecute(true)
	permbits.Chmod(oName, permissions)
	return oName, nil
}