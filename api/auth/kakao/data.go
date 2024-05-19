package kakao

import (
	"context"
	"encoding/json"

	authapi "github.com/capdale/was/api/auth"
	"golang.org/x/oauth2"
)

type userInfo struct {
	Account kakaoAccount `json:"kakao_account"`
}

type kakaoAccount struct {
	IsEmailValid    bool   `json:"is_email_valid"`
	IsEmailVerified bool   `json:"is_email_verified"`
	Email           string `json:"email"`
}

func (k *KakaoAuth) getUserEmail(ctx context.Context, t *oauth2.Token) (string, error) {
	client := k.OAuthConfig.Client(ctx, t)
	body, err := authapi.GetBody(client, userInfoEndpoint)
	if err != nil {
		return "", err
	}

	info := &userInfo{}
	err = json.Unmarshal(body, info)
	if err != nil {
		return "", err
	}

	if !(info.Account.IsEmailValid || info.Account.IsEmailVerified) {
		return "", authapi.ErrNoValidEmail
	}

	return info.Account.Email, nil
}
