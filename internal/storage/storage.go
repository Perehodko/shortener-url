package storage

import "errors"

type Storage interface {
	PutURL(shortLink, urlForCuts string) error
	GetURL(shortURL string) (string, error)
}

type URLStorage struct {
	URLs map[string]string
}

// NewURLStore returns a new/empty URLStorage
func NewURLStore() *URLStorage {
	return &URLStorage{
		URLs: make(map[string]string),
	}
}

func (s *URLStorage) PutURL(shortLink, urlForCuts string) error {
	s.URLs[shortLink] = urlForCuts
	return nil
}

func (s *URLStorage) GetURL(shortURL string) (string, error) {
	initialURL, ok := s.URLs[shortURL]
	if !ok {
		return "", errors.New("in map no shortURL from request")
	}
	return initialURL, nil
}
