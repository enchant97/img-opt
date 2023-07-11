package main

import (
	"log"

	"github.com/enchant97/img-opt/config"
	"github.com/enchant97/img-opt/web"
)

func main() {
	appConfig, err := config.Config{}.LoadFromYaml("config.yaml")
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(web.Run(appConfig))
}
