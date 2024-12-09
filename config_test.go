package cliex_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/maxbolgarin/cliex"
	"github.com/stretchr/testify/assert"
)

func TestConfig_WithBaseURL(t *testing.T) {
	config := cliex.Config{}
	assert.Empty(t, config.BaseURL)

	cliex.WithBaseURL("http://example.com")(&config)
	assert.Equal(t, "http://example.com", config.BaseURL)
}

func TestConfig_WithRequestTimeout(t *testing.T) {
	config := cliex.Config{}
	assert.Zero(t, config.RequestTimeout)

	cliex.WithRequestTimeout(10 * time.Second)(&config)
	assert.Equal(t, 10*time.Second, config.RequestTimeout)
}

func TestConfig_WithUserAgent(t *testing.T) {
	config := cliex.Config{}
	assert.Empty(t, config.UserAgent)

	cliex.WithUserAgent("MyUserAgent")(&config)
	assert.Equal(t, "MyUserAgent", config.UserAgent)
}

func TestConfig_WithAuthToken(t *testing.T) {
	config := cliex.Config{}
	assert.Empty(t, config.AuthToken)

	cliex.WithAuthToken("my-token")(&config)
	assert.Equal(t, "my-token", config.AuthToken)
}

func TestConfig_WithProxyAddress(t *testing.T) {
	config := cliex.Config{}
	assert.Empty(t, config.ProxyAddress)

	cliex.WithProxyAddress("http://proxy.example.com")(&config)
	assert.Equal(t, "http://proxy.example.com", config.ProxyAddress)
}

func TestConfig_WithLogger(t *testing.T) {
	config := cliex.Config{}
	assert.Nil(t, config.Logger)

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cliex.WithLogger(logger)(&config)
	assert.NotNil(t, config.Logger)
}

func TestConfig_WithRestyLogger(t *testing.T) {
	config := cliex.Config{}
	assert.Nil(t, config.RestyLogger)

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cliex.WithRestyLogger(newRestyLogger(logger))(&config)
	assert.NotNil(t, config.RestyLogger)
}

func TestConfig_WithInsecure(t *testing.T) {
	config := cliex.Config{}
	assert.False(t, config.Insecure)

	cliex.WithInsecure(true)(&config)
	assert.True(t, config.Insecure)
}

func TestConfig_WithDebug(t *testing.T) {
	config := cliex.Config{}
	assert.False(t, config.Debug)

	cliex.WithDebug(true)(&config)
	assert.True(t, config.Debug)
}

func TestConfig_WithCAFiles(t *testing.T) {
	config := cliex.Config{}
	assert.Nil(t, config.CAFiles)

	cliex.WithCAFiles("ca1.pem", "ca2.pem")(&config)
	assert.Equal(t, []string{"ca1.pem", "ca2.pem"}, config.CAFiles)
}

func TestConfig_WithClientCertFile(t *testing.T) {
	config := cliex.Config{}
	assert.Empty(t, config.ClientCertFile)

	cliex.WithClientCertFile("cert.pem")(&config)
	assert.Equal(t, "cert.pem", config.ClientCertFile)
}

func TestConfig_WithClientKeyFile(t *testing.T) {
	config := cliex.Config{}
	assert.Empty(t, config.ClientKeyFile)

	cliex.WithClientKeyFile("key.pem")(&config)
	assert.Equal(t, "key.pem", config.ClientKeyFile)
}

func TestGetConfigForTest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	responseMap := cliex.ResponseMapForTest{
		"/success": func(ctx context.Context, req *http.Request) (interface{}, error) {
			return map[string]string{"message": "success"}, nil
		},
		"/error": func(ctx context.Context, req *http.Request) (interface{}, error) {
			return nil, cliex.ErrInternalServerError
		},
	}

	var requestCounter atomic.Int64
	config := cliex.GetConfigForTest(ctx, &requestCounter, responseMap)

	client := &http.Client{}

	// Test success response
	req, _ := http.NewRequest("GET", config.BaseURL+"/success", nil)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test error response
	req, _ = http.NewRequest("GET", config.BaseURL+"/error", nil)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Ensure requestCounter increment
	assert.Equal(t, int64(2), requestCounter.Load())
}

type restyLogger struct {
	l cliex.Logger
}

func newRestyLogger(l cliex.Logger) restyLogger {
	return restyLogger{l: l}
}

func (l restyLogger) Debugf(format string, v ...any) {
	l.l.Debug(fmt.Sprintf(format, v...))
}

func (l restyLogger) Warnf(format string, v ...any) {
	l.l.Warn(fmt.Sprintf(format, v...))
}

func (l restyLogger) Errorf(format string, v ...any) {
	l.l.Error(fmt.Sprintf(format, v...))
}
