package main

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Config() *Config {
	return &(s.store.Config)
}

func (s *Service) Shutdown() {

}
