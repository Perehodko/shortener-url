package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Storage interface {
	//PutURL(shortLink, urlForCuts string) error
	//GetURL(shortURL string) (string, error)
	PutURL(uid, shortLink, urlForCuts string) error
	GetURL(uid string) (string, error)
}

func (s *URLStorage) PutURL(uid, shortLink, urlForCuts string) error {
	s.URLs[uid] = append(s.URLs[uid], []string{shortLink, urlForCuts}...)
	return nil
}

func (s *URLStorage) GetURL(uid string) (string, error) {
	if len(s.URLs[uid]) == 0 {
		return "", errors.New("in map no shortURL from request")
	} else {
		initialURL := s.URLs[uid][0]
		return initialURL, nil
	}
}

type URLStorage struct {
	URLs map[string][]string
}

// NewURLStore returns a new/empty URLStorage
func NewURLStore() *URLStorage {
	return &URLStorage{
		URLs: make(map[string][]string),
	}
}

//func (s *URLStorage) PutURL(shortLink, urlForCuts string) error {
//	s.URLs[shortLink] = urlForCuts
//	return nil
//}

//func (s *URLStorage) GetURL(shortURL string) (string, error) {
//	initialURL, ok := s.URLs[shortURL]
//	if !ok {
//		return "", errors.New("in map no shortURL from request")
//	}
//	return initialURL, nil
//}

func NewMemStorage() *URLStorage { //  возвращаем интерфейс
	return &URLStorage{URLs: make(map[string][]string)}
}

// file
type FileStorage struct {
	ms *URLStorage // сделаем внутреннюю хранилку в памяти тоже интерфейсом, на случай если захотим ее замокать
	f  *os.File
}

func (fs *FileStorage) GetURL(key string) (value string, err error) {
	return fs.ms.GetURL(key)
}

func (fs *FileStorage) PutURL(uid, key, value string) (err error) {
	if err = fs.ms.PutURL(uid, key, value); err != nil {
		return fmt.Errorf("unable to add new key in memorystorage: %w", err)
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
	m := make(map[string][]string)
	if err := json.NewDecoder(file).Decode(&m); err != nil && err != io.EOF { // проверка на io.EOF тк файл может быть пустой
		return nil, fmt.Errorf("unable to decode contents of file %s: %w", filename, err)
	}

	return &FileStorage{
		ms: &URLStorage{URLs: m},
		f:  file,
	}, nil
}
