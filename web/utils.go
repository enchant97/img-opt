package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/enchant97/img-opt/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Ease of use method, when binding & validation is needed.
func BindAndValidate(ctx echo.Context, i interface{}) error {
	if err := ctx.Bind(i); err != nil {
		return err
	} else if err := ctx.Validate(i); err != nil {
		return err
	}
	return nil
}

// Handle ETag stuff
// Returns (need new content)
func HandleETag(ctx echo.Context, currentETag string) bool {
	ctx.Response().Header().Add("ETag", "\""+currentETag+"\"")
	if headerValue := ctx.Request().Header.Get("If-None-Match"); headerValue != "" {
		tags := strings.Split(headerValue, ",")
		for _, tag := range tags {
			tag = strings.Trim(strings.TrimSpace(tag), "\"")
			if tag == currentETag {
				return false
			}
		}
	}
	return true
}

func SetCacheHeader(ctx echo.Context, appConfig config.Config) {
	var maxAge uint = 86400
	var maxStaleAge uint = 7200
	if appConfig.BrowserTTL != nil {
		maxAge = appConfig.BrowserTTL.Max
		maxStaleAge = appConfig.BrowserTTL.MaxStale
	}
	ctx.Response().Header().Set(
		"Cache-Control",
		fmt.Sprintf("public, max-age=%d, stale-while-revalidate=%d", maxAge, maxStaleAge),
	)
}

func Run(appConfig config.Config) error {
	e := echo.New()
	e.HideBanner = true
	e.RouteNotFound("*", func(c echo.Context) error { return c.NoContent(http.StatusNotFound) })
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	v := Validator{}.New()
	e.Validator = &v
	e.Use(
		serverNameMiddleware,
		appConfigMiddleware(appConfig),
		jobRunLimiterMiddleware(appConfig),
	)
	if appConfig.Metrics {
		e.GET("/metrics", getMetrics)
	}
	e.GET("/o/:path", getOriginalImage)
	e.GET("/a/:path", getAutoOptimized)
	e.GET("/p/:path", getPresetOptimizedImage)

	address := fmt.Sprintf("%s:%d", appConfig.Bind.Host, appConfig.Bind.Port)
	if appConfig.Bind.TLS != nil {
		return e.StartTLS(
			address,
			appConfig.Bind.TLS.CertFile,
			appConfig.Bind.TLS.KeyFile,
		)
	} else {
		return e.Start(address)
	}
}
