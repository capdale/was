package auth

func (a *Auth) IsBlacklist(token string) (bool, error) {
	return a.Store.IsBlacklist(token)
}
