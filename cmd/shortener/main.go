package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/mux"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var storageURLs = make(map[string]string)

func getURLForCut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// читаем Body
	defer r.Body.Close()
	bodyData, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	//удаляем лишние скобки
	linkFromBody := strings.ReplaceAll(string(bodyData), "{", "")
	linkFromBody = strings.ReplaceAll(linkFromBody, "}", "")

	urlForCuts := linkFromBody
	getHost := r.Host

	shortLink := shorting()
	shortURL := "http://" + getHost + "/" + shortLink
	//fmt.Println("shortURL", shortURL, "gettingHost", getHost)

	//записываем в мапу пару shortLink:оригинальная ссылка
	storageURLs[shortLink] = urlForCuts

	w.Write([]byte(shortURL))

}

func notFoundFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}

func redirectTo(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	shortURL := vars["id"]

	initialURL := storageURLs[shortURL]

	fmt.Println("initialURL", initialURL)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", initialURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func shorting() string {
	b := make([]byte, 5)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(3 * time.Second))

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", getURLForCut)
	r.Get("/", notFoundFunc)
	r.Get("/{id}", redirectTo)

	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
