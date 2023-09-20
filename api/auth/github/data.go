package githubAuth

import (
	"context"
	"encoding/json"

	authapi "github.com/capdale/was/api/auth"
	"golang.org/x/oauth2"
)

type userInfo struct {
	Login string `json:"login"`
}

type emailInfo struct {
	Email    string
	Primary  bool
	Verified bool
}

func (g *GithubAuth) getUserId(ctx context.Context, t *oauth2.Token) (string, error) {
	client := g.OAuthConfig.Client(ctx, t)
	body, err := authapi.GetBody(client, userInfoEndpoint)
	if err != nil {
		return "", err
	}

	info := &userInfo{}
	err = json.Unmarshal(body, info)
	if err != nil {
		return "", err
	}
	return info.Login, nil
}

func (g *GithubAuth) getEmail(ctx context.Context, t *oauth2.Token) (string, error) {
	client := g.OAuthConfig.Client(ctx, t)
	body, err := authapi.GetBody(client, emailInfoEndpoint)
	if err != nil {
		return "", err
	}

	infos := &[]emailInfo{}
	err = json.Unmarshal(body, infos)
	if err != nil {
		return "", err
	}

	email := ""
	for _, info := range *infos {
		if info.Verified && info.Primary {
			email = info.Email
			break
		}
	}
	if email == "" {
		return "", authapi.ErrNoValidEmail
	}
	return email, nil
}
