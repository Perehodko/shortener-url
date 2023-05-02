package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Perehodko/shortener-url/internal/dbstorage"
	"github.com/Perehodko/shortener-url/internal/middlewares"
	"github.com/Perehodko/shortener-url/internal/storage"
	"github.com/Perehodko/shortener-url/internal/utils"
	"github.com/Perehodko/shortener-url/internal/workwithcookie"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FileName      string `env:"FILE_STORAGE_PATH"`
	dbAddress     string `env:"DATABASE_DSN"`
}

var cfg Config

func getURLForCut(s storage.Storage, UUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		uid, err := workwithcookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = UUID
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

		//записываем в мапу s.URLs[UUID] = map[shortLink]urlForCuts{}
		err = s.PutURL(UUID, shortLink, urlForCuts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		workwithcookie.SetUUIDCookie(w, uid)
		w.WriteHeader(http.StatusCreated)
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

		uid, err := workwithcookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = UUID
		}

		initialURL, err := s.GetURL(UUID, shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		workwithcookie.SetUUIDCookie(w, uid)
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

		uid, err := workwithcookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = UUID
		}

		decoder := json.NewDecoder(r.Body)
		var u URLStruct

		err = decoder.Decode(&u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//получаю из хранилища результат
		urlForCuts := u.URL

		shortLink := utils.GenerateRandomString()
		shortURL := cfg.BaseURL + "/" + shortLink

		err = s.PutURL(UUID, shortLink, urlForCuts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tx := Res{Result: shortURL}
		// преобразуем tx в JSON-формат
		txBz, err := json.Marshal(tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		workwithcookie.SetUUIDCookie(w, uid)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
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

		uid, err := workwithcookie.ExtractUID(r.Cookies())
		if err != nil {
			http.Error(w, "no links", http.StatusNoContent)
			return
		}

		getUserURLs, err := s.GetUserURLs(UUID)
		if err != nil {
			http.Error(w, "internal error", http.StatusNoContent)
			return
		}
		if len(getUserURLs) == 0 {
			http.Error(w, "no links", http.StatusNoContent)
			return
		}

		// мапа для пар short_url:original_url из хранилища
		type M map[string]interface{}
		var myMapSlice []M

		for i, j := range getUserURLs {
			res := M{"short_url": cfg.BaseURL + "/" + i, "original_url": j}
			myMapSlice = append(myMapSlice, res)
		}
		//преобразуем в нужный формат
		myJSON, err := json.MarshalIndent(myMapSlice, "", "    ")
		if err != nil {
			http.Error(w, "no links", http.StatusNoContent)
			return
		}

		workwithcookie.SetUUIDCookie(w, uid)
		w.Header().Set("Content-Type", "application/json")
		w.Write(myJSON)
		w.WriteHeader(http.StatusOK)

	}
}

func PingDBPostgres(DBAddress string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("postgres", DBAddress)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		// check db
		if db.Ping() != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		defer db.Close()
		w.WriteHeader(http.StatusOK)
	}
}

type URLStructBatch struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLStructBatchResponse struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type News []URLStructBatchResponse

func batch(s storage.Storage, DBAddress, UUID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		//defer cancel()

		uid, err := workwithcookie.ExtractUID(r.Cookies())
		if err != nil {
			uid = UUID
		}
		// получить данные из тела запроса
		decoder := json.NewDecoder(r.Body)

		var u URLStructBatch
		//buffer size
		//size := 50
		store := make(map[string]string)

		// read open bracket
		decoder.Token()

		// while the array contains values
		for decoder.More() {
			// decode an array value (Message)
			err := decoder.Decode(&u)
			if err != nil {
				log.Fatal(err)
			}
			store[u.CorrelationId] = u.OriginalURL
			fmt.Printf("%v: %v\n", u.CorrelationId, u.CorrelationId)
			fmt.Println("store:", store)

			//if len(store) == size {
			//	err = s.PutURLsBatch(ctx, uid, store)
			//	//clear map
			//	store = make(map[string]string)
			//	if err != nil {
			//		http.Error(w, err.Error(), http.StatusBadRequest)
			//		return
			//	}
			//}
			//err = s.PutURLsBatch(ctx, uid, store) // вот тут сбросить оставшиеся данные в БД
		}
		// read closing bracket
		decoder.Token()

		var resp News

		for correlationId, _ := range store {
			//shortURL, err := s.GetURLByCorrelationId(correlationId)
			shortLink := utils.GenerateRandomString()

			resp = append(resp, URLStructBatchResponse{
				CorrelationId: correlationId,
				ShortURL:      cfg.BaseURL + "/" + shortLink,
			})
		}
		fmt.Println("resp!!!!", resp)

		tx := resp
		// преобразуем tx в JSON-формат
		txBz, err := json.Marshal(tx)
		//txBz, err := json.Marshal(map[string]interface{}{"correlation_id": "1", "short_url": "ldld"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		workwithcookie.SetUUIDCookie(w, uid)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(txBz)

	}
}

func main() {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "password"
		dbname   = "postgres"
		sslmode  = "disable"
	)

	//PSQLConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

	// получаем UUID
	UUID := uuid.New()
	UUIDStr := UUID.String()

	baseURL := flag.String("b", "http://localhost:8080", "BASE_URL из cl")
	severAddress := flag.String("a", ":8080", "SERVER_ADDRESS из cl")
	fileStoragePath := flag.String("f", "store.json", "FILE_STORAGE_PATH из cl")
	dbAddress := flag.String("d", "", "DATABASE_DSN")
	flag.Parse()

	// вставляем в структуру cfg значения из флагов
	cfg.ServerAddress = *severAddress
	cfg.BaseURL = *baseURL
	cfg.FileName = *fileStoragePath
	cfg.dbAddress = *dbAddress

	// перезатираем их значениями энвов
	// если значения в энве для поля структуры нет - то в поле останется значение из флага
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	var s storage.Storage
	if cfg.dbAddress != "" {
		s = dbstorage.NewDBStorage(*dbAddress)
	} else {
		s, err = NewStorage(cfg.FileName)
		if err != nil {
			log.Fatal(err)
		}
	}

	r := chi.NewRouter()

	ServerAddr := cfg.ServerAddress
	if len(ServerAddr) == 0 {
		ServerAddr = *severAddress
	}

	DBAddress := cfg.dbAddress
	if len(DBAddress) == 0 {
		DBAddress = *dbAddress
	}

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Compress(5),
		middlewares.Decompress)

	r.Post("/", getURLForCut(s, UUIDStr))
	r.Get("/{id}", redirectTo(s, UUIDStr))
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", shorten(s, UUIDStr))
	r.Get("/api/user/urls", getUserURLs(s, UUIDStr))
	r.Get("/ping", PingDBPostgres(DBAddress))
	r.Post("/api/shorten/batch", batch(s, DBAddress, UUIDStr))

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
