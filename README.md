# testpilot

`testpilot` is a Go package that provides a testing framework for making HTTP requests and validating responses. 
It is designed to help create test plans and perform assertions on HTTP requests and responses.

## Features

- Create and manage test plans for HTTP testing.
- Perform HTTP requests with customizable methods, URLs, headers, and body content.
- Validate HTTP responses with assertions on status codes, headers, response body, and JSON paths.
- Cache and retrieve responses for later use in tests.

## Installation
To install `testpilot`, run the following command:
```bash
go get github.com/trapajim/testpilot
```

## Usage
### Creating a Test Plan
Create a new test plan using the NewPlan function:
```go
func TestYourFunction(t *testing.T) {
    p := NewPlan(t, "test")
    p.Request("GET", "https://api.sampleapis.com/futurama/episodes").
    Expect().
    Status(200).
    Header("Content-Type", Equal("application/json")).
    Body(AssertPath(".0.id", func(val int) error {
        if val != 1 {
            return errors.New("expected 1 got " + strconv.Itoa(val))
        }
        return nil
    }))
    p.Run()
}
```

### Validating Response Headers
You can validate response headers using the `Header` method with composable assertions:
```go
func TestYourFunction(t *testing.T) {
    p := NewPlan(t, "test")
    p.Request("GET", "https://api.sampleapis.com/futurama/episodes").
    Expect().
    Status(200).
    Header("Content-Type", Equal("application/json")).  // check exact value
    Header("X-Request-Id", Exists())                    // check header exists
    p.Run()
}
```

### Adding a body and headers to the request

You can add a body and headers to the request using the Body and Headers methods:
```go
type Episode struct {
    Id int `json:"id"`
    Title string `json:"title"`
}
func TestYourFunction(t *testing.T) {
    newEpisode := Episode{
        Id: 1,
        Title: "Space Pilot 3000",
    }
    p := NewPlan(t, "test")
    p.Request("POST", "https://api.sampleapis.com/futurama/episodes").
    Body(JSON(newEpisode)).
    Headers(map[string]string{
        "Content-Type": "application/json",
    }).
    Expect().
    Status(201)
    p.Run()
}
```

### Utilizing Placeholder
You can utilize placeholders in the URL of the request. Placeholders are defined using curly braces `{}` in the URL and should be a json path. The value of the placeholder is extracted from the response of the previous request.
```go
func TestYourFunction(t *testing.T) {
    p := NewPlan(t, "test")
    p.Request("GET", "https://api.sampleapis.com/futurama/episodes").
    Expect().
    Status(200)
    // add a second request that uses the response of the first request as a placeholder
    p.Request("GET", "https://api.sampleapis.com/futurama/episodes/{.0.id}"). // the placeholder is the id of the first episode
    Expect().
    Status(200)
    p.Run()
}
```

### Storing and Retrieving Responses
You can store the response of a request and retrieve it later in the test plan. This is useful when you want to use the response of a request in multiple assertions.
```go
func TestYourFunction(t *testing.T) {
    p := NewPlan(t, "test")
    p.Request("GET", "https://api.sampleapis.com/futurama/episodes").
    Store("episodes"). // store the response of this request
    Expect().
    Status(200).
    
    // add a second request that uses the stored response
    p.Request("GET", "https://api.sampleapis.com/futurama/episodes/{.0.id}"). // the placeholder is the id of the first episode
    Expect().
    Status(200).

    // add a third request, this time using the stored response
    p.Request("GET", "https://api.sampleapis.com/futurama/episodes/{episodes.1.id}"). // starting the placeholder with the name of the stored response 
    Expect().
    Status(200)

	// run the test plan
    p.Run()
}
```

Responses can also be used in `Body` method with `p.Response()` or `p.ResponseForKey("name")`:

See tests for more comprehensive examples.

## Contributing
If you find a bug or want to suggest a new feature for testpilot, please open an issue on GitHub or submit a pull request. We welcome contributions from the community.