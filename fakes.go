package fakes

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/stretchr/testify/assert"
)

// FakeService - the core of this lib.
// features a ref to a gin.Engine and a
// slice of Endpoints.
// BaseURL is an exposed field which is
// set at the point at which the Run method
// is called.
type FakeService struct {
	router     *gin.Engine
	testserver *httptest.Server
	Endpoints  []*Endpoint

	BaseURL string
}

// NewFakeHTTP - a constructor that spins up
// a new reference to a FakeService.
func NewFakeHTTP() *FakeService {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return &FakeService{
		router:     router,
		testserver: httptest.NewUnstartedServer(router),
	}
}

type Headers map[string]string

// Endpoint - represents an Endpoint defined
// under the context of a FakeService.
type Endpoint struct {
	Path        string
	Response    string
	StatusCode  int
	Methods     []string
	ContentType string
	Expectation func(*http.Request)
	Headers     Headers

	calls int
	mutex sync.Mutex
}

// recordCall - a threadsafe method that safely
// increments the `calls` field on the endpoint.
func (e *Endpoint) recordCall() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.calls++
}

// AddEndpoint - registers a new endpoint on the fake service.
// This will set some sensible defaults should the Endpoint not have
// them explicitly defined.
// For example, we default to HTTP 200 statuses and a Content-Type of
// 'application/json'.
// Whenever said endpoint is called, we ensure that we record the call
// and increment the `calls` field.
func (f *FakeService) AddEndpoint(e *Endpoint) {
	f.Endpoints = append(f.Endpoints, e)

	// if the user of the lib doesn't explicitly set the
	// methods on the Endpoint, we assume that we can match any
	if len(e.Methods) == 0 {
		e.Methods = []string{
			http.MethodGet,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPatch,
			http.MethodPost,
			http.MethodPut,
			http.MethodTrace,
			http.MethodConnect,
		}
	}

	f.router.Match(e.Methods, e.Path, func(c *gin.Context) {
		// If there are specific expectations attached
		// to a given endpoint, run through these expectations now.
		if e.Expectation != nil {
			e.Expectation(c.Request)
		}

		if e.ContentType != "" {
			c.Header("Content-Type", e.ContentType)
		} else {
			c.Header("Content-Type", "application/json")
		}

		status := e.StatusCode
		if status == 0 {
			status = http.StatusOK
		}
		fmt.Printf("%s: %s - HTTP %d\n%s\n", c.Request.Method, c.Request.URL, status, e.Response)
		e.recordCall()

		for header, value := range e.Headers {
			fmt.Println(header)
			fmt.Println(value)
			c.Header(header, value)
		}

		c.Render(status, render.Data{
			ContentType: e.ContentType,
			Data:        []byte(e.Response),
		})
	})
}

// TidyUp - this method ranges over all of the endpoints defined
// under this FakeService and ensures that each of them have been called
// at least once. If the call count is 0, then this will fail the test
// that depends on this fake service.
func (f *FakeService) TidyUp(t *testing.T) {
	t.Log("FakeService tidyup...")
	for _, e := range f.Endpoints {
		assert.GreaterOrEqual(t, e.calls, 1, "endpoint %s has not been called within this test")
	}
	f.testserver.Close()
}

// Run - starts up the fake service. This creates a custom net listener
// which then replaces the testserver listener. This was due to communication
// issues between docker containers originally, however, this argument may
// no longer hold water.
func (f *FakeService) Run(t *testing.T) {
	t.Log("Fake Service Starting up...")
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Errorf("Failed to listen: %s", err.Error())
		return
	}
	err = f.testserver.Listener.Close()
	if err != nil {
		t.Errorf("Failed to close the testserver listener: %s", err.Error())
		return
	}
	f.testserver.Listener = l
	f.testserver.Start()
	f.BaseURL = f.testserver.URL
	t.Logf("Fake Service Successfully Started: %s", f.testserver.Listener.Addr())

}
