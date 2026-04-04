package app

type App struct {
	Port string `envconfig:"PORT" default:"8080"`
	Config
}
