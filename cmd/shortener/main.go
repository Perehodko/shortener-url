package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
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
	r := mux.NewRouter()
	r.HandleFunc("/", getURLForCut).Methods(http.MethodPost)
	r.HandleFunc("/", notFoundFunc).Methods(http.MethodGet)

	r.HandleFunc("/{id}", redirectTo).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
