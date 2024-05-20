package kakao

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func getUserEmailWithAccessToken(accessToken string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s?property_keys=[\"kakao_account.email\"]", userInfoEndpoint), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-type", "application/x-www-form-urlencoded;charset=utf-8")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return getEmailFromBody(&b)
}

func (k *KakaoAuth) getUserEmail(ctx context.Context, t *oauth2.Token) (string, error) {
	client := k.OAuthConfig.Client(ctx, t)
	body, err := authapi.GetBody(client, userInfoEndpoint)
	if err != nil {
		return "", err
	}
	return getEmailFromBody(&body)
}

func getEmailFromBody(body *[]byte) (string, error) {
	info := &userInfo{}
	if err := json.Unmarshal(*body, info); err != nil {
		return "", err
	}
	if !(info.Account.IsEmailValid || info.Account.IsEmailVerified) {
		return "", fmt.Errorf("%w, email valid: %v, verified: %v", authapi.ErrNoValidEmail, info.Account.IsEmailValid, info.Account.IsEmailVerified)
	}
	return info.Account.Email, nil
}
