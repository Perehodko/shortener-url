package main

import (
	"encoding/json"
	"fmt"
	"github.com/Perehodko/shortener-url/internal/storage"
	"github.com/Perehodko/shortener-url/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
)

type newStruct struct {
	st storage.Storage
}

func (s *newStruct) getURLForCut(w http.ResponseWriter, r *http.Request) {
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
	err = s.st.PutURL(shortLink, urlForCuts)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	//fmt.Println(storage.URLStorage{})

	w.Write([]byte(shortURL))
}

func notFoundFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found"))
}

func (s *newStruct) redirectTo(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	initialURL, err := s.st.GetURL(shortURL)

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

type Url struct {
	URL string `json:"url"`
}

type Res struct {
	Result string `json:"result"`
}

func (s *newStruct) shorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	decoder := json.NewDecoder(r.Body)
	var u Url

	err := decoder.Decode(&u)
	if err != nil {
		panic(err)
	}
	log.Println(u.URL, "1!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!1")
	//получаю из хранилища результат
	//urlForCuts := linkFromBody
	urlForCuts := string(u.URL)
	getHost := r.Host

	shortLink := utils.GenerateRandomString()
	shortURL := "http://" + getHost + "/" + shortLink

	//записываем в мапу пару shortLink:оригинальная ссылка
	err = s.st.PutURL(shortLink, urlForCuts)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	//fmt.Println(storage.URLStorage{})

	tx := Res{Result: shortURL}
	// преобразуем tx в JSON-формат
	txBz, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}
	// txBz — это []byte, поэтому приводим его к типу string для печати
	fmt.Println(string(txBz), "RESULT!!!!!!!!!!!!!!!!!!!!!!!!")
	w.Write(txBz)

}

func main() {
	r := chi.NewRouter()

	n := newStruct{
		st: storage.NewURLStore(),
	}

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Timeout(3 * time.Second))

	r.Post("/", n.getURLForCut)
	r.Get("/{id}", n.redirectTo)
	r.Get("/", notFoundFunc)
	r.Post("/api/shorten", n.shorten)

	log.Fatal(http.ListenAndServe(":8080", r))
}
