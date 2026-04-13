package rediscache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const notesTTL = 60 * time.Second
const notesKey = "notes:all"

type Cache struct {
	client *redis.Client
}

func New(addr, password string, db int) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &Cache{client: client}, nil
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}

func (c *Cache) Set(ctx context.Context, key string, data []byte) error {
	return c.client.Set(ctx, key, data, notesTTL).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) NotesKey() string {
	return notesKey
}
