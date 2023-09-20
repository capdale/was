package store

import (
	"context"
	"errors"

	"github.com/capdale/was/config"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	Store *redis.Client
}

var ctx = context.Background()

var ErrFailConnectRedis = errors.New("fail to connect redis")

func New(redisConfig *config.Redis) (*Store, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Address,
		Password: redisConfig.Password,
		DB:       redisConfig.Db,
	})

	if rdb == nil {
		return nil, ErrFailConnectRedis
	}
	return &Store{
		Store: rdb,
	}, nil
}
