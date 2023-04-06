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

func getURLForCut(s storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = work_with_cookie.UserID()
		}
		fmt.Println("getURLForCut - uid", uid)

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
		fmt.Println("getURLForCut -- shortURL", shortURL)

		//записываем в мапу пару shortLink:оригинальная ссылка
		//err = s.PutURL(shortLink, urlForCuts)
		err = s.PutURL("1", shortLink, urlForCuts)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		work_with_cookie.SetUUIDCookie(w, uid)
		w.WriteHeader(http.StatusCreated)
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
		fmt.Println("shortURL", shortURL)

		a, b := r.Cookie("session")
		fmt.Println("redirectTo -- session cookie - a, b", a, b)

		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = work_with_cookie.UserID()
		}
		fmt.Println("redirectTo - uid", uid)

		initialURL, err := s.GetURL("1", shortURL)
		fmt.Println("redirectTo -- initialURL, shortURL", initialURL, shortURL)
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

func shorten(s storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = work_with_cookie.UserID()
		}

		fmt.Println("shorten - uid", uid)

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

		err = s.PutURL("1", shortLink, urlForCuts)
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

func getUserURLs(s storage.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		uid, err := work_with_cookie.ExtractUID(r.Cookies())
		if err != nil {
			http.Error(w, "no links", http.StatusNoContent)
			return
		}

		//encryptedUUIDStr := fmt.Sprintf("%x", encryptedUUID)
		fmt.Println("getUserURLs - uid = UUID", uid)
		getUserURLs, err := s.GetUserURLs("1")
		fmt.Println("getUserURLs", getUserURLs, len(getUserURLs), err)
		if err != nil {
			http.Error(w, "internal error", http.StatusNoContent)
			return
		}
		if len(getUserURLs) == 0 {
			http.Error(w, "no links", http.StatusNoContent)
			return
		}

		type M map[string]interface{}
		var myMapSlice []M

		for i, j := range getUserURLs {
			res := M{"short_url": cfg.BaseURL + "/" + i, "original_url": j}
			myMapSlice = append(myMapSlice, res)
		}

		// or you could use `json.Marshal(myMapSlice)` if you want
		myJson, err := json.MarshalIndent(myMapSlice, "", "    ")
		//fmt.Println(string(myJson))
		if err != nil {
			http.Error(w, "no links", http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(myJson)
		w.WriteHeader(http.StatusOK)

	}
}

func main() {
	//encryptedUUIDKey, _, key, UUID, nonce := work_with_cookie.EncryptedUUID()
	//keyToFunc := fmt.Sprintf("%x", encryptedUUIDKey)

	//fmt.Println("keyToFunc", keyToFunc)

	//uuid := uuid.New()
	//UUID := uuid.String()

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

	r.Post("/", getURLForCut(fileStorage))
	r.Get("/{id}", redirectTo(fileStorage))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(fileStorage))
	r.Get("/api/user/urls", getUserURLs(fileStorage))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
