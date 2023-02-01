package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotFoundFunc(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test 1: check status code 404, response body and header contentType",
			want: want{
				code:        http.StatusNotFound,
				response:    "Not found",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodGet, "/", nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(notFoundFunc)
			h.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			assert.Equal(t, tt.want.code, w.Code, "Expected status code must be equal to received")

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want.response, string(resBody),
				"Expected text response must be equal to received")

			// заголовок ответа
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"),
				"Expected header must be equal to received")
		})
	}
}
