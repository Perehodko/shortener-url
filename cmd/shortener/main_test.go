package main

import (
	"github.com/Perehodko/shortener-url/internal/utils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetURLForCut(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		bodyLen     int
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		// определяем тесты
		{
			name: "test 1: checking Content-Type header, status code is 201 and len(body)>0",
			want: want{
				contentType: "application/json",
				bodyLen:     0,
				code:        http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := strings.NewReader(`http://privet.com/lalalala`)
			request := httptest.NewRequest(http.MethodPost, "/", bodyReader)

			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			h := http.HandlerFunc(getURLForCut)
			// запускаем сервер
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
			assert.Greater(t, len(resBody), tt.want.bodyLen)

			// заголовок ответа
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"),
				"Expected header must be equal to received")
		})
	}
}

func TestShorting(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "test 1: len of return function >0",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := utils.GenerateRandomString()
			assert.Greater(t, len(gotResult), tt.want)
		})
	}
}

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
				response:    `{"message": "not found"}`,
				contentType: "application/json",
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
				"Expected status code must be equal to received")

			// заголовок ответа
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"),
				"Expected header must be equal to received")
		})
	}
}

func TestRedirectTo(t *testing.T) {
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
			name: "test 1: negative. check status code 400 if short URL not in storage",
			want: want{
				code:        http.StatusBadRequest,
				response:    "URl not in storage\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(redirectTo)
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
			assert.Equal(t, tt.want.response, string(resBody))

			// заголовок ответа
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"),
				"Expected header must be equal to received")
		})
	}
}
