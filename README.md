Fakes
======

More complex applications tend to require communicating with a plethora of different services. It's not always
possible to write tests that exercise the APIs you depend upon, but this lib is designed to make defining fake
test fixtures easier and more ergonomic for developers.

This lib removes the complexity of lifecycle management around the `httptest` server and provides some niceties 
that make fleshing out fakes far more straightforward. This means your acceptance/integration tests aren't awash
with various HTTP setup code and you can focus on what really matters.

## Getting Started

This lib is somewhat opinionated such that writing happy path test cases can be done in as few lines as possible.

```go
import "github.com/elliotforbes/fakes"
```

You can then start defining in-memory fakes that have lifecycles as long as your test functions like so:

```go
downstreamAPI := fakes.NewFakeHTTP()
downstreamAPI.AddEndpoint(&fakes.Endpoint{
    Path: "/some/path/my/app/hits",
    Response: `{"status": "great success"}`,
})
downstreamAPI.Run(t)
```

Hitting this API will result in the following response:

```
HTTP 200 
Content-Type: application/json

{"status": "great success"}
```

### Endpoints

Endpoints are a core concept for this lib. You can define 1 or more Endpoints for your fake service and you have 
full control of what happens in the event of this endpoint being called:

```go
downstreamAPI.AddEndpoint(&fakes.Endpoint{
    Path: "/some/path/my/app/hits",
    Response: `{"status": "great success"}`,
    ContentType: "plaintext",
    StatusCode: http.StatusUnauthorised,
    Headers: fakes.Headers{
		"Authorization": "Bearer some-bearer",
    },
    Expectation: func(r *http.Request) {
        // run assertions on the incoming http request to ensure
        // that you are sending the right data to any of the APIs
        // that you depend upon.
    },
})
```

> Note: if you define an endpoint for your test, it will need to be hit at least once in order for the tests to pass. This is
to help prevent unused endpoints from being defined within your tests.

