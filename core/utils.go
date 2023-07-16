package core

import (
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
)

var UnknownImageTypeErr = errors.New("unknown image type")

var ImageTypeToFormatName = map[vips.ImageType]string{
	vips.ImageTypeJPEG: "jpeg",
	vips.ImageTypePNG:  "png",
	vips.ImageTypeWEBP: "webp",
	vips.ImageTypeAVIF: "avif",
}

var FormatNameToImageType = map[string]vips.ImageType{
	"jpeg": vips.ImageTypeJPEG,
	"png":  vips.ImageTypePNG,
	"webp": vips.ImageTypeWEBP,
	"avif": vips.ImageTypeAVIF,
}

var ImageTypeToMime = map[vips.ImageType]string{
	vips.ImageTypeJPEG: "image/jpeg",
	vips.ImageTypePNG:  "image/png",
	vips.ImageTypeWEBP: "image/webp",
	vips.ImageTypeAVIF: "image/avif",
}

func ReadWholeFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}

func DetermineImageType(imagePath string) (vips.ImageType, error) {
	buf, err := ReadWholeFile(imagePath)
	if err != nil {
		return 0, err
	}
	imageType := vips.DetermineImageType(buf)
	if imageType == vips.ImageTypeUnknown {
		return imageType, UnknownImageTypeErr
	}
	return imageType, nil
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
