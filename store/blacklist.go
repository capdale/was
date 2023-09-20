package store

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Store) IsBlacklist(token string) (bool, error) {
	err := s.Store.Get(ctx, "token").Err()
	if err == redis.Nil {
		return false, nil
	}
	return true, err
}

func (s *Store) SetBlacklist(token string, expiration time.Duration) error {
	return s.Store.Set(ctx, "token", nil, expiration).Err()
}
