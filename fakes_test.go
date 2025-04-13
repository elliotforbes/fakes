package fakes_test

import (
	"net/http"
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
}
