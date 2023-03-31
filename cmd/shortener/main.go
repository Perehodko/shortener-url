package main

import (
	"crypto/aes"
	"encoding/json"
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

		if len(checkCookieExist(r)) == 0 || !checkKeyIsValid([]byte(key), []byte(encryptedUUID), UUID) {
			cookie := http.Cookie{
				Name:  "session",
				Value: encryptedUUID}
			fmt.Println("encryptedUUID", encryptedUUID, string(encryptedUUID))
			http.SetCookie(w, &cookie)
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
		//err = s.PutURL(shortLink, urlForCuts)
		err = s.PutURL(encryptedUUID, shortLink, urlForCuts)
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

func redirectTo(s storage.Storage, encryptedUUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")

		//initialURL, err := s.GetURL(encryptedUUID)
		initialURL, err := s.GetURL(encryptedUUID, shortURL)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

func shorten(s storage.Storage, encryptedUUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		decoder := json.NewDecoder(r.Body)
		var u URLStruct

		fmt.Println("r.Body!!!", r.Body)
		err := decoder.Decode(&u)
		if err != nil {
			//panic(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//получаю из хранилища результат
		urlForCuts := u.URL
		fmt.Println("urlForCuts-shoren", urlForCuts)

		shortLink := utils.GenerateRandomString()
		shortURL := cfg.BaseURL + "/" + shortLink

		//записываем в мапу encryptedUUID: [shortLink:urlForCuts]
		err = s.PutURL(encryptedUUID, shortLink, urlForCuts)
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

func generateKey() (string, error, string, string) {
	UUID := uuid.New()
	fmt.Println(UUID.String(), UUID)

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
	encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)
	return encryptedUUIDStr, nil, string(key), UUID.String()
}

type ResUsersLinks struct {
	shortLink string `json:"short_url"`
	longLink  string `json:"original_url"`
}

func doSmth(s storage.Storage, encryptedUUIDKey string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		getUserURLs, err := s.GetUserURLs(encryptedUUIDKey)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNoContent)
		} else {
			type M map[string]interface{}

			var myMapSlice []M

			for i, j := range getUserURLs {
				res := M{"short_url": i, "original_url": j}
				myMapSlice = append(myMapSlice, res)

			}

			// or you could use `json.Marshal(myMapSlice)` if you want
			myJson, _ := json.MarshalIndent(myMapSlice, "", "    ")
			fmt.Println(string(myJson))
			w.Write(myJson)

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}
	}
}

func main() {
	encryptedUUIDKey, _, key, UUID := generateKey()
	keyToFunc := encryptedUUIDKey

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
	r.Get("/{id}", redirectTo(fileStorage, keyToFunc))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(fileStorage, keyToFunc))
	r.Get("/api/user/urls", doSmth(fileStorage, keyToFunc))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
