package web

import (
	"errors"
	"net/http"
	"os"
	"path"

	"github.com/enchant97/img-opt/config"
	"github.com/enchant97/img-opt/core"
	"github.com/h2non/bimg"
	"github.com/labstack/echo/v4"
)

type ImageQuery struct {
	Type   *string `query:"type"`
	Format *string `query:"format"`
}

func getOptimizedImage(ctx echo.Context) error {
	appConfig := ctx.Get(AppConfigKey).(config.Config)

	var query ImageQuery
	if err := BindAndValidate(ctx, &query); err != nil {
		return err
	}

	fullPath := path.Join(appConfig.OriginalsBase, ctx.Param("path"))

	if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
		return ctx.NoContent(http.StatusNotFound)
	} else if err != nil {
		ctx.Logger().Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	imageFormat := "original"
	imageType := "original"

	// process image format request
	if query.Format == nil || *query.Format == "auto" {
		acceptHeader := ctx.Request().Header.Get("Accept")
		nonStandardSupport := core.NonStandardFromAcceptHeader(acceptHeader)
		if nonStandardSupport.AVIF {
			imageFormat = "avif"
		} else if nonStandardSupport.WEBP {
			imageFormat = "webp"
		}
	} else {
		imageFormat = *query.Format
	}

	// process image type request
	if query.Type == nil {
		imageType = "original"
	} else {
		imageType = *query.Type
	}

	ctx.Response().Header().Add("Vary", "Accept")

	// just want the original
	if imageFormat == "original" && imageType == "original" {
		return ctx.File(fullPath)
	}

	optimiseJob := core.OptimiseJob{
		FullPath: fullPath,
		Quality:  80, // XXX remove this (defined with image types)
	}

	if imageFormat == "original" {
		var err error
		imageFormat, err = core.GetImageType(fullPath)
		if err != nil {
			return err
		}
	} else if imageFormat == "jpeg" {
		optimiseJob.OutType = bimg.JPEG
	} else if imageFormat == "webp" {
		optimiseJob.OutType = bimg.WEBP
	} else if imageFormat == "avif" {
		optimiseJob.OutType = bimg.AVIF
	} else {
		return ctx.NoContent(http.StatusBadRequest)
	}

	// TODO handle image types here

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	return ctx.Blob(http.StatusOK, "image/"+imageFormat, img)
}
