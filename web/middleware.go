package web

import (
	"github.com/enchant97/img-opt/config"
	"github.com/labstack/echo/v4"
)

const AppConfigKey = "AppConfig"

func appConfigMiddleware(appConfig config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(AppConfigKey, appConfig)
			return next(c)
		}
	}
}
