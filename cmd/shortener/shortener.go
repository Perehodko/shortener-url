package shortener

import (
	"github.com/go-chi/chi/v5"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var storageURLs = make(map[string]string)

func GetURLForCut(w http.ResponseWriter, r *http.Request) {
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

	//записываем в мапу пару shortLink:оригинальная ссылка
	storageURLs[shortLink] = urlForCuts

	w.Write([]byte(shortURL))
}

func NotFoundFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}

func RedirectTo(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")

	initialURL := storageURLs[shortURL]
	if initialURL == "" {
		http.Error(w, "URl not in storage", http.StatusBadRequest)
		return
	}

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
