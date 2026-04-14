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

type RabbitConfig struct {
	Host     string `envconfig:"RABBIT_HOST" default:"localhost"`
	Port     string `envconfig:"RABBIT_PORT" default:"5672"`
	Username string `envconfig:"RABBIT_USERNAME" default:"guest"`
	Password string `envconfig:"RABBIT_PASSWORD" default:"guest"`
}

func (c *RabbitConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.Username, c.Password, c.Host, c.Port)
}
