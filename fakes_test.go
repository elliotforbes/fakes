package fakes_test

import (
	"net/http"
	"testing"

	"github.com/elliotforbes/fakes"
	"github.com/stretchr/testify/assert"
)

func TestFakes(t *testing.T) {

	t.Run("we can run an in-memory test fake", func(t *testing.T) {
		fakeServer := fakes.NewFakeHTTP("8080")
		fakeServer.AddEndpoint(&fakes.Endpoint{
			Path:     "/",
			Response: "{}",
		})
		fakeServer.Run(t)

		request, err := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
		assert.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)

	})
}
