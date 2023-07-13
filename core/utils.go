package core

import "github.com/h2non/bimg"

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
