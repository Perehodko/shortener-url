package main

import (
	"github.com/Perehodko/shortener-url/cmd/shortener"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"time"
)

func main() {
	r := chi.NewRouter()

	// middleware для стабильности прилождения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(3 * time.Second))

	r.Post("/", shortener.GetURLForCut)
	r.Get("/{id}", shortener.RedirectTo)
	r.Get("/", shortener.NotFoundFunc)

	log.Fatal(http.ListenAndServe(":8080", r))
}
