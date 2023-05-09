package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Perehodko/shortener-url/internal/dbstorage"
	"io"
	"os"
)

type Storage interface {
	PutURL(uid, shortLink, urlForCuts string) (string, error)
	GetURL(uid, shortURL string) (string, error)
	GetUserURLs(uid string) (map[string]string, error)
	PutURLsBatch(ctx context.Context, uid string, store map[string][]string) error
}

type URLStorage struct {
	URLs map[string]map[string]string
}

func (s *URLStorage) PutURL(uid, shortLink, urlForCuts string) (string, error) {
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

func (s *URLStorage) GetURL(uid, shortLink string) (string, error) {
	if len(s.URLs[uid]) == 0 {
		return "", errors.New("in map no shortURL from request")
	} else {
		initialURL := s.URLs[uid][shortLink]
		return initialURL, nil
	}
}

func (s *URLStorage) GetUserURLs(uid string) (map[string]string, error) {
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

// file storage
type FileStorage struct {
	ms *URLStorage // сделаем внутреннюю хранилку в памяти тоже интерфейсом, на случай если захотим ее замокать
	f  *os.File
}

func (fs *FileStorage) GetUserURLs(uid string) (map[string]string, error) {
	return fs.ms.GetUserURLs(uid)
}

func (fs *FileStorage) GetURL(uid, key string) (value string, err error) {
	return fs.ms.GetURL(uid, key)
}

func (fs *FileStorage) PutURL(uid, key, value string) (string, error) {
	if _, err := fs.ms.PutURL(uid, key, value); err != nil {
		return "", fmt.Errorf("unable to add new key in memorystorage: %w", err)
	}

	// перезаписываем файл с нуля
	err := fs.f.Truncate(0)
	if err != nil {
		return "", fmt.Errorf("unable to truncate file: %w", err)
	}
	_, err = fs.f.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("unable to get the beginning of file: %w", err)
	}

	err = json.NewEncoder(fs.f).Encode(&fs.ms.URLs)
	if err != nil {
		return "", fmt.Errorf("unable to encode data into the file: %w", err)
	}
	return "", nil
}

func (fs *FileStorage) PutURLsBatch(_ context.Context, uid string, store map[string][]string) (err error) {
	if _, ok := fs.ms.URLs[uid]; !ok {
		fs.ms.URLs[uid] = map[string]string{}
	}
	for CorrelationID, OriginalURL := range store {
		fs.ms.URLs[uid][CorrelationID] = OriginalURL[1]
	}

	// перезаписываем файл с нуля
	err = fs.f.Truncate(0)
	if err != nil {
		return fmt.Errorf("unable to truncate file: %w", err)
	}
	_, err = fs.f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("unable to get the beginning of file: %w", err)
	}

	err = json.NewEncoder(fs.f).Encode(&fs.ms.URLs)
	if err != nil {
		return fmt.Errorf("unable to encode data into the file: %w", err)
	}
	return nil
}

func NewFileStorage(filename string) (*FileStorage, error) { // и здесь мы тоже возвраащем интерфейс
	// мы открываем (или создаем файл если он не существует (os.O_CREATE)), в режиме чтения и записи (os.O_RDWR) и дописываем в конец (os.O_APPEND)
	// у созданного файла будут права 0777 - все пользователи в системе могут его читать, изменять и исполнять
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %s: %w", filename, err)
	}

	// восстанавливаем данные из файла, мы будем их хранить в формате JSON
	m := make(map[string]map[string]string)
	if err := json.NewDecoder(file).Decode(&m); err != nil && err != io.EOF { // проверка на io.EOF тк файл может быть пустой
		return nil, fmt.Errorf("unable to decode contents of file %s: %w", filename, err)
	}

	return &FileStorage{
		ms: &URLStorage{URLs: m},
		f:  file,
	}, nil
}
