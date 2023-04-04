package main

import (
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

func getURLForCut(s storage.Storage, UUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = UUID
		}
		//encryptedUUID, err := work_with_cookie.ExtractUID(r.Cookies())
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//}
		//encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)

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
		err = s.PutURL(uid, shortLink, urlForCuts)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		work_with_cookie.SetUUIDCookie(w, uid)
		w.Write([]byte(shortURL))
	}
}

func notFoundFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found"))
}

func redirectTo(s storage.Storage, UUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		fmt.Println("shortURL", shortURL)

		//initialURL, err := s.GetURL(encryptedUUID)
		//encryptedUUID, err := work_with_cookie.ExtractUID(r.Cookies())
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//}
		//encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)

		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = UUID
		}

		initialURL, err := s.GetURL(uid, shortURL)
		//fmt.Println("encryptedUUID, initialURL, shortURL, c", encryptedUUID, initialURL, shortURL)
		fmt.Println("err", err)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		work_with_cookie.SetUUIDCookie(w, uid)
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

func shorten(s storage.Storage, UUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		//encryptedUUID, err := work_with_cookie.ExtractUID(r.Cookies())
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//}

		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = UUID
		}

		fmt.Println("uid", uid)

		decoder := json.NewDecoder(r.Body)
		var u URLStruct

		err = decoder.Decode(&u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//получаю из хранилища результат
		urlForCuts := u.URL
		//fmt.Println("urlForCuts-shoren", urlForCuts)

		shortLink := utils.GenerateRandomString()
		shortURL := cfg.BaseURL + "/" + shortLink

		err = s.PutURL(uid, shortLink, urlForCuts)
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

		work_with_cookie.SetUUIDCookie(w, uid)
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

func getUserURLs(s storage.Storage, UUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			http.Error(w, "no links in storage at current UUID", http.StatusNoContent)
			return
		}
		//uid = UUID

		fmt.Println("getUserURLs - uid-2", uid)

		//encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)

		getUserURLs, err := s.GetUserURLs(uid)
		fmt.Println("getUserURLs", getUserURLs, len(getUserURLs), err)
		if len(getUserURLs) == 0 {
			http.Error(w, "no links", http.StatusNoContent)
			return
		}

		//cookieIsValid := work_with_cookie.CheckKeyIsValid([]byte(key), encryptedUUIDKey, UUID, nonce)
		//fmt.Println("cookieIsValid???", cookieIsValid)

		//if err != nil {
		//	//fmt.Println("err in if", err)
		//	w.Header().Set("Content-Type", "application/json")
		//	w.WriteHeader(http.StatusNoContent)
		//} else {
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

func main() {
	//encryptedUUIDKey, _, key, UUID, nonce := work_with_cookie.EncryptedUUID()
	//keyToFunc := fmt.Sprintf("%x", encryptedUUIDKey)

	//fmt.Println("keyToFunc", keyToFunc)

	uuid := uuid.New()
	UUID := uuid.String()

	//var key, _ = work_with_cookie.GenerateRandom(32)

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

	r.Post("/", getURLForCut(fileStorage, UUID))
	r.Get("/{id}", redirectTo(fileStorage, UUID))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(fileStorage, UUID))
	r.Get("/api/user/urls", getUserURLs(fileStorage, UUID))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
