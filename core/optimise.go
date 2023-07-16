package core

import (
	"errors"

	"github.com/davidbyttow/govips/v2/vips"
)

var UnsupportedImageTypeErr = errors.New("unsupported image type")

type OptimiseJob struct {
	FullPath string
	OutType  vips.ImageType
	MaxWidth *int
	Quality  int
}

func (job *OptimiseJob) Optimise() ([]byte, error) {
	img, err := vips.NewImageFromFile(job.FullPath)
	if err != nil {
		return nil, err
	}
	if err := img.AutoRotate(); err != nil {
		return nil, err
	}
	if job.MaxWidth != nil && img.Width() > *job.MaxWidth {
		img.Thumbnail(*job.MaxWidth, *job.MaxWidth, vips.InterestingNone)
	}
	var imgBytes []byte

	switch job.OutType {
	case vips.ImageTypeJPEG:
		p := vips.NewJpegExportParams()
		p.StripMetadata = true
		p.Quality = job.Quality
		imgBytes, _, err = img.ExportJpeg(p)
	case vips.ImageTypePNG:
		p := vips.NewPngExportParams()
		p.StripMetadata = true
		p.Quality = job.Quality
		imgBytes, _, err = img.ExportPng(p)
	case vips.ImageTypeWEBP:
		p := vips.NewWebpExportParams()
		p.StripMetadata = true
		p.Quality = job.Quality
		imgBytes, _, err = img.ExportWebp(p)
	case vips.ImageTypeAVIF:
		p := vips.NewAvifExportParams()
		p.StripMetadata = true
		p.Quality = job.Quality
		imgBytes, _, err = img.ExportAvif(p)
	default:
		return imgBytes, UnsupportedImageTypeErr
	}
	return imgBytes, err
}
