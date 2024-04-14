package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(url string) (*Redis, error) {
	ctx := context.Background()

	redisOpts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed parse redis url: %w", err)
	}

	client := redis.NewClient(redisOpts)

	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed ping redis: %w", err)
	}

	return &Redis{Client: client}, nil
}

func (r *Redis) Close() error {
	if r.Client != nil {
		err := r.Client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
