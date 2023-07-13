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

	ctx.Response().Header().Add("Vary", "Accept")

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
	if (query.Format == nil || *query.Format == "auto") && appConfig.AutoOptimize {
		acceptHeader := ctx.Request().Header.Get("Accept")
		nonStandardSupport := core.NonStandardFromAcceptHeader(acceptHeader)
		if nonStandardSupport.AVIF {
			imageFormat = "avif"
		} else if nonStandardSupport.WEBP {
			imageFormat = "webp"
		}
	} else if query.Format != nil && *query.Format != "auto" {
		imageFormat = *query.Format
	}

	// process image type request
	if query.Type == nil {
		imageType = "original"
	} else {
		imageType = *query.Type
	}

	// just want the original
	if imageFormat == "original" && imageType == "original" {
		return ctx.File(fullPath)
	}

	optimiseJob := core.OptimiseJob{
		FullPath: fullPath,
		// other fields set below
	}

	switch imageFormat {
	case "original":
		var err error
		imageFormat, err = core.GetImageType(fullPath)
		if err != nil {
			return err
		}
	case "jpeg":
		optimiseJob.OutType = bimg.JPEG
	case "webp":
		optimiseJob.OutType = bimg.WEBP
	case "avif":
		optimiseJob.OutType = bimg.AVIF
	default:
		return ctx.JSON(http.StatusNotFound, "unknown image format requested")
	}

	if imageFormat == "original" {
		switch imageType {
		case "webp":
		case "avif":
			optimiseJob.Quality = 60
		default:
			optimiseJob.Quality = 80
		}
	} else if optConfig, exists := appConfig.Optimizations.Types[imageType]; exists {
		if ifConfig, exists := optConfig.Formats[imageFormat]; !exists || !ifConfig.Enable {
			return ctx.JSON(http.StatusNotFound, "unknown image format requested")
		} else {
			optimiseJob.Quality = ifConfig.Quality
			optimiseJob.Width = &optConfig.Width
		}
	} else {
		return ctx.JSON(http.StatusNotFound, "unknown image type requested")
	}

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	return ctx.Blob(http.StatusOK, "image/"+imageFormat, img)
}
