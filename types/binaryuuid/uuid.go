package binaryuuid

import (
	"database/sql/driver"
	"reflect"

	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u *UUID) GormDataType() string {
	return "binary(16)"
}

func (u *UUID) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	uuidBytes, err := uuid.FromBytes(bytes)
	*u = UUID(uuidBytes)
	return err
}

func (u UUID) Value() (driver.Value, error) {
	return uuid.UUID(u).MarshalBinary()
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func MustParse(s string) UUID {
	uuid := uuid.MustParse(s)
	return UUID(uuid)
}

func Parse(s string) (UUID, error) {
	uuid, err := uuid.Parse(s)
	uid := UUID(uuid)
	return uid, err
}

func (b UUID) MarshalText() ([]byte, error) {
	s := uuid.UUID(b)
	str := s.String()
	return []byte(str), nil
}

func (b *UUID) UnmarshalText(by []byte) error {
	s, err := uuid.ParseBytes(by)
	*b = UUID(s)
	return err
}

func FromBytes(by []byte) (UUID, error) {
	b, err := uuid.FromBytes(by)
	nb := UUID(b)
	return nb, err
}

func ValidateUUID(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(UUID); ok {
		return valuer.String()
	}
	return nil
}
