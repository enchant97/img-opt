package web

import (
	"net/http"
	"net/url"
	"path"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/enchant97/img-opt/config"
	"github.com/enchant97/img-opt/core"
	"github.com/labstack/echo/v4"
)

func getOriginalImage(ctx echo.Context) error {
	appConfig := ctx.Get(AppConfigKey).(config.Config)
	relativePath, err := url.QueryUnescape(ctx.Param("path"))
	if err != nil {
		return err
	}
	fullPath := path.Join(appConfig.OriginalsBase, relativePath)

	ctx.Response().Header().Add("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")

	if exists, err := core.DoesFileExist(fullPath); err != nil {
		ctx.Logger().Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	} else if !exists {
		return ctx.NoContent(http.StatusNotFound)
	}

	if currentETag, err := core.CreateETagFromFile(fullPath); err != nil {
		return err
	} else if needNewContent := HandleETag(ctx, currentETag); !needNewContent {
		return ctx.NoContent(http.StatusNotModified)
	}

	ctx.Response().Header().Add("Content-Optimized", "false")

	return ctx.File(fullPath)
}

func getAutoOptimized(ctx echo.Context) error {
	appConfig := ctx.Get(AppConfigKey).(config.Config)
	jobLimiter := ctx.Get(JobLimiterKey).(*core.JobLimiter)
	relativePath, err := url.QueryUnescape(ctx.Param("path"))
	if err != nil {
		return err
	}
	fullPath := path.Join(appConfig.OriginalsBase, relativePath)

	ctx.Response().Header().Add("Vary", "Accept")
	ctx.Response().Header().Add("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")

	if exists, err := core.DoesFileExist(fullPath); err != nil {
		ctx.Logger().Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	} else if !exists {
		return ctx.NoContent(http.StatusNotFound)
	}

	if currentETag, err := core.CreateETagFromFile(fullPath); err != nil {
		return err
	} else if needNewContent := HandleETag(ctx, currentETag); !needNewContent {
		return ctx.NoContent(http.StatusNotModified)
	}

	// skip any other unneeded processing
	if !appConfig.AutoOptimize.Enable {
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	originalImageFormat, err := core.DetermineImageType(fullPath)
	if err != nil {
		return err
	}

	imageFormat := originalImageFormat

	if originalImageFormat == vips.ImageTypeSVG {
		// TODO optimise svg content somehow?
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	acceptHeader := ctx.Request().Header.Get("Accept")
	nonStandardSupport := core.NonStandardFromAcceptHeader(acceptHeader)
	if nonStandardSupport.AVIF && appConfig.AutoOptimize.AVIF {
		imageFormat = vips.ImageTypeAVIF
	} else if nonStandardSupport.WEBP {
		imageFormat = vips.ImageTypeWEBP
	}

	optimiseJob := core.OptimiseJob{
		FullPath: fullPath,
		OutType:  imageFormat,
		MaxWidth: appConfig.AutoOptimize.MaxWidth,
		// other fields set below
	}

	switch imageFormat {
	case vips.ImageTypeJPEG:
	case vips.ImageTypeWEBP:
	case vips.ImageTypeAVIF:
		optimiseJob.Quality = 80
	default:
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	if err := jobLimiter.AddJob(); err != nil {
		ctx.Response().Header().Del("Cache-Control")
		ctx.Response().Header().Add("Retry-After", "5")
		return ctx.NoContent(http.StatusServiceUnavailable)
	} else {
		defer jobLimiter.RemoveJob()
	}

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	ctx.Response().Header().Add("Content-Optimized", "true")
	return ctx.Blob(http.StatusOK, core.ImageTypeToMime[imageFormat], img)
}

type ImageQuery struct {
	Type   string `query:"type"`
	Format string `query:"format"`
}

func getTypeOptimizedImage(ctx echo.Context) error {
	appConfig := ctx.Get(AppConfigKey).(config.Config)
	jobLimiter := ctx.Get(JobLimiterKey).(*core.JobLimiter)
	relativePath, err := url.QueryUnescape(ctx.Param("path"))
	if err != nil {
		return err
	}
	fullPath := path.Join(appConfig.OriginalsBase, relativePath)

	ctx.Response().Header().Add("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")

	var query ImageQuery
	if err := BindAndValidate(ctx, &query); err != nil {
		return err
	}

	if exists, err := core.DoesFileExist(fullPath); err != nil {
		ctx.Logger().Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	} else if !exists {
		return ctx.NoContent(http.StatusNotFound)
	}

	if currentETag, err := core.CreateETagFromFile(fullPath); err != nil {
		return err
	} else if needNewContent := HandleETag(ctx, currentETag); !needNewContent {
		return ctx.NoContent(http.StatusNotModified)
	}

	// just want the original
	if query.Type == "" && query.Format == "" {
		ctx.Response().Header().Add("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	var imageFormat vips.ImageType

	if query.Format == "" {
		var err error
		imageFormat, err = core.DetermineImageType(fullPath)
		if err != nil {
			return err
		}
		if imageFormat == vips.ImageTypeSVG {
			ctx.Response().Header().Add("Content-Optimized", "false")
			return ctx.File(fullPath)
		}
		if _, compatible := core.ImageTypeToFormatName[imageFormat]; !compatible {
			return ctx.JSON(http.StatusBadRequest, "unsupported format requested")
		}
	} else {
		var compatible bool
		if imageFormat, compatible = core.FormatNameToImageType[query.Format]; !compatible {
			return ctx.JSON(http.StatusBadRequest, "unsupported format requested")
		}
	}

	optimiseJob := core.OptimiseJob{
		FullPath: fullPath,
		OutType:  imageFormat,
		// other fields set below
	}

	if optConfig, exists := appConfig.TypeOptimize.Types[query.Type]; exists {
		if ifConfig, exists := optConfig.Formats[query.Format]; exists {
			optimiseJob.Quality = ifConfig.Quality
			optimiseJob.MaxWidth = &optConfig.MaxWidth
		} else {
			return ctx.JSON(http.StatusBadRequest, "unsupported type+format requested")
		}
	} else {
		return ctx.JSON(http.StatusBadRequest, "unsupported type requested")
	}

	if err := jobLimiter.AddJob(); err != nil {
		ctx.Response().Header().Del("Cache-Control")
		ctx.Response().Header().Add("Retry-After", "5")
		return ctx.NoContent(http.StatusServiceUnavailable)
	} else {
		defer jobLimiter.RemoveJob()
	}

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	ctx.Response().Header().Add("Content-Optimized", "true")
	return ctx.Blob(http.StatusOK, core.ImageTypeToMime[imageFormat], img)
}
