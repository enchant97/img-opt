package core

import (
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"os"

	"github.com/h2non/bimg"
)

func GetImageType(imagePath string) (string, error) {
	rawImage, err := bimg.Read(imagePath)
	if err != nil {
		return "", err
	}
	loadedImage := bimg.NewImage(rawImage)
	if err != nil {
		return "", err
	}
	return loadedImage.Type(), nil
}

func CreateETagFromFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := crc64.New(crc64.MakeTable(crc64.ISO))
	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}
		if _, err := h.Write(buf); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}

func DoesFileExist(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
