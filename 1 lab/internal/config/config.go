package config

import "fmt"

type PGConfig struct {
	Username string `envconfig:"PG_USERNAME" default:"postgres"`
	Password string `envconfig:"PG_PASSWORD" default:"postgres"`
	Host     string `envconfig:"PG_HOST" default:"localhost"`
	Port     string `envconfig:"PG_PORT" default:"5432"`
	Database string `envconfig:"PG_DATABASE" default:"postgres"`
}

func (c *PGConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}
