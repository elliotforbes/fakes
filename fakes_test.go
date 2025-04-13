package fakes_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/elliotforbes/fakes"
	"github.com/stretchr/testify/assert"
)

func TestFakes(t *testing.T) {

	t.Run("we can run an in-memory test fake", func(t *testing.T) {
		fakeServer := fakes.NewFakeHTTP()
		fakeServer.AddEndpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(http.MethodGet, fakeServer.BaseURL, nil)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("test the new BaseURL field on the fakeServer struct", func(t *testing.T) {
		fakeServer := fakes.NewFakeHTTP()
		fakeServer.AddEndpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
		})
		fakeServer.Run(t)
		t.Logf("BaseURL: %s\n", fakeServer.BaseURL)
		request, err := http.NewRequest(http.MethodGet, fakeServer.BaseURL, nil)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("test path parameters work as expected", func(t *testing.T) {
		fakeServer := fakes.NewFakeHTTP()
		fakeServer.AddEndpoint(&fakes.Endpoint{
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

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}
