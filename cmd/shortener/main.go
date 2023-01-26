package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Perehodko/shortener-url/internal/storage"
	"github.com/Perehodko/shortener-url/internal/utils"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
	"os"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FileName      string `env:"FILE_STORAGE_PATH"`
}

var cfg Config

func NewMemStorage() storage.Storage { // обрати внимание, что возвращаем интерфейс
	return &storage.URLStorage{URLs: make(map[string]string)}
}

// file
type FileStorage struct {
	ms *storage.URLStorage // сделаем внутреннюю хранилку в памяти тоже интерфейсом, на случай если захотим ее замокать
	f  *os.File
}

func (fs *FileStorage) GetURL(key string) (value string, err error) {
	return fs.ms.GetURL(key)
}

func (fs *FileStorage) PutURL(key, value string) (err error) {
	if err = fs.ms.PutURL(key, value); err != nil {
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

func NewFileStorage(filename string) (storage.Storage, error) { // и здесь мы тоже возвраащем интерфейс
	// мы открываем (или создаем файл если он не существует (os.O_CREATE)), в режиме чтения и записи (os.O_RDWR) и дописываем в конец (os.O_APPEND)
	// у созданного файла будут права 0777 - все пользователи в системе могут его читать, изменять и исполнять
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %s: %w", filename, err)
	}

	// восстанавливаем данные из файла, мы будем их хранить в формате JSON
	m := make(map[string]string)
	if err := json.NewDecoder(file).Decode(&m); err != nil && err != io.EOF { // проверка на io.EOF тк файл может быть пустой
		return nil, fmt.Errorf("unable to decode contents of file %s: %w", filename, err)
	}

	return &FileStorage{
		ms: &storage.URLStorage{URLs: m},
		f:  file,
	}, nil
}

func getURLForCut(s storage.Storage, flagb string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusCreated)

		// читаем Body
		defer r.Body.Close()
		bodyData, err := io.ReadAll(r.Body)
		// обрабатываем ошибку
		if err != nil {
			http.Error(w, err.Error(), 500)
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

		//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		//w.Header().Set("Content-Type", "gzip")
		w.Header().Set("Location", initialURL)
		w.Header().Set("Content-Type", "text/html")
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
		w.Header().Set("Accept-Encoding", "gzip")
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
			panic(err)
		}
		w.Write(txBz)
	}
}

func File(flagf string) (storage.Storage, error) {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	fn := cfg.FileName
	if len(fn) == 0 {
		fn = flagf
	}

	if len(fn) != 0 {
		fileStorage, err := NewFileStorage(fn)

		if err != nil {
			log.Fatalf("unable to create file storage: %v", err)
		}
		return fileStorage, err
	} else {
		fileStorage := NewMemStorage()
		return fileStorage, nil
	}

}

func main() {
	baseURL := flag.String("b", "http://localhost:8080", "BASE_URL из cl")
	severAddress := flag.String("a", ":8080", "SERVER_ADDRESS из cl")
	fileStoragePath := flag.String("f", "store.json", "FILE_STORAGE_PATH из cl")
	flag.Parse()

	fileStorage, err := File(*fileStoragePath)
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
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Timeout(3 * time.Second))
	//compressor := middleware.NewCompressor(flate.DefaultCompression)
	//r.Use(compressor.Handler)
	r.Use(middleware.Compress(5))

	//v := middleware.Compress(5)
	//r.Use(v)
	r.Post("/", getURLForCut(fileStorage, *baseURL))
	r.Get("/{id}", redirectTo(fileStorage))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(fileStorage, *baseURL))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
