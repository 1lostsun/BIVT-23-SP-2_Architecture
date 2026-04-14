package main

import (
	"log"

	"notes-api/internal/app"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg app.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}
	if err := app.New(cfg).Run(); err != nil {
		log.Fatal(err)
	}
}
