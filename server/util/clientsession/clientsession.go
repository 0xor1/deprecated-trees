package clientsession

func New() *Store {
	return &Store{
		Cookies: map[string]string{},
	}
}

type Store struct {
	Cookies map[string]string
}
