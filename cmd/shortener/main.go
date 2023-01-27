package main

import (
	"encoding/json"
	"flag"
	"github.com/Perehodko/shortener-url/internal/middlewares"
	"github.com/Perehodko/shortener-url/internal/storage"
	"github.com/Perehodko/shortener-url/internal/utils"
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

func getURLForCut(s storage.Storage, flagb string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		// читаем Body
		defer r.Body.Close()
		bodyData, err := io.ReadAll(r.Body)
		// обрабатываем ошибку
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlForCuts := string(bodyData)

		BaseURL := cfg.BaseURL

		if len(BaseURL) == 0 {
			BaseURL = flagb
		}

		shortLink := utils.GenerateRandomString()
		shortURL := BaseURL + "/" + shortLink

		//записываем в мапу пару shortLink:оригинальная ссылка
		err = s.PutURL(shortLink, urlForCuts)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		//fmt.Println(storage.URLStorage{})

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

		if initialURL == "" {
			http.Error(w, "URl not in storage", http.StatusBadRequest)
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

func shorten(s storage.Storage, flag1 string) func(w http.ResponseWriter, r *http.Request) {
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

		BaseURL := cfg.BaseURL
		if len(BaseURL) == 0 {
			BaseURL = flag1
		}

		shortLink := utils.GenerateRandomString()
		//shortURL := "http://" + getHost + "/" + shortLink
		shortURL := BaseURL + "/" + shortLink

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

func NewStorage(filename string) (storage.Storage, error) {
	if len(filename) != 0 {
		fileStorage, err := storage.NewFileStorage(filename)

		if err != nil {
			log.Fatalf("unable to create file storage: %v", err)
		}
		return fileStorage, err
	} else {
		fileStorage := storage.NewMemStorage()
		return fileStorage, nil
	}

}

func main() {
	baseURL := flag.String("b", "http://localhost:8080", "BASE_URL из cl")
	severAddress := flag.String("a", ":8080", "SERVER_ADDRESS из cl")
	fileStoragePath := flag.String("f", "store.json", "FILE_STORAGE_PATH из cl")
	flag.Parse()

	fn := cfg.FileName
	if len(fn) == 0 {
		fn = *fileStoragePath
	}

	fileStorage, err := NewStorage(fn)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	err = env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

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

	r.Post("/", getURLForCut(fileStorage, *baseURL))
	r.Get("/{id}", redirectTo(fileStorage))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(fileStorage, *baseURL))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
