package main

import (
	"github.com/Perehodko/shortener-url/internal/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestNewStructRedirectTo(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
		target      string
		shortLink   string
		expectURL   string
		urlForCut   string
	}
	cases := []struct {
		name string
		want want
	}{
		{
			name: "test 1: checking Content-Type header, status code is 400 and message",
			want: want{
				contentType: "text/plain; charset=utf-8",
				response:    "in map no shortURL from request\n",
				code:        http.StatusBadRequest,
				target:      "/123",
				shortLink:   "",
				urlForCut:   "",
				expectURL:   "",
			},
		},
		{
			name: "test 2: checking Content-Type header, status code is 400 and URL from store",
			want: want{
				contentType: "text/plain; charset=utf-8",
				response:    "in map no shortURL from request\n",
				code:        http.StatusBadRequest,
				target:      "/xyz",
				shortLink:   "xyz",
				urlForCut:   "https://music.yandex.ru/artist/421792/tracks",
				expectURL:   "https://music.yandex.ru/artist/421792/tracks",
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := &newStruct{
				st: storage.NewURLStore(),
			}
			request := httptest.NewRequest(http.MethodGet, tt.want.target, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.redirectTo)

			h.ServeHTTP(w, request)

			res := w.Result()

			if tt.want.shortLink != "" {
				//prepare real storage
				s.st.PutURLInStorage(tt.want.shortLink, tt.want.urlForCut)
				shortURLFromService, err := s.st.GetURLFromStorage(tt.want.shortLink)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.want.expectURL, shortURLFromService,
					"Expected URL from storage must be equal to received")
			}

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

func TestNewStructGetURLForCut(t *testing.T) {
	type want struct {
		code        int
		bodyLen     int
		contentType string
	}
	cases := []struct {
		name string
		want want
	}{
		//определяем тесты
		{
			name: "test 1: checking Content-Type header, status code is 201 and len(body)>0",
			want: want{
				contentType: "text/plain; charset=utf-8",
				bodyLen:     0,
				code:        http.StatusCreated,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := strings.NewReader(`http://privet.com/lalalala`)
			request := httptest.NewRequest(http.MethodPost, "/", bodyReader)

			s := &newStruct{
				st: storage.NewURLStore(),
			}

			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			h := http.HandlerFunc(s.getURLForCut)
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
