Fakes
======

A Handy HTTP Fake library that helps you to build out acceptance tests.

## Features

- A Fake test server setup that's perfect for writing succinct and strong acceptance tests for your applications.
- A flexible assertion setup that allows you to perform strong assertions that downstream services are getting the right information in your requests.

## Example

The key benefit of this package is clear if you do a lot of acceptance tests prior to deploying your service to production.

This package allows you to spin up in-memory test servers within the tests themselves that your application can then hit and you can
do some assertions against.

```go
// Basic Example - you just want your application to be able to hit a fake service
t.Run("My awesome acceptance test", func(t *testing.T) {
    downstreamAPI := fakes.NewFakeHTTP("8000")
    downstreamAPI.AddEndpoint(fakes.Endpoint{
        Path: "/some/path/my/app/hits",
        Response: `{"status": "great success"}`,
    })
    downstreamAPI.Run(t)

    // make a HTTP request to the endpoint that hits our system
})

```