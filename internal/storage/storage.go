package storage

import "errors"

type Storage interface {
	PutURLInStorage(shortLink, urlForCuts string) error
	GetURLFromStorage(shortURL string) (string, error)
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

func (s *URLStorage) PutURLInStorage(shortLink, urlForCuts string) error {
	s.URLs[shortLink] = urlForCuts
	return nil
}

func (s *URLStorage) GetURLFromStorage(shortURL string) (string, error) {
	initialURL := s.URLs[shortURL]
	if initialURL == "" {
		return "", errors.New("in map no shortURL from request")
	} else {
		return initialURL, nil
	}
}
