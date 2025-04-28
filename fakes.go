package fakes

import (
	"fmt"
	"math/rand"
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

	Port    int
	BaseURL string
}

// NewFakeHTTP - a constructor that spins up
// a new reference to a FakeService.
func New(opts ...func(*FakeService)) *FakeService {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	fakeService := &FakeService{
		router:     router,
		testserver: httptest.NewUnstartedServer(router),
	}

	for _, o := range opts {
		o(fakeService)
	}

	return fakeService
}

func WithPort(port int) func(*FakeService) {
	return func(f *FakeService) {
		f.Port = port
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
	Headers     Headers
	Handler     func(*gin.Context)
	// FailureRatePercent - allows you to specify the probability
	// of failure for your Endpoint. I.e. 0.8 represents and 80%
	// chance you'll be met with a 500 response.
	FailureRatePercent int
	// FailureHandler - allows you to define a custom failure
	// handler so that you can model how your upstream dependencies
	// could fail.
	FailureHandler func(*gin.Context)
	// MaxFailureCount - the maximum number of times chaos
	// can ensue within these fakes. Defaults to 3
	MaxFailureCount int

	// Expectation - it can be handy to specify assertions
	// in the context of the tests you are developing. This
	// will allow you to make assertions on the request that
	// eventually makes it to your fake.
	Expectation func(*http.Request)

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

// Endpoint - registers a new endpoint on the fake service.
// This will set some sensible defaults should the Endpoint not have
// them explicitly defined.
// For example, we default to HTTP 200 statuses and a Content-Type of
// 'application/json'.
// Whenever said endpoint is called, we ensure that we record the call
// and increment the `calls` field.
func (f *FakeService) Endpoint(e *Endpoint) *FakeService {
	f.Endpoints = append(f.Endpoints, e)
	// sensible default
	e.MaxFailureCount = 3

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
		e.recordCall()

		// We only want to return errors up to a point, this
		// will help keep a level of determinism within our
		// acceptance test setups and prevent flaky tests.
		if shouldReturnError(e.FailureRatePercent) &&
			e.calls <= e.MaxFailureCount-1 {
			e.FailureHandler(c)
			return
		}

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

		for header, value := range e.Headers {
			fmt.Println(header)
			fmt.Println(value)
			c.Header(header, value)
		}

		if e.Handler != nil {
			e.Handler(c)
			return
		}

		c.Render(status, render.Data{
			ContentType: e.ContentType,
			Data:        []byte(e.Response),
		})
	})

	return f
}

// shouldReturnError - given the endpoint's failure
// rate - calculates a random int and then compares that
// against the failureRatePercent.
//
// if the failureRatePercent is 20, that represents a 20% chance
// of returning an error.
func shouldReturnError(failureRatePercent int) bool {
	return failureRatePercent > rand.Intn(100)
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
func (f *FakeService) Run(t *testing.T) *FakeService {
	t.Log("Fake Service Starting up...")
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", f.Port))
	if err != nil {
		t.Errorf("Failed to listen: %s", err.Error())
		return f
	}
	err = f.testserver.Listener.Close()
	if err != nil {
		t.Errorf("Failed to close the testserver listener: %s", err.Error())
		return f
	}
	f.testserver.Listener = l
	f.testserver.Start()
	f.BaseURL = f.testserver.URL
	t.Logf("Fake Service Successfully Started: %s", f.testserver.Listener.Addr())

	return f
}
