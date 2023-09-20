package auth

func (a *Auth) IsBlacklist(token string) bool {
	_, err := a.Store.IsBlacklist(token)
	return err != nil
}
