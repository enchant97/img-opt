package web

import (
	"fmt"
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

	SetCacheHeader(ctx, appConfig)

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

	ctx.Response().Header().Set("Content-Optimized", "false")

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

	ctx.Response().Header().Set("Vary", "Accept")
	SetCacheHeader(ctx, appConfig)

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
		ctx.Response().Header().Set("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	originalImageFormat, err := core.DetermineImageType(fullPath)
	if err != nil {
		return err
	}

	if originalImageFormat == vips.ImageTypeSVG {
		// TODO optimise svg content somehow?
		ctx.Response().Header().Set("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	// check if browser supports 'fancy' formats
	acceptHeader := ctx.Request().Header.Get("Accept")
	nonStandardSupport := core.NonStandardFromAcceptHeader(acceptHeader)

	optimiseJob := core.OptimiseJob{
		FullPath: fullPath,
		MaxWidth: appConfig.AutoOptimize.MaxWidth,
		// other fields set below
	}

	// determine suitable format
	avifConfig, avifEnabled := appConfig.AutoOptimize.Formats[core.ImageTypeToFormatName[vips.ImageTypeAVIF]]
	webpConfig, webpEnabled := appConfig.AutoOptimize.Formats[core.ImageTypeToFormatName[vips.ImageTypeWEBP]]
	pngConfig, pngEnabled := appConfig.AutoOptimize.Formats[core.ImageTypeToFormatName[vips.ImageTypePNG]]
	jpegConfig, jpegEnabled := appConfig.AutoOptimize.Formats[core.ImageTypeToFormatName[vips.ImageTypeJPEG]]
	if nonStandardSupport.AVIF && avifEnabled {
		optimiseJob.OutType = vips.ImageTypeAVIF
		optimiseJob.Quality = avifConfig.Quality
	} else if nonStandardSupport.WEBP && webpEnabled {
		optimiseJob.OutType = vips.ImageTypeWEBP
		optimiseJob.Quality = webpConfig.Quality
	} else if nonStandardSupport.WEBP && webpEnabled {
	} else if originalImageFormat == vips.ImageTypePNG && pngEnabled {
		optimiseJob.OutType = vips.ImageTypePNG
		optimiseJob.Quality = pngConfig.Quality
	} else if originalImageFormat != vips.ImageTypePNG && jpegEnabled {
		optimiseJob.OutType = vips.ImageTypeJPEG
		optimiseJob.Quality = jpegConfig.Quality
	} else {
		// no optimizations could be done
		ctx.Response().Header().Set("Content-Optimized", "false")
		return ctx.File(fullPath)
	}

	// reserve job slot
	if err := jobLimiter.AddJob(); err != nil {
		ctx.Response().Header().Del("Cache-Control")
		ctx.Response().Header().Set("Retry-After", "5")
		return ctx.NoContent(http.StatusServiceUnavailable)
	} else {
		defer jobLimiter.RemoveJob()
	}

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	if mimeType, known := core.ImageTypeToMime[originalImageFormat]; known {
		ctx.Response().Header().Set("Source-Type", mimeType)
	}

	ctx.Response().Header().Set("Content-Optimized", "true")
	return ctx.Blob(http.StatusOK, core.ImageTypeToMime[optimiseJob.OutType], img)
}

type ImageQuery struct {
	Preset string `query:"preset"`
	Format string `query:"format"`
}

func getPresetOptimizedImage(ctx echo.Context) error {
	appConfig := ctx.Get(AppConfigKey).(config.Config)
	jobLimiter := ctx.Get(JobLimiterKey).(*core.JobLimiter)
	relativePath, err := url.QueryUnescape(ctx.Param("path"))
	if err != nil {
		return err
	}
	fullPath := path.Join(appConfig.OriginalsBase, relativePath)

	SetCacheHeader(ctx, appConfig)

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
	if query.Preset == "" && query.Format == "" {
		ctx.Response().Header().Set("Content-Optimized", "false")
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
			ctx.Response().Header().Set("Content-Optimized", "false")
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

	if presetConfig, exists := appConfig.PresetOptimize.Presets[query.Preset]; exists {
		if formatConfig, exists := presetConfig.Formats[query.Format]; exists {
			optimiseJob.Quality = formatConfig.Quality
			optimiseJob.MaxWidth = &presetConfig.MaxWidth
		} else {
			return ctx.JSON(http.StatusBadRequest, "unsupported preset+format requested")
		}
	} else {
		return ctx.JSON(http.StatusBadRequest, "unsupported preset requested")
	}

	if err := jobLimiter.AddJob(); err != nil {
		ctx.Response().Header().Del("Cache-Control")
		ctx.Response().Header().Set("Retry-After", "5")
		return ctx.NoContent(http.StatusServiceUnavailable)
	} else {
		defer jobLimiter.RemoveJob()
	}

	img, err := optimiseJob.Optimise()
	if err != nil {
		return err
	}

	ctx.Response().Header().Set("Content-Optimized", "true")
	return ctx.Blob(http.StatusOK, core.ImageTypeToMime[imageFormat], img)
}

func getMetrics(ctx echo.Context) error {
	jobLimiter := ctx.Get(JobLimiterKey).(*core.JobLimiter)
	var vipsMemStats vips.MemoryStats
	vips.ReadVipsMemStats(&vipsMemStats)
	content := fmt.Sprintf(`# HELP active_jobs Current image processing jobs
# TYPE active_jobs gauge
active_jobs %d
# HELP vips_mem Current memory used by libvips
# TYPE vips_mem gauge
vips_mem %d
# HELP vips_mem_high Highest memory used by libvips
# TYPE vips_mem_high counter
vips_mem_high %d
# HELP vips_allocs Current memory allocations used by libvips
# TYPE vips_allocs gauge
vips_allocs %d
# HELP vips_files Current 'files' open by libvips
# TYPE vips_files gauge
vips_files %d
`,
		jobLimiter.Jobs(),
		vipsMemStats.Mem,
		vipsMemStats.MemHigh,
		vipsMemStats.Allocs,
		vipsMemStats.Files,
	)
	return ctx.String(http.StatusOK, content)
}
