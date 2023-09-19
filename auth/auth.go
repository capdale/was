package auth

type database interface {
}

type store interface {
	IsBlacklist(token string) (bool, error)
	SetBlacklist(token string) error
}

type Auth struct {
	DB     database
	Store  store
	secret []byte
}

func (a *Auth) New(database database, store store) *Auth {
	return &Auth{
		DB:    database,
		Store: store,
	}
}
