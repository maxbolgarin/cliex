package cliex_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/maxbolgarin/cliex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTP_Req(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.Req(ctx, http.MethodGet, "/test", nil, &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_Get(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.Get(ctx, "/test", &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_GetQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.URL.Query().Get("q") != "abc" {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.GetQ(ctx, "/test", &responseBody, "q", "abc")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_Post(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestBody = struct {
		Key string `json:"key"`
	}{
		Key: "value",
	}

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodPost {
				return nil, cliex.ErrBadRequest
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var reqb struct {
				Key string `json:"key"`
			}
			err = json.Unmarshal(body, &reqb)
			if err != nil {
				return nil, err
			}
			if reqb.Key != requestBody.Key {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.Post(ctx, "/test", requestBody, &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_PostQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestBody = struct {
		Key string `json:"key"`
	}{
		Key: "value",
	}

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodPost {
				return nil, cliex.ErrBadRequest
			}
			if req.URL.Query().Get("q") != "abc" {
				return nil, cliex.ErrBadRequest
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var reqb struct {
				Key string `json:"key"`
			}
			err = json.Unmarshal(body, &reqb)
			if err != nil {
				return nil, err
			}
			if reqb.Key != requestBody.Key {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.PostQ(ctx, "/test", requestBody, &responseBody, "q", "abc")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_Put(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestBody = struct {
		Key string `json:"key"`
	}{
		Key: "value",
	}

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodPut {
				return nil, cliex.ErrBadRequest
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var reqb struct {
				Key string `json:"key"`
			}
			err = json.Unmarshal(body, &reqb)
			if err != nil {
				return nil, err
			}
			if reqb.Key != requestBody.Key {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.Put(ctx, "/test", requestBody, &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_PutQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestBody = struct {
		Key string `json:"key"`
	}{
		Key: "value",
	}

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodPut {
				return nil, cliex.ErrBadRequest
			}
			if req.URL.Query().Get("q") != "abc" {
				return nil, cliex.ErrBadRequest
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var reqb struct {
				Key string `json:"key"`
			}
			err = json.Unmarshal(body, &reqb)
			if err != nil {
				return nil, err
			}
			if reqb.Key != requestBody.Key {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.PutQ(ctx, "/test", requestBody, &responseBody, "q", "abc")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_Patch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestBody = struct {
		Key string `json:"key"`
	}{
		Key: "value",
	}

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodPatch {
				return nil, cliex.ErrBadRequest
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var reqb struct {
				Key string `json:"key"`
			}
			err = json.Unmarshal(body, &reqb)
			if err != nil {
				return nil, err
			}
			if reqb.Key != requestBody.Key {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.Patch(ctx, "/test", requestBody, &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_PatchQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestBody = struct {
		Key string `json:"key"`
	}{
		Key: "value",
	}

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodPatch {
				return nil, cliex.ErrBadRequest
			}
			if req.URL.Query().Get("q") != "abc" {
				return nil, cliex.ErrBadRequest
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var reqb struct {
				Key string `json:"key"`
			}
			err = json.Unmarshal(body, &reqb)
			if err != nil {
				return nil, err
			}
			if reqb.Key != requestBody.Key {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.PatchQ(ctx, "/test", requestBody, &responseBody, "q", "abc")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_Delete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodDelete {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.Delete(ctx, "/test", &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestHTTP_DeleteQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var requestCounter atomic.Int64
	responseMap := cliex.ResponseMapForTest{
		"/test": func(ctx context.Context, req *http.Request) (interface{}, error) {
			if req.Method != http.MethodDelete {
				return nil, cliex.ErrBadRequest
			}
			if req.URL.Query().Get("q") != "abc" {
				return nil, cliex.ErrBadRequest
			}
			return map[string]string{"key": "value"}, nil
		},
	}
	cfg := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client, err := cliex.NewWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	var responseBody map[string]string
	resp, err := client.DeleteQ(ctx, "/test", &responseBody, "q", "abc")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "value", responseBody["key"])

	assert.Equal(t, int64(1), requestCounter.Load())
}

func TestRetryMechanism(t *testing.T) {
	// Count of how many times the handler has been invoked.
	var requestCount int32

	// Mock server responding with error until the 3rd attempt.
	retryThreshold := int32(30)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentAttempt := atomic.AddInt32(&requestCount, 1)

		if currentAttempt < retryThreshold {
			http.Error(w, "AAA", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Configure the HTTP client to retry 4 times max, with short retry times for testing.
	client, err := cliex.NewWithConfig(cliex.Config{
		BaseURL:        server.URL,
		RequestTimeout: 1 * time.Second,
		CircuitBreaker: false,
	})
	require.NoError(t, err)

	type Response struct {
		Message string `json:"message"`
	}

	var result Response
	ctx := context.Background()
	_, err = client.Request(ctx, "/", cliex.RequestOpts{
		Method:                http.MethodGet,
		Result:                &result,
		ForceContentType:      cliex.MIMETypeJSON,
		RetryOnlyServerErrors: true,
		RetryCount:            40,
		RetryWaitTime:         10 * time.Millisecond,
		RetryMaxWaitTime:      50 * time.Millisecond,
	})
	assert.ErrorContains(t, err, "bad request")
	assert.False(t, strings.Contains(err.Error(), "retr"))

	requestCount = 0

	response, err := client.Request(ctx, "/", cliex.RequestOpts{
		Method:           http.MethodGet,
		Result:           &result,
		ForceContentType: cliex.MIMETypeJSON,
		RetryCount:       40,
		RetryWaitTime:    10 * time.Millisecond,
		RetryMaxWaitTime: 50 * time.Millisecond,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode())
	assert.Equal(t, "success", result.Message)
	assert.Equal(t, retryThreshold, requestCount)

	requestCount = 0

	_, err = client.Request(ctx, "/", cliex.RequestOpts{
		Method:           http.MethodGet,
		Result:           &result,
		ForceContentType: cliex.MIMETypeJSON,
		RetryCount:       20,
		RetryWaitTime:    10 * time.Millisecond,
		RetryMaxWaitTime: 50 * time.Millisecond,
	})
	if assert.ErrorContains(t, err, "bad request") {
		assert.True(t, strings.Contains(err.Error(), "retr"))
	}
}

func TestCircuitBreaker(t *testing.T) {
	var noError atomic.Bool
	// Mock server always returns 500 Internal Server Error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if noError.Load() {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/error" {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Create an HTTP client with circuit breaker enabled
	httpClient, err := cliex.NewWithConfig(cliex.Config{
		BaseURL:                mockServer.URL,
		CircuitBreaker:         true,
		CircuitBreakerTimeout:  500 * time.Millisecond,
		CircuitBreakerFailures: 4,
	})
	assert.NoError(t, err)

	for i := 0; i < 4; i++ {
		_, err = httpClient.Get(context.Background(), "/error")
		assert.ErrorContains(t, err, "internal server error")
	}

	for i := 0; i < 10; i++ {
		_, err = httpClient.Get(context.Background(), "/error")
		assert.ErrorContains(t, err, "circuit breaker is open")
	}

	// another url
	for i := 0; i < 10; i++ {
		_, err = httpClient.Get(context.Background(), "/ok")
		assert.NoError(t, err)
	}

	time.Sleep(time.Second)

	// The server still returns an error, so this should cause the circuit breaker to open again

	_, err = httpClient.Get(context.Background(), "/error")
	assert.ErrorContains(t, err, "internal server error")

	for i := 0; i < 10; i++ {
		_, err = httpClient.Get(context.Background(), "/error")
		assert.ErrorContains(t, err, "circuit breaker is open")
	}

	time.Sleep(time.Second)

	noError.Store(true)

	for i := 0; i < 10; i++ {
		_, err = httpClient.Get(context.Background(), "/error")
		assert.NoError(t, err)
	}
}

func TestIsServerError(t *testing.T) {
	cases := []struct {
		err      error
		expected bool
	}{
		{errors.New("code 500, internal server error"), true},
		{errors.New("code 404, not found"), false},
		{nil, false},
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			isServerError := cliex.IsServerError(c.err)
			assert.Equal(t, c.expected, isServerError)
		})
	}
}

func TestGetCodeFromError(t *testing.T) {
	cases := []struct {
		err      error
		expected int
	}{
		{cliex.ErrBadGateway, 502},
		{cliex.ErrNotFound, 404},
		{fmt.Errorf("some error %w: %s", cliex.ErrNotFound, "bad request"), 404},
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			code := cliex.GetCodeFromError(c.err)
			assert.Equal(t, c.expected, code)
		})
	}
}
