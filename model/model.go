package model

import "errors"

var ErrAnonymousCreate = errors.New("invalid permission, this record not allowed to create by anonymous")
