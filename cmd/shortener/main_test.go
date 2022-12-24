package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_getURLforCut(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		// определяем все тесты
		{
			name: "test #1",
			want: want{
				code:        404,
				response:    `{"message": "not found"}`,
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// запускаем каждый тест
			request := httptest.NewRequest(http.MethodPost, "/", nil)

			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			h := http.HandlerFunc(notFoundFunc)
			// запускаем сервер
			h.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}

			// заголовок ответа
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestRedirectTo(t *testing.T) {
	type want struct {
		code     int
		response string
		location string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test #1",
			want: want{
				code:     404,
				response: `{"message": "not found"}`,
				location: "123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// запускаем каждый тест
			request := httptest.NewRequest(http.MethodGet, "/abcde", nil)

			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			h := http.HandlerFunc(notFoundFunc)
			// запускаем сервер
			h.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// заголовок ответа
			if res.Header.Get("Location") != tt.want.location {
				t.Errorf("Expected Location %s, got %s", tt.want.location, res.Header.Get("Location"))
			}
		})
	}
}
