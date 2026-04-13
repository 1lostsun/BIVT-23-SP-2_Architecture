package main

import (
	"lab2/internal/app/server"
	"log"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg server.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}
	a := server.New(cfg)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
