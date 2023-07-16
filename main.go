package main

import (
	"log"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/enchant97/img-opt/config"
	"github.com/enchant97/img-opt/web"
)

func main() {
	vips.LoggingSettings(nil, vips.LogLevelCritical)
	vips.Startup(&vips.Config{
		MaxCacheFiles: 0,
		MaxCacheMem:   0,
		MaxCacheSize:  0,
	})
	defer vips.Shutdown()

	appConfig, err := config.Config{}.LoadFromYaml("config.yaml")
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(web.Run(appConfig))
}
