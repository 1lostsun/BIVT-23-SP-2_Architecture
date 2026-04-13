package app

import (
	"context"
	"log"
	"net/http"

	"arch/internal/config"
	"arch/internal/handler"
	"arch/internal/repo/pg"
	"arch/internal/usecase"
)

type Config struct {
	Port string `envconfig:"PORT" default:"8080"`
	config.PGConfig
}

type App struct {
	cfg Config
}

func New(cfg Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run() error {
	repo, err := pg.New(a.cfg.PGConfig.DSN())
	if err != nil {
		return err
	}

	if err := repo.Migrate(context.Background()); err != nil {
		return err
	}

	uc := usecase.New(repo)
	h := handler.New(uc)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("server listening on :%s", a.cfg.Port)
	return http.ListenAndServe(":"+a.cfg.Port, mux)
}
