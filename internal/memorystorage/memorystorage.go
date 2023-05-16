package memorystorage

import (
	"context"
	"errors"
	"github.com/Perehodko/shortener-url/internal/dbstorage"
)

type Storage interface {
	PutURL(ctx context.Context, uid, shortLink, urlForCuts string) (string, error)
	GetURL(ctx context.Context, uid, shortURL string) (string, error)
	GetUserURLs(ctx context.Context, uid string) (map[string]string, error)
	PutURLsBatch(ctx context.Context, uid string, store map[string][]string) error
}

type URLStorage struct {
	URLs map[string]map[string]string
}

func (s *URLStorage) PutURL(_ context.Context, uid, shortLink, urlForCuts string) (string, error) {
	sh := shortLink
	if _, ok := s.URLs[uid]; !ok {
		s.URLs[uid] = map[string]string{}
	}

	for key, value := range s.URLs[uid] {
		if value == urlForCuts {
			return key, dbstorage.NewLinkExistsError(key)
		}
	}

	s.URLs[uid][sh] = urlForCuts
	return sh, nil
}

func (s *URLStorage) GetURL(_ context.Context, uid, shortLink string) (string, error) {
	if len(s.URLs[uid]) == 0 {
		return "", errors.New("in map no shortURL from request")
	} else {
		initialURL := s.URLs[uid][shortLink]
		return initialURL, nil
	}
}

func (s *URLStorage) GetUserURLs(_ context.Context, uid string) (map[string]string, error) {
	if _, ok := s.URLs[uid]; !ok {
		return map[string]string{}, errors.New("in map no shortURL from request")
	} else {
		return s.URLs[uid], nil
	}
}

func (s *URLStorage) PutURLsBatch(_ context.Context, uid string, store map[string][]string) error {
	if _, ok := s.URLs[uid]; !ok {
		s.URLs[uid] = map[string]string{}
	}
	for CorrelationID, OriginalURL := range store {
		s.URLs[uid][CorrelationID] = OriginalURL[1]
	}
	return nil
}

// NewURLStore returns a new/empty URLStorage
func NewURLStore() *URLStorage {
	return &URLStorage{
		URLs: make(map[string]map[string]string),
	}
}

func NewMemStorage() *URLStorage { //  возвращаем интерфейс
	return &URLStorage{URLs: make(map[string]map[string]string)}
}
