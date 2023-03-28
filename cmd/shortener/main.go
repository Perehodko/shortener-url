package main

import (
	"crypto/aes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Perehodko/shortener-url/internal/middlewares"
	"github.com/Perehodko/shortener-url/internal/storage"
	"github.com/Perehodko/shortener-url/internal/utils"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FileName      string `env:"FILE_STORAGE_PATH"`
}

var cfg Config

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func checkCookieExist(r *http.Request) string {
	cookie, err := r.Cookie("session")
	var sessionCookie string
	if err == nil {
		sessionCookie = cookie.Value
	} else if err != http.ErrNoCookie {
		log.Println(err)
	}
	return sessionCookie
}

func checkKeyIsValid(key []byte, encryptedUUID []byte, UUID string) bool {
	// получаем cipher.Block
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// расшифровываем
	src2 := make([]byte, aes.BlockSize)
	aesblock.Decrypt(src2, encryptedUUID)
	fmt.Printf("decrypted: %s\n", src2)

	if UUID == string(src2) {
		return true
	} else {
		return false
	}
}

func getURLForCut(s storage.Storage, encryptedUUID string, key string, UUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		if len(checkCookieExist(r)) == 0 || checkKeyIsValid([]byte(key), []byte(encryptedUUID), UUID) {
			cookie := http.Cookie{
				Name:  "session",
				Value: encryptedUUID}

			http.SetCookie(w, &cookie)
			//отладка
			//fmt.Println(&cookie, "&cookie")
		}

		// читаем Body
		defer r.Body.Close()
		bodyData, err := io.ReadAll(r.Body)
		// обрабатываем ошибку
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlForCuts := string(bodyData)

		shortLink := utils.GenerateRandomString()
		shortURL := cfg.BaseURL + "/" + shortLink

		//записываем в мапу пару shortLink:оригинальная ссылка
		err = s.PutURL(shortLink, urlForCuts)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Write([]byte(shortURL))
	}
}

func notFoundFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found"))
}

func redirectTo(s storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		initialURL, err := s.GetURL(shortURL)

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Location", initialURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

type URLStruct struct {
	URL string `json:"url"`
}

type Res struct {
	Result string `json:"result"`
}

func shorten(s storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusCreated)

		decoder := json.NewDecoder(r.Body)
		var u URLStruct

		err := decoder.Decode(&u)
		if err != nil {
			panic(err)
		}
		//получаю из хранилища результат
		urlForCuts := u.URL

		shortLink := utils.GenerateRandomString()
		shortURL := cfg.BaseURL + "/" + shortLink

		//записываем в мапу пару shortLink:оригинальная ссылка
		err = s.PutURL(shortLink, urlForCuts)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		tx := Res{Result: shortURL}
		// преобразуем tx в JSON-формат
		txBz, err := json.Marshal(tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(txBz)
	}
}

func NewStorage(fileName string) (storage.Storage, error) {
	if len(fileName) != 0 {
		fileStorage, err := storage.NewFileStorage(fileName)
		return fileStorage, err
	} else {
		fileStorage := storage.NewMemStorage()
		return fileStorage, nil
	}
}

func checkKeyAndRead() string {
	fileDirectory := "/Users/nperekhodko/Desktop/I/yandex_precticum/shortener-url/cmd/shortener/"
	fileName := "key.txt"
	if _, err := os.Stat(fileDirectory + fileName); err == nil {
		// path/to/whatever exists
		file, err := os.OpenFile(fileDirectory+fileName, os.O_RDWR|os.O_CREATE, 0755)

		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err = file.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		buf := make([]byte, 1024)
		keyFromFile, err := file.Read(buf)
		//fmt.Println("с и без стр", keyFromFile, string(keyFromFile))
		return string(keyFromFile)
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return ""
	}
	return ""
}

func writeToFile(key string) {
	filePath := "/Users/nperekhodko/Desktop/I/yandex_precticum/shortener-url/cmd/shortener/"
	fileName := "key.txt"
	f, err := os.Create(filePath + fileName)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err2 := f.Write([]byte(key))

	if err2 != nil {
		log.Fatal(err2)
	}
}

func generateKey() (string, error, string, string) {
	//fmt.Println("currentCoockie", currentCoockie)
	//cookie
	UUID := uuid.New()
	//fmt.Println(UUID.String(), "UUID")

	//подписываю куки
	//1 перевожу в байты
	uuidByte := []byte(UUID.String()) // данные, которые хотим зашифровать
	//2 константа aes.BlockSize определяет размер блока и равна 16 байтам
	// будем использовать AES256, создав ключ длиной 32 байта
	key, err := generateRandom(aes.BlockSize) // ключ шифрования
	//fmt.Println("crypto key", key)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	//3 получаем cipher.Block
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	//4 зашифровываем
	encryptedUUID := make([]byte, aes.BlockSize)
	aesblock.Encrypt(encryptedUUID, uuidByte)
	//fmt.Printf("encrypted: %x\n", encryptedUUID)
	//fmt.Println("encrypted string: ", string(encryptedUUID))

	return string(encryptedUUID), nil, string(key), UUID.String()
}

func main() {
	//keyToFunc := ""
	//
	//isKeyExist := checkKeyAndRead()

	encryptedUUIDKey, _, key, UUID := generateKey()
	keyToFunc := encryptedUUIDKey
	//if len(isKeyExist) == 0 {
	//	fmt.Println("encryptedUUIDKey to write in file", encryptedUUIDKey)
	//	writeToFile(encryptedUUIDKey)
	//	keyToFunc = encryptedUUIDKey
	//} else {
	//	keyToFunc = isKeyExist
	//}
	//fmt.Println("keyToFunc", keyToFunc)

	baseURL := flag.String("b", "http://localhost:8080", "BASE_URL из cl")
	severAddress := flag.String("a", ":8080", "SERVER_ADDRESS из cl")
	fileStoragePath := flag.String("f", "store.json", "FILE_STORAGE_PATH из cl")
	flag.Parse()

	// вставляем в структуру cfg значения из флагов
	cfg.ServerAddress = *severAddress
	cfg.BaseURL = *baseURL
	cfg.FileName = *fileStoragePath

	// перезатираем их значениями энвов
	// если значения в энве для поля структуры нет - то в поле останется значение из флага
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	fileStorage, err := NewStorage(cfg.FileName)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	ServerAddr := cfg.ServerAddress
	if len(ServerAddr) == 0 {
		ServerAddr = *severAddress
	}

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Compress(5),
		middlewares.Decompress)

	r.Post("/", getURLForCut(fileStorage, keyToFunc, key, UUID))
	r.Get("/{id}", redirectTo(fileStorage))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(fileStorage))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
