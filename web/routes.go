package web

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/enchant97/img-opt/config"
	"github.com/enchant97/img-opt/core"
	"github.com/h2non/bimg"
	"github.com/labstack/echo/v4"
)

func getAutoOptimized(ctx echo.Context) error {
	appConfig := ctx.Get(AppConfigKey).(config.Config)
	fullPath := path.Join(appConfig.OriginalsBase, ctx.Param("path"))

	ctx.Response().Header().Add("Vary", "Accept")
	ctx.Response().Header().Add("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")

	if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
		return ctx.NoContent(http.StatusNotFound)
	} else if err != nil {
		ctx.Logger().Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if currentETag, err := core.CreateETagFromFile(fullPath); err != nil {
		return err
	} else {
		ctx.Response().Header().Add("ETag", "\""+currentETag+"\"")
		if headerValue := ctx.Request().Header.Get("If-None-Match"); headerValue != "" {
			tags := strings.Split(headerValue, ",")
			for _, tag := range tags {
				tag = strings.Trim(strings.TrimSpace(tag), "\"")
				if tag == currentETag {
					return ctx.NoContent(http.StatusNotModified)
				}
			}
		}
	}

	// skip any other unneeded processing
	if !appConfig.AutoOptimize.Enable {
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	imageFormat := ""
	originalImageFormat, err := core.GetImageType(fullPath)
	if err != nil {
		return err
	}

	acceptHeader := ctx.Request().Header.Get("Accept")
	nonStandardSupport := core.NonStandardFromAcceptHeader(acceptHeader)
	if nonStandardSupport.AVIF {
		imageFormat = "avif"
	} else if nonStandardSupport.WEBP {
		imageFormat = "webp"
	}

	if imageFormat == "" || imageFormat == originalImageFormat {
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	optimiseJob := core.OptimiseJob{
		FullPath: fullPath,
		// other fields set below
	}

	switch imageFormat {
	case "avif":
		optimiseJob.OutType = bimg.AVIF
		optimiseJob.Quality = 60
	case "webp":
		optimiseJob.OutType = bimg.WEBP
		optimiseJob.Quality = 60
	default:
		return ctx.NoContent(http.StatusInternalServerError)
	}

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	ctx.Response().Header().Add("Content-Optimized", "true")
	return ctx.Blob(http.StatusOK, "image/"+imageFormat, img)

}

type ImageQuery struct {
	Type   string `query:"type"`
	Format string `query:"format"`
}

func getTypeOptimizedImage(ctx echo.Context) error {
	appConfig := ctx.Get(AppConfigKey).(config.Config)
	fullPath := path.Join(appConfig.OriginalsBase, ctx.Param("path"))

	ctx.Response().Header().Add("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")

	var query ImageQuery
	if err := BindAndValidate(ctx, &query); err != nil {
		return err
	}

	if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
		return ctx.NoContent(http.StatusNotFound)
	} else if err != nil {
		ctx.Logger().Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if currentETag, err := core.CreateETagFromFile(fullPath); err != nil {
		return err
	} else {
		ctx.Response().Header().Add("ETag", "\""+currentETag+"\"")
		if headerValue := ctx.Request().Header.Get("If-None-Match"); headerValue != "" {
			tags := strings.Split(headerValue, ",")
			for _, tag := range tags {
				tag = strings.Trim(strings.TrimSpace(tag), "\"")
				if tag == currentETag {
					return ctx.NoContent(http.StatusNotModified)
				}
			}
		}
	}

	// just want the original
	if query.Type == "" && query.Format == "" {
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	optimiseJob := core.OptimiseJob{
		FullPath: fullPath,
		// other fields set below
	}

	if query.Format == "" {
		var err error
		query.Format, err = core.GetImageType(fullPath)
		if err != nil {
			return err
		}
	}

	switch query.Format {
	case "jpeg":
		optimiseJob.OutType = bimg.JPEG
	case "webp":
		optimiseJob.OutType = bimg.WEBP
	case "avif":
		optimiseJob.OutType = bimg.AVIF
	}

	if optConfig, exists := appConfig.TypeOptimize.Types[query.Type]; exists {
		if ifConfig, exists := optConfig.Formats[query.Format]; exists {
			optimiseJob.Quality = ifConfig.Quality
		}
		optimiseJob.Width = &optConfig.Width
	} else {
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	ctx.Response().Header().Add("Content-Optimized", "true")
	return ctx.Blob(http.StatusOK, "image/"+query.Format, img)
}
