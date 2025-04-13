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
	port       string
	router     *gin.Engine
	testserver *httptest.Server
	Endpoints  []*Endpoint
}

func NewFakeHTTP(port string) *FakeService {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return &FakeService{
		port:       port,
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
	t.Logf("FakeService tidyup - port:%s", f.port)
	for _, e := range f.Endpoints {
		assert.GreaterOrEqual(t, e.calls, 1, "endpoint %s has not been called within this test")
	}
	f.testserver.Close()
}

func (f *FakeService) Run(t *testing.T) {
	t.Logf("Fake Service Starting Up on port: %s", f.port)
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", f.port))
	if err != nil {
		t.Errorf(fmt.Sprintf("Failed to listen: %s", err.Error()))
		return
	}
	err = f.testserver.Listener.Close()
	if err != nil {
		t.Errorf(fmt.Sprintf("Failed to close the testserver listener: %s", err.Error()))
		return
	}
	f.testserver.Listener = l
	f.testserver.Start()
	t.Log("Fake Service Successfully Started")

}
