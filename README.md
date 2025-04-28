Fakes
======

More complex applications tend to require communicating with a plethora of different services. It's not always
possible to write tests that exercise the APIs you depend upon, but this lib is designed to make defining fake
test fixtures easier and more ergonomic for developers.

This lib removes the complexity of lifecycle management around the `httptest` server and provides some niceties 
that make fleshing out fakes far more straightforward. This means your acceptance/integration tests aren't awash
with various HTTP setup code and you can focus on what really matters.

This is ideal if you need to fake out multiple downstream dependencies whilst
keeping your acceptance tests readable and succinct. 

## Getting Started

This lib is somewhat opinionated such that writing happy path test cases can be done in as few lines as possible.

```go
import "github.com/elliotforbes/fakes"
```

You can then start defining in-memory fakes that have lifecycles as long as your test functions like so:

```go
downstreamAPI := fakes.New()
downstreamAPI.Endpoint(&fakes.Endpoint{
    Path: "/some/path/my/app/hits",
    Response: `{"status": "great success"}`,
})
downstreamAPI.Run(t)
```

Hitting this API will result in the following response:

```
HTTP GET /some/path/my/app/hits 

// returns a response:

HTTP 200 
Content-Type: application/json

{"status": "great success"}
```

## Specifying A Port

There are some instances where you need to specify a port upon which these fakes will run. To achieve this, you
can pass in the `fakes.WithPort(N)` optional parameter:

```go
downstreamAPI := fakes.New(fakes.WithPort(10000))
...
```



### Endpoints

Endpoints are a core concept for this lib. You can define 1 or more Endpoints for your fake service and you have 
full control of what happens in the event of this endpoint being called:

```go
downstreamAPI.Endpoint(&fakes.Endpoint{
    Path: "/some/path/my/app/hits",
    Response: `{"status": "great success"}`,
    // if ContentType is not specified, we assume `application/json`
    ContentType: "plaintext",
    // if statuscode is not specified, we assume a 200
    // status code response
    StatusCode: http.StatusUnauthorised,
    // if Methods is not specified, we assume it could
    // hit any HTTP method
    Methods: []string{
        http.MethodGet,
    },
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

## Stringing Your Setup Together

The `fakes` library includes the ability to string your setup together in
one succinct assertion. Let's take a look at an example of setting this up 
with 2 distinct endpoints:

```go
fakeServer := fakes.New().
    Endpoint(&fakes.Endpoint{
        Path:     "/",
        Response: "{}",
    }).
    Endpoint(&fakes.Endpoint{
        Path:     "/hello",
        Response: `{"message":"hello"}`,
    }).Run(t)
```

A nice bit of syntactic sugar for those that prefer this approach!

## Chaos Engineering

In your acceptance tests, it can be useful to model some level of unreliability
within your tests so that you are consistently handling retries and have a good
level of resiliency baked into your system.

This can invariably lead to flaky tests depending on how unlucky you are. You might
have clients that have backoff retries and a retry limit of 5, but somehow, you
roll 1's 5 times and all of these requests fail. 

I might eventually need a higher level of determinism within my acceptance tests, 
if that time comes, I can start to consider how I can expose controls that will mean
that subsequent requests pass.

If you'd like to include an element of chaos into your tests, you can specify the
`FailureRatePercent` field on each of your endpoints and it'll fail depending on that
provided percentage:

```go
fakeServer := fakes.New().
    Endpoint(&fakes.Endpoint{
        Path:               "/",
        Response:           "{}",
        // 100 means this will always fail
        FailureRatePercent: 100,
        // We can specify how we'd like the failure
        // to respond
        FailureHandler: func(c *gin.Context) {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "something bad happened",
            })
        },
    }).Run(t)
```

#### Controlling Chaos

Whilst chaos is a great way of testing retry limits on your tests, you do also
need to have some level of determinism to ensure you're not being plagued by flaky tests.

```go
fakeServer := fakes.New().
    Endpoint(&fakes.Endpoint{
        Path:               "/",
        Response:           "{}",
        // 100 means this will always fail
        FailureRatePercent: 100,

        // defines how many times this endpoint
        // can fail before it reverts to a happy
        // path
        MaxFailureCount: 5,
        // We can specify how we'd like the failure
        // to respond
        FailureHandler: func(c *gin.Context) {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "something bad happened",
            })
        },
    }).Run(t)