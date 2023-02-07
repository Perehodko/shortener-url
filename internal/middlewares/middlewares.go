package middlewares

import (
	"compress/gzip"
	"net/http"
)

func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// не забывайте потом закрыть *gzip.Reader
			defer gz.Close()

			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
