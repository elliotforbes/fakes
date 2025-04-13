package fakes

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type Endpoint struct {
	Path        string
	Response    string
	StatusCode  int
	ContentType string
	Expectation func(*http.Request)

	calls int
	mutex sync.Mutex
}

func (e *Endpoint) recordCall() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.calls++
}

type FakeService struct {
	router     *gin.Engine
	testserver *httptest.Server
	Endpoints  []*Endpoint

	BaseURL string
}

func NewFakeHTTP() *FakeService {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return &FakeService{
		router:     router,
		testserver: httptest.NewUnstartedServer(router),
	}
}

func (f *FakeService) AddEndpoint(e *Endpoint) {
	f.Endpoints = append(f.Endpoints, e)
	f.router.Any(e.Path, func(c *gin.Context) {
		// If there are specific expectations attached
		// to a given endpoint, run through these expectations now.
		if e.Expectation != nil {
			e.Expectation(c.Request)
		}

		if e.ContentType != "" {
			c.Header("Content-Type", e.ContentType)
		}

		status := e.StatusCode
		if status == 0 {
			status = http.StatusOK
		}
		fmt.Printf("%s: %s - HTTP %d\n%s", c.Request.Method, c.Request.URL, status, e.Response)
		e.recordCall()

		c.String(status, e.Response)
	})
}

func (f *FakeService) TidyUp(t *testing.T) {
	t.Log("FakeService tidyup...")
	for _, e := range f.Endpoints {
		assert.GreaterOrEqual(t, e.calls, 1, "endpoint %s has not been called within this test")
	}
	f.testserver.Close()
}

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
