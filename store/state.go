package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

func (s *Store) decodeState(state string) (*string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(state))
	if err != nil {
		return nil, err
	}

	md := hash.Sum(nil)
	hashedState := hex.EncodeToString(md)

	return &hashedState, nil

}

var ErrInvalidPopKey = errors.New("try to pop invalid key")

func (s *Store) SetState(state string, expired time.Duration) error {
	hashedState, err := s.decodeState(state)
	if err != nil {
		return err
	}

	return s.Store.Set(ctx, fmt.Sprintf("state_%s", *hashedState), "s", expired).Err()
}

func (s *Store) PopState(state string) error {
	hashedState, err := s.decodeState(state)
	if err != nil {
		return err
	}

	value, err := s.Store.GetDel(ctx, fmt.Sprintf("state_%s", *hashedState)).Result()
	if err != nil {
		return err
	}

	if value != "s" {
		return ErrInvalidPopKey
	}
	return nil
}
