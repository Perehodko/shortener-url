package filetorage

import (
	"context"
	"encoding/json"
	"fmt"
	st "github.com/Perehodko/shortener-url/internal/memorystorage"
	"io"
	"os"
)

// file memorystorage
type FileStorage struct {
	ms *st.URLStorage // сделаем внутреннюю хранилку в памяти тоже интерфейсом, на случай если захотим ее замокать
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
		ms: &st.URLStorage{URLs: m},
		f:  file,
	}, nil
}
