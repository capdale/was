package store

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Store) IsBlacklist(token string) (bool, error) {
	err := s.Store.Get(ctx, token).Err()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return true, err
	}
	return true, nil
}

func (s *Store) SetBlacklist(token string, expiration time.Duration) error {
	return s.Store.Set(ctx, token, 0, expiration).Err()
}
