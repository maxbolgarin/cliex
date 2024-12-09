# cliex

[![Go Version][version-img]][doc] [![GoDoc][doc-img]][doc] [![Build][ci-img]][ci] [![GoReport][report-img]][report]


The **cliex** package is a robust and extensible HTTP client designed for making HTTP requests in Go applications, leveraging the capabilities of the `resty` library. It simplifies the process of sending HTTP requests and provides advanced features such as configuration options, circuit breaking, retries, and more.

This package is a wrapper on [resty](https://github.com/go-resty/resty) — it makes HTTP requests less verbose, adds request based retry settings and a circuit breaker.

## Table of Contents

1. [Introduction](#introduction)
2. [Features](#features)
3. [Installation](#installation)
4. [Usage](#usage)
   - [Initialization](#initialization)
   - [Using HTTPSet for Multiple Clients](#using-httpset-for-multiple-clients)
   - [Handling Broken Clients](#handling-broken-clients)
5. [Configuration Options](#configuration-options)
6. [Request Options](#request-options)
7. [Contributing](#contributing)
8. [License](#license)

## Features

- **Customizable HTTP Client:** Configure base URL, user agent, proxy, authentication tokens, and more.
- **Circuit Breaker Support:** Protects your application from cascading failures by using circuit breakers for HTTP requests.
- **Requests with Retries:** Automatically retries failed requests with exponential backoff.
- **Logging Support:** Integrates with custom loggers for request and error logging.
- **HTTP Client Sets:** Manage groups of HTTP clients, enabling requests to multiple endpoints with error handling for broken clients.
- **Flexible Request Options:** Supports custom headers, query parameters, form data, body payloads, and more.
- **SSL/TLS Configuration:** Allows configuration of CA files, client certificates, and insecure requests.
- **Debug Mode:** Enables additional logging for request troubleshooting.

## Installation

To install the package, use:

```bash
go get -u github.com/maxbolgarin/cliex
```

## Usage

### Initialization

Create an HTTP client using the provided configuration options.

```go
package main

import (
	"context"
	"log"
	"github.com/maxbolgarin/cliex"
)

func main() {
	// Creating a new HTTP client with default settings
	client := cliex.New(
		cliex.WithBaseURL("https://api.example.com"),
		cliex.WithUserAgent("MyApp/1.0"),
		cliex.WithAuthToken("Bearer YOUR_TOKEN"),
		cliex.WithRequestTimeout(10*time.Second),
		cliex.WithDebug(true),
	)

	// Making a GET request
	resp, err := client.Get(context.Background(), "/endpoint")
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Response: %s", resp.String())
}
```

### Using HTTPSet for Multiple Clients

Create a set of HTTP clients and perform operations on them collectively.

```go
package main

import (
	"context"
	"log"
	"github.com/maxbolgarin/cliex"
)

func main() {
	clientSet := cliex.NewSetFromConfigs(
		cliex.Config{BaseURL: "https://service1.example.com"},
		cliex.Config{BaseURL: "https://service2.example.com"},
	)

	resp, err := clientSet.Get(context.Background(), "/shared-resource")
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	for _, r := range resp {
		log.Printf("Response from client: %s", r.String())
	}
}
```

### Handling Broken Clients

You can manage failing clients within a set and choose to retry or handle them separately.

```go
package main

import (
	"context"
	"log"
	"github.com/maxbolgarin/cliex"
)

func main() {
	clientSet, err := cliex.NewSetFromConfigs(
		cliex.Config{BaseURL: "https://service1.example.com"},
		cliex.Config{BaseURL: "https://service2.example.com"},
	)
	if err != nil {
		log.Fatalf("Error creating client set: %v", err)
	}

	_, err = clientSet.Get(context.Background(), "/resource")
	if err != nil {
		log.Printf("Request failed: %v", err)
		brokenClients := clientSet.GetBroken()

		// Retry or handle broken clients
		for _, index := range brokenClients {
			log.Printf("Broken client index: %d", index)
		}
	}
}
```

## Configuration Options

- `BaseURL`: Sets the base URL for HTTP requests.
- `UserAgent`: Sets the User-Agent header for each request.
- `AuthToken`: Provides an Authorization header with a bearer token.
- `ProxyAddress`: Defines a proxy server for sending requests.
- `RequestTimeout`: Configures the maximum amount of time to wait for a request.
- `CAFiles`: Loads CA certificates for SSL validation.
- `ClientCertFile`/`ClientKeyFile`: Client-side certificate and key for TLS.
- `Insecure`: Allows insecure SSL connections.
- `Debug`: Enables detailed logging.
- `CircuitBreaker`: Activates the circuit breaker feature.

## Request Options

| Option                  | Description                                                                                              | Type                          |
|-------------------------|----------------------------------------------------------------------------------------------------------|-------------------------------|
| `Method`                | The HTTP method to use (e.g., GET, POST, PUT, DELETE).                                                   | `string`                      |
| `Headers`               | A map of header keys and values to include in the request.                                               | `map[string]string`           |
| `Query`                 | A map of query string parameters and their values.                                                       | `map[string]string`           |
| `PathParams`            | Path parameters for the request URL (e.g., `/v1/users/{userId}`).                                        | `map[string]string`           |
| `Cookies`               | Cookies to include in the request.                                                                       | `[]*http.Cookie`              |
| `FormData`              | Form data to include when submitting a form.                                                             | `map[string]string`           |
| `Files`                 | Files to upload, where the key is the file name and the value is the file path.                          | `map[string]string`           |
| `AuthToken`             | Authentication token for the request.                                                                    | `string`                      |
| `BasicAuthUser`         | Username for basic authentication.                                                                       | `string`                      |
| `BasicAuthPass`         | Password for basic authentication.                                                                       | `string`                      |
| `ForceContentType`      | Specifies a custom content type to parse the response (e.g., `application/json`).                         | `string`                      |
| `Body`                  | The body of the request, can be any type.                                                                | `any`                         |
| `Result`                | A variable to store the response body.                                                                   | `any`                         |
| `OutputPath`            | File path to save the response output.                                                                   | `string`                      |
| `RequestName`           | Name of the request for logging purposes.                                                                | `string`                      |
| `RetryCount`            | Number of times to retry the request if it fails.                                                        | `int`                         |
| `RetryWaitTime`         | Initial wait time between retries (default: 100 milliseconds).                                           | `time.Duration`               |
| `RetryMaxWaitTime`      | Maximum wait time between retries (default: 2 seconds).                                                  | `time.Duration`               |
| `InfiniteRetry`         | Whether to retry the request indefinitely.                                                               | `bool`                        |
| `RetryOnlyServerErrors` | Whether to retry only for server (5xx) errors.                                                           | `bool`                        |
| `NoLogRetryError`       | Whether to suppress logging of retry errors.                                                             | `bool`                        |
| `EnableTrace`           | Enable tracing of the request, accessible via `resp.Request.TraceInfo()`.                                | `bool`                        |


## Contributing

Contributions to improve the package are welcome. Please ensure any changes come with tests, and are validated against existing integrations. 

## License

This project is licensed under the terms of the [MIT License](LICENSE).

[MIT License]: LICENSE.txt
[version-img]: https://img.shields.io/badge/Go-%3E%3D%201.22-%23007d9c
[doc-img]: https://pkg.go.dev/badge/github.com/maxbolgarin/cliex
[doc]: https://pkg.go.dev/github.com/maxbolgarin/cliex
[ci-img]: https://github.com/maxbolgarin/cliex/actions/workflows/go.yaml/badge.svg
[ci]: https://github.com/maxbolgarin/cliex/actions
[report-img]: https://goreportcard.com/badge/github.com/maxbolgarin/cliex
[report]: https://goreportcard.com/report/github.com/maxbolgarin/cliex
