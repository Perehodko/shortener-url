package storage

type Storage interface {
	PutURLInStorage(shortLink, urlForCuts string)
	GetURLFromStorage(shortURL string) string
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

func (s *URLStorage) PutURLInStorage(shortLink, urlForCuts string) {
	s.URLs[shortLink] = urlForCuts
}

func (s *URLStorage) GetURLFromStorage(shortURL string) string {
	initialURL := s.URLs[shortURL]
	if initialURL == "" {
		return ""
	}
	return initialURL
}
