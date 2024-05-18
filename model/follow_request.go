package model

import "github.com/capdale/was/types/binaryuuid"

type FollowRequest struct {
	Code     binaryuuid.UUID `json:"code"`
	Username string          `json:"username"`
}
