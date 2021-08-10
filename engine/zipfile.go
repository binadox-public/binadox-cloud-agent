package engine

import (
	"archive/zip"
	"crypto/ecdsa"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)



func ZipFiles(filename string, file string, comment string) error {

    newZipFile, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer newZipFile.Close()

    zipWriter := zip.NewWriter(newZipFile)
    defer zipWriter.Close()

    // Add files to zip

    if err = AddFileToZip(zipWriter, file, comment); err != nil {
    	return err
    }
    return nil

}

func AddFileToZip(zipWriter *zip.Writer, filename string, comment string) error {

    fileToZip, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer fileToZip.Close()

    // Get the file information
    info, err := fileToZip.Stat()
    if err != nil {
        return err
    }

    header, err := zip.FileInfoHeader(info)
    if err != nil {
        return err
    }

    // Using FileInfoHeader() above only uses the basename of the file. If we want
    // to preserve the folder structure we can overwrite this with the full path.
    header.Name = filename

    // Change to deflate to gain better compression
    // see http://golang.org/pkg/archive/zip/#pkg-constants
    header.Method = zip.Deflate
	header.Comment = comment

    writer, err := zipWriter.CreateHeader(header)
    if err != nil {
        return err
    }
    _, err = io.Copy(writer, fileToZip)
    return err
}

func CreateDistZip(inputFile string, outputFile string, privKeyFile string) error {

	var privKeyFileData []byte
	var err error

	privKeyFileData, err = ioutil.ReadFile(privKeyFile)
	if err != nil {
		return err
	}
	var privKey *ecdsa.PrivateKey
	privKey, err = DecodePrivateKey(string(privKeyFileData))
	if err != nil {
		return err
	}
	var fileData [] byte
	fileData, err = ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}
	var signature Signature
	signature, err = SignMessage(fileData, privKey)

	if err != nil {
		return err
	}
	var signStr string
	signStr, err = SerializeSignature(signature)
	if err != nil {
		return err
	}
	err = ZipFiles(outputFile, inputFile, signStr)
	if err != nil {
		return err
	}
	return nil
}

const (
	PUBKEY = "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEAy0TarLZWMH+eoHal0YppID3+hy1\nzrbhAu9rfwzaBeNvfXYI+ETVujpopwYFhTWi8ht/qRZj+X6tofAynTmLdA==\n-----END PUBLIC KEY-----"
)

type FileInfo struct {
	Name string
	Comment string
}


func ListFilesInZip(src string) ([]FileInfo, error) {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	result := []FileInfo{}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		var inf FileInfo
		inf.Name = file.Name
		inf.Comment = file.Comment
		result = append(result, inf)
	}

	return result, nil
}

func Unzip(archive, targetDir, prefix string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}


	for _, file := range reader.File {
		pathFile := filepath.Join(targetDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(pathFile, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()
		pathFile = filepath.Join(targetDir, prefix + file.Name)
		targetFile, err := os.OpenFile(pathFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
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

func VerifyZip(inFile string) error {
	var filesInfo []FileInfo
	var err error

	filesInfo, err = ListFilesInZip(inFile)
	if err != nil {
		return err
	}
	if len(filesInfo) != 1 {
		return errors.New("Single file is expected to be included in zip.")
	}


	dir, err := ioutil.TempDir(".", "binadox-temp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	err = Unzip(inFile, dir, "")
	if err != nil {
		return err
	}
	oName := path.Join(dir, filesInfo[0].Name)
	return VerifyFile(oName, filesInfo[0].Comment)
}

func VerifyFile(inFile string, signature string) error {
	if len(signature) == 0 {
		return errors.New("Invalid signature.")
	}
	var (
		sig Signature
		err error
		buff []byte
		key *ecdsa.PublicKey
	)

	sig, err = DeserializeSignature(signature)
	if err != nil {
		return err
	}

	key, err = DecodePublicKey(PUBKEY)
	if err != nil {
		return err
	}

	buff, err = ioutil.ReadFile(inFile)
	if err != nil {
		return err
	}
	ok := VerifyMessage(buff, key, sig)
	if !ok {
		return errors.New("Verification failed.")
	}
	return nil
}