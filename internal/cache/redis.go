package cache

import (
	"context"
	"fmt"
	"time"

	"choice-matrix-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func ConnectRedis(cfg config.Config) error {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}

	Client = client
	return nil
}
