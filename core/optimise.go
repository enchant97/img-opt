package core

import "github.com/h2non/bimg"

type OptimiseJob struct {
	FullPath string
	OutType  bimg.ImageType
	Width    *int
	Quality  int
}

func (job *OptimiseJob) Optimise() ([]byte, error) {
	rawImage, err := bimg.Read(job.FullPath)
	if err != nil {
		return nil, err
	}
	loadedImage := bimg.NewImage(rawImage)
	if err != nil {
		return nil, err
	}
	options := bimg.Options{
		Type:          job.OutType,
		StripMetadata: true,
		Quality:       int(job.Quality),
	}
	if job.Width != nil {
		options.Width = *job.Width
	}
	rawImage, err = loadedImage.Process(options)
	return rawImage, err
}
