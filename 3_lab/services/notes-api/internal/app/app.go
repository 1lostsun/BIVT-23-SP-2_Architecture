package app

import (
	"context"
	"log"
	"net/http"

	"notes-api/internal/config"
	"notes-api/internal/handler"
	"notes-api/internal/publisher"
	"notes-api/internal/repo/pg"
	"notes-api/internal/usecase"
)

type Config struct {
	Port string `envconfig:"PORT" default:"8080"`
	config.PGConfig
	config.RabbitConfig
}

type App struct{ cfg Config }

func New(cfg Config) *App { return &App{cfg: cfg} }

func (a *App) Run() error {
	repo, err := pg.New(a.cfg.PGConfig.DSN())
	if err != nil {
		return err
	}
	if err := repo.Migrate(context.Background()); err != nil {
		return err
	}

	pub, err := publisher.New(a.cfg.RabbitConfig.URL())
	if err != nil {
		return err
	}
	defer pub.Close()

	uc := usecase.New(repo, pub)
	h := handler.New(uc)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("server listening on :%s", a.cfg.Port)
	return http.ListenAndServe(":"+a.cfg.Port, mux)
}
