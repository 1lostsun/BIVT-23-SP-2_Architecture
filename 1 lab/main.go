package main

import (
	"log"

	"arch/internal/app"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg app.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	a := app.New(cfg)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
