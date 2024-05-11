package claimer

import (
	"database/sql/driver"

	"github.com/capdale/was/types/binaryuuid"
	"github.com/google/uuid"
)

type Claimer uuid.UUID

func New(uuid *binaryuuid.UUID) *Claimer {
	copy := *uuid
	claimer := Claimer(copy)
	return &claimer
}

func (c *Claimer) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	uuidBytes, err := uuid.FromBytes(bytes)
	*c = Claimer(uuidBytes)
	return err
}

func (c Claimer) Value() (driver.Value, error) {
	return uuid.UUID(c).MarshalBinary()
}

func (c Claimer) String() string {
	return uuid.UUID(c).String()
}

func MustParse(s string) Claimer {
	uuid := uuid.MustParse(s)
	return Claimer(uuid)
}

func Parse(s string) (Claimer, error) {
	uuid, err := uuid.Parse(s)
	uid := Claimer(uuid)
	return uid, err
}

func (c Claimer) MarshalText() ([]byte, error) {
	s := uuid.UUID(c)
	str := s.String()
	return []byte(str), nil
}

func (c *Claimer) UnmarshalText(by []byte) error {
	s, err := uuid.ParseBytes(by)
	*c = Claimer(s)
	return err
}

// no error
func (c Claimer) MarshalBinary() ([]byte, error) {
	return c[:], nil
}
