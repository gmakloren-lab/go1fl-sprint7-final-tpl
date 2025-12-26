package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

// TestCafeSearch проверяет, что параметр search возвращает только те кафе,
// в названии которых содержится указанная подстрока (без учёта регистра).
func TestCafeSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mainHandle))
	defer server.Close()

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, tt := range requests {
		url := fmt.Sprintf(
			"%s/cafe?city=moscow&search=%s",
			server.URL,
			tt.search,
		)

		resp, err := http.Get(url)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		result := strings.TrimSpace(string(body))

		var cafes []string
		if result != "" {
			cafes = strings.Split(result, ", ")
		}

		assert.Equal(t, tt.wantCount, len(cafes))

		for _, cafe := range cafes {
			assert.True(
				t,
				strings.Contains(
					strings.ToLower(cafe),
					strings.ToLower(tt.search),
				),
			)
		}
	}
}

// TestCafeCount проверяет, что параметр count корректно ограничивает
// количество возвращаемых кафе.
func TestCafeCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(mainHandle))
	defer server.Close()

	city := "moscow"

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, len(cafeList[city])},
	}

	for _, tt := range requests {
		url := fmt.Sprintf("%s/cafe?city=%s&count=%d", server.URL, city, tt.count)
		resp, err := http.Get(url)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		result := strings.TrimSpace(string(body))

		var cafes []string
		if result != "" {
			cafes = strings.Split(result, ", ")
		}

		assert.Equal(t, tt.want, len(cafes))
	}
}
