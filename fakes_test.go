package fakes_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elliotforbes/fakes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFakes(t *testing.T) {

	t.Run("we can run an in-memory test fake", func(t *testing.T) {
		fakeServer := fakes.New()
		fakeServer.Endpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(http.MethodGet, fakeServer.BaseURL, nil)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)
		//nolint
		defer response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	})

	t.Run("test the new BaseURL field on the fakeServer struct", func(t *testing.T) {
		fakeServer := fakes.New()
		fakeServer.Endpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
		})
		fakeServer.Run(t)
		request, err := http.NewRequest(http.MethodGet, fakeServer.BaseURL, nil)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)
		//nolint
		defer response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("test path parameters work as expected", func(t *testing.T) {
		fakeServer := fakes.New()
		fakeServer.Endpoint(&fakes.Endpoint{
			Path:     "/:id",
			Response: "{}",
			Expectation: func(r *http.Request) {
				id := strings.TrimPrefix(r.URL.Path, "/")
				assert.Equal(t, "some-id", id)
			},
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/some-id", fakeServer.BaseURL),
			nil,
		)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)
		//nolint
		defer response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("test we can set headers", func(t *testing.T) {
		fakeServer := fakes.New()
		fakeServer.Endpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
			Headers: fakes.Headers{
				"Authorization": "Bearer some-bearer",
			},
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(
			http.MethodGet,
			fakeServer.BaseURL,
			nil,
		)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)
		//nolint
		defer response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "Bearer some-bearer", response.Header["Authorization"][0])
	})

	t.Run("test we can use the override handler", func(t *testing.T) {
		fakeServer := fakes.New()
		fakeServer.Endpoint(&fakes.Endpoint{
			Path: "/",
			Handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"hello": "world"})
			},
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(
			http.MethodGet,
			fakeServer.BaseURL,
			nil,
		)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)
		//nolint
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `{"hello":"world"}`, string(body))
	})

	t.Run("test that the fake lib only sets up explicit paths", func(t *testing.T) {
		fakeServer := fakes.New()
		fakeServer.Endpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
			Methods:  []string{http.MethodGet},
			Headers: fakes.Headers{
				"Authorization": "Bearer some-bearer",
			},
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(
			http.MethodGet,
			fakeServer.BaseURL,
			nil,
		)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)
		//nolint
		defer response.Body.Close()

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "Bearer some-bearer", response.Header["Authorization"][0])

		request, err = http.NewRequest(
			http.MethodPost,
			fakeServer.BaseURL,
			nil,
		)
		assert.Nil(t, err)

		response, err = http.DefaultClient.Do(request)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	})

	t.Run("test that we can specify a distinct port", func(t *testing.T) {
		fakeServer := fakes.New(fakes.WithPort(10000))
		fakeServer.Endpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
			Methods:  []string{http.MethodGet},
			Headers: fakes.Headers{
				"Authorization": "Bearer some-bearer",
			},
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(
			http.MethodGet,
			"http://localhost:10000",
			nil,
		)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)
		//nolint
		defer response.Body.Close()
	})
}
