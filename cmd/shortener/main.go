package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Perehodko/shortener-url/internal/middlewares"
	"github.com/Perehodko/shortener-url/internal/storage"
	"github.com/Perehodko/shortener-url/internal/utils"
	"github.com/Perehodko/shortener-url/internal/work-with-cookie"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FileName      string `env:"FILE_STORAGE_PATH"`
}

var cfg Config

func getURLForCut(s storage.Storage, encryptedUUID []byte, key string, UUID string, nonce []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)

		if len(work_with_cookie.CheckCookieExist(r)) == 0 || !work_with_cookie.CheckKeyIsValid([]byte(key), encryptedUUID, UUID, nonce) {
			cookie := http.Cookie{
				Name:  "session",
				Value: encryptedUUIDStr}
			//fmt.Println("encryptedUUID", encryptedUUID, string(encryptedUUID), encryptedUUIDStr)
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
		err = s.PutURL(encryptedUUIDStr, shortLink, urlForCuts)
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
		fmt.Println("shortURL", shortURL)

		//initialURL, err := s.GetURL(encryptedUUID)
		initialURL, err := s.GetURL(encryptedUUID, shortURL)
		fmt.Println("encryptedUUID, initialURL, shortURL, c", encryptedUUID, initialURL, shortURL)
		fmt.Println("err", err)
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
		//encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)
		//fmt.Println("shorten - encryptedUUID", encryptedUUID)
		//fmt.Println("shorten - encryptedUUIDStr", encryptedUUIDStr)

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

func encryptesUUID() ([]byte, error, string, string, []byte) {
	UUID := uuid.New()
	fmt.Println(UUID.String(), UUID)

	src := []byte(UUID.String()) // данные, которые хотим зашифровать
	fmt.Printf("original: %s\n", src)

	// будем использовать AES256, создав ключ длиной 32 байта
	key, err := work_with_cookie.GenerateRandom(2 * aes.BlockSize) // ключ шифрования
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// создаём вектор инициализации
	nonce, err := work_with_cookie.GenerateRandom(aesgcm.NonceSize())
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println("nonce", nonce)

	encryptedUUID := aesgcm.Seal(nil, nonce, src, nil) // зашифровываем
	//fmt.Printf("encrypted: %x\n", encryptedUUID)

	//encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)
	//fmt.Println(encryptedUUIDStr)

	return encryptedUUID, nil, string(key), UUID.String(), nonce
}

func getUserURLs(s storage.Storage, encryptedUUIDKey []byte, key, UUID string, nonce []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUIDKey)

		getUserURLs, err := s.GetUserURLs(encryptedUUIDStr)
		//fmt.Println("getUserURLs", getUserURLs, len(getUserURLs), err)

		cookieIsValid := work_with_cookie.CheckKeyIsValid([]byte(key), encryptedUUIDKey, UUID, nonce)
		//fmt.Println("cookieIsValid???", cookieIsValid)

		if err != nil || !cookieIsValid || len(getUserURLs) == 0 {
			//fmt.Println("err in if", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
		} else {
			type M map[string]interface{}

			var myMapSlice []M

			for i, j := range getUserURLs {
				res := M{"short_url": cfg.BaseURL + "/" + i, "original_url": j}
				myMapSlice = append(myMapSlice, res)

			}

			// or you could use `json.Marshal(myMapSlice)` if you want
			myJson, _ := json.MarshalIndent(myMapSlice, "", "    ")
			//fmt.Println(string(myJson))

			w.Header().Set("Content-Type", "application/json")
			w.Write(myJson)
			w.WriteHeader(http.StatusOK)
		}
	}
}

func main() {
	encryptedUUIDKey, _, key, UUID, nonce := encryptesUUID()
	keyToFunc := fmt.Sprintf("%x", encryptedUUIDKey)

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

	r.Post("/", getURLForCut(fileStorage, encryptedUUIDKey, key, UUID, nonce))
	r.Get("/{id}", redirectTo(fileStorage, keyToFunc))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(fileStorage, keyToFunc))
	r.Get("/api/user/urls", getUserURLs(fileStorage, encryptedUUIDKey, key, UUID, nonce))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
