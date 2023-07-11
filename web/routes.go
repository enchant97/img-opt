package web

import (
	"log"
	"path"

	"github.com/enchant97/img-opt/config"
	"github.com/labstack/echo/v4"
)

type ImageQuery struct {
	Type   string `query:"type" validate:"required"`
	Format string `query:"format" validate:"required"`
}

func getOptimizedImage(ctx echo.Context) error {
	var query ImageQuery
	if err := BindAndValidate(ctx, &query); err != nil {
		return err
	}
	log.Println(ctx.Param("path"))
	log.Println(query)

    appConfig := ctx.Get(AppConfigKey).(config.Config)

    fullPath := path.Join(appConfig.OriginalsBase, ctx.Param("path"))
    log.Println(fullPath)

	return ctx.NoContent(200)
}
