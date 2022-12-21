package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
)

var storageURLs = make(map[string]string)

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get called"}`))
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// читаем Body
	bodeData, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	//фигачим в json
	var dat map[string]interface{}
	if err := json.Unmarshal(bodeData, &dat); err != nil {
		panic(err)
	}
	//fmt.Println(dat["URL"])

	urlForCuts := dat["URL"].(string)
	getHost := r.Host

	//resForCut := strings.ReplaceAll(urlForCuts, "http://"+getHost+"/", "")
	//w.Write([]byte(resForCut))
	shotrLink := shorting()
	fmt.Println(shotrLink)
	shortURL := "http://" + getHost + "/" + shotrLink
	fmt.Println(shortURL)

	storageURLs[shotrLink] = urlForCuts
	//fmt.Println(storageURLs)

	w.Write([]byte(shortURL))

}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}

func redirectTo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTemporaryRedirect)
	//var link string
	vars := mux.Vars(r)
	shortURl := vars["id"]
	initialURL := storageURLs[shortURl]

	//fmt.Println(storageURLs)
	//fmt.Println("initialURL:", initialURL)

	http.Redirect(w, r, initialURL, http.StatusSeeOther)
	w.Header().Set("Location", initialURL)

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
	//r.HandleFunc("/", get).Methods(http.MethodGet)
	r.HandleFunc("/", post).Methods(http.MethodPost)
	r.HandleFunc("/", notFound)

	r.HandleFunc("/{id}", redirectTo).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
