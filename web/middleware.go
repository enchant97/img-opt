package web

import (
	"github.com/enchant97/img-opt/config"
	"github.com/enchant97/img-opt/core"
	"github.com/labstack/echo/v4"
)

const AppConfigKey = "AppConfig"
const JobLimiterKey = "JobLimiter"

func appConfigMiddleware(appConfig config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(AppConfigKey, appConfig)
			return next(c)
		}
	}
}

func jobRunLimiterMiddleware(appConfig config.Config) echo.MiddlewareFunc {
	jobLimiter := core.NewJobLimiter(appConfig.JobLimit)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(JobLimiterKey, jobLimiter)
			return next(c)
		}
	}
}

func serverNameMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Server", core.AppName)
		return next(c)
	}
}
