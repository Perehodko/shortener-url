package main

import (
	"github.com/Perehodko/shortener-url/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
	"time"
)

var storageURLs = make(map[string]string)

func getURLForCut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	// читаем Body
	defer r.Body.Close()
	bodyData, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//urlForCuts := linkFromBody
	urlForCuts := string(bodyData)
	getHost := r.Host

	shortLink := utils.GenerateRandomString()
	shortURL := "http://" + getHost + "/" + shortLink

	//записываем в мапу пару shortLink:оригинальная ссылка
	storageURLs[shortLink] = urlForCuts

	w.Write([]byte(shortURL))
}

func notFoundFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found"))
}

func redirectTo(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")

	initialURL := storageURLs[shortURL]
	if initialURL == "" {
		http.Error(w, "URl not in storage", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Location", initialURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(3 * time.Second))

	r.Post("/", getURLForCut)
	r.Get("/{id}", redirectTo)
	r.Get("/", notFoundFunc)

	log.Fatal(http.ListenAndServe(":8080", r))
}
