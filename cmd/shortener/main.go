package main

import (
	"log"
	"net/http"
)

var storageURLs = make(map[string]string)

func getURLAndCut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusTemporaryRedirect)

		url := r.URL.RequestURI()
		w.Write([]byte(r.Host + storageURLs[url]))
		//w.Write([]byte(`lflflflf`))

		w.Header().Set("Location", r.Host+storageURLs[url])

	case "POST":
		w.WriteHeader(http.StatusCreated)
		urlForCuts := r.URL.RequestURI()

		if len(urlForCuts[1:]) >= 3 {
			cutURL := urlForCuts[0:3]
			storageURLs[cutURL] = urlForCuts
			w.Write([]byte(cutURL))

		} else {
			storageURLs[urlForCuts] = urlForCuts
			w.Write([]byte(urlForCuts))
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
	}
}

func returnLongURL(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		url := r.URL.RequestURI()
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte(storageURLs[url]))
	default:
		w.Write([]byte(r.Method))
	}
}

func main() {
	// маршрутизация запросов обработчику
	http.HandleFunc("/", getURLAndCut)
	//http.HandleFunc("/", returnLongURL)

	// конструируем свой сервер
	server := &http.Server{
		Addr: "localhost:8080",
	}

	log.Fatal(server.ListenAndServe())
}
