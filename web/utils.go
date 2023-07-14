package web

import (
	"fmt"

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

func Run(appConfig config.Config) error {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	v := Validator{}.New()
	e.Validator = &v
	e.Use(
		appConfigMiddleware(appConfig),
		jobRunLimiterMiddleware(appConfig),
	)
	e.GET("/a/:path", getAutoOptimized)
	e.GET("/t/:path", getTypeOptimizedImage)
	address := fmt.Sprintf("%s:%d", appConfig.Bind.Host, appConfig.Bind.Port)
	return e.Start(address)
}
