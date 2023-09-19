package auth

func (a *Auth) IsBlacklist(token string) bool {
	_, err := a.Store.IsBlacklist(token)
	return err != nil
}

func (a *Auth) SetBlacklist(token string) error {
	err := a.Store.SetBlacklist(token)
	if err != nil {
		return err
	}
	return nil
}
