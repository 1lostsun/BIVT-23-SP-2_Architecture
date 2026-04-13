package server

import (
	"context"
	"log"
	"net/http"

	rediscache "lab2/internal/cache/redis"
	"lab2/internal/config"
	"lab2/internal/handler"
	"lab2/internal/repo/pg"
	"lab2/internal/usecase"
)

type Config struct {
	Port string `envconfig:"PORT" default:"8080"`
	config.PGConfig
	config.RedisConfig
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

	cache, err := rediscache.New(a.cfg.RedisConfig.Addr(), a.cfg.RedisConfig.Password, a.cfg.RedisConfig.DB)
	if err != nil {
		return err
	}

	uc := usecase.New(repo, cache)
	h := handler.New(uc)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("server listening on :%s", a.cfg.Port)
	return http.ListenAndServe(":"+a.cfg.Port, mux)
}
