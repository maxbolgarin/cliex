package cliex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/maxbolgarin/lang"
)

const (
	defaultUserAgent      = "Golang HTTP client"
	defaultRequestTimeout = 30 * time.Second

	defaultWaitTime    = time.Second
	defaultMaxWaitTime = 10 * time.Second

	defaultCircuitBreakerTimeout  = 30 * time.Second
	defaultCircuitBreakerFailures = 5
)

// Config is the config for the HTTP client.
type Config struct {
	// BaseURL is the base URL of the server. URL appends to this address.
	// Format "http://localhost:8080/URL" or "https://localhost:8080/URL".
	// Default is empty, means you should provide full URL in Request methods.
	BaseURL string `yaml:"base_url" json:"base_url" env:"CLIEX_BASE_URL"`
	// UserAgent is the User-Agent header that is used for every request.
	// Default is "Golang HTTP client".
	UserAgent string `yaml:"user_agent" json:"user_agent" env:"CLIEX_USER_AGENT"`
	// AuthToken is the Bearer token that is used for every request.
	AuthToken string `yaml:"auth_token" json:"auth_token" env:"CLIEX_AUTH_TOKEN"`
	// ProxyAddress is the address of the proxy server.
	// format "http://localhost:3128".
	// If empty, no proxy will be used.
	ProxyAddress string `yaml:"proxy_address" json:"proxy_address" env:"CLIEX_PROXY_ADDRESS"`
	// RequestTimeout is the timeout for every request in seconds.
	// Default is 30 seconds.
	RequestTimeout time.Duration `yaml:"request_timeout" json:"request_timeout" env:"CLIEX_REQUEST_TIMEOUT"`

	// CAFiles is the list of CA files that are used to verify the server certificate.
	CAFiles []string `yaml:"ca_files" json:"ca_files" env:"CLIEX_CA_FILES"`

	// ClientCertFile and ClientKeyFile are the files that are used to authenticate the client to the server.
	ClientCertFile string `yaml:"client_cert_file" json:"client_cert_file" env:"CLIEX_CLIENT_CERT_FILE"`

	// ClientKeyFile and ClientKeyFile are the files that are used to authenticate the client to the server.
	ClientKeyFile string `yaml:"client_key_file" json:"client_key_file" env:"CLIEX_CLIENT_KEY_FILE"`

	// Insecure is the flag that allows to make requests to the server with invalid SSL certificate.
	// Default is false.
	Insecure bool `yaml:"insecure" json:"insecure" env:"CLIEX_INSECURE"`

	// Debug enables the debug mode.
	Debug bool `yaml:"debug" json:"debug" env:"CLIEX_DEBUG"`

	// CircuitBreaker enables the circuit breaker for url.
	// Default is false.
	CircuitBreaker bool `yaml:"circuit_breaker" json:"circuit_breaker" env:"CLIEX_CIRCUIT_BREAKER"`

	// CircuitBreakerTimeout is the timeout for circuit breaker that turns open state to half-open.
	// Default is 30 seconds.
	CircuitBreakerTimeout time.Duration `yaml:"circuit_breaker_timeout" json:"circuit_breaker_timeout" env:"CLIEX_CIRCUIT_BREAKER_TIMEOUT"`

	// CircuitBreakerFailures is the number of consecutive failures before circuit breaker turns to open state.
	// Default is 5.
	CircuitBreakerFailures uint32 `yaml:"circuit_breaker_failures" json:"circuit_breaker_failures" env:"CLIEX_CIRCUIT_BREAKER_FAILURES"`

	// Logger is the logger that is used in cliex.
	// Default is noop logger, if Debug == true default is JSON debug slog in stderr.
	Logger Logger `yaml:"-" json:"-"`

	// RestyLogger is the logger that is used in resty.
	// Default is Logger.
	RestyLogger resty.Logger `yaml:"-" json:"-"`
}

// WithBaseURL sets the BaseURL field of the Config.
func WithBaseURL(baseURL string) func(*Config) {
	return func(cfg *Config) {
		cfg.BaseURL = baseURL
	}
}

// WithUserAgent sets the UserAgent field of the Config.
func WithUserAgent(userAgent string) func(*Config) {
	return func(cfg *Config) {
		cfg.UserAgent = userAgent
	}
}

// WithAuthToken sets the AuthToken field of the Config.
func WithAuthToken(authToken string) func(*Config) {
	return func(cfg *Config) {
		cfg.AuthToken = authToken
	}
}

// WithProxyAddress sets the ProxyAddress field of the Config.
func WithProxyAddress(proxyAddress string) func(*Config) {
	return func(cfg *Config) {
		cfg.ProxyAddress = proxyAddress
	}
}

// WithRequestTimeout sets the RequestTimeout field of the Config.
func WithRequestTimeout(requestTimeout time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.RequestTimeout = requestTimeout
	}
}

// WithInsecure sets the Insecure field of the Config.
func WithInsecure(insecure bool) func(*Config) {
	return func(cfg *Config) {
		cfg.Insecure = insecure
	}
}

// WithLogger sets the Logger field of the Config.
func WithLogger(logger Logger) func(*Config) {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}

// WithRestyLogger sets the RestyLogger field of the Config.
func WithRestyLogger(logger resty.Logger) func(*Config) {
	return func(cfg *Config) {
		cfg.RestyLogger = logger
	}
}

// WithDebug sets the Debug field of the Config.
func WithDebug(debug bool) func(*Config) {
	return func(cfg *Config) {
		cfg.Debug = debug
	}
}

// WithCAFiles sets the CAFiles field of the Config.
func WithCAFiles(caFiles ...string) func(*Config) {
	return func(cfg *Config) {
		cfg.CAFiles = caFiles
	}
}

// WithClientCertFile sets the ClientCertFile field of the Config.
func WithClientCertFile(clientCertFile string) func(*Config) {
	return func(cfg *Config) {
		cfg.ClientCertFile = clientCertFile
	}
}

// WithClientKeyFile sets the ClientKeyFile field of the Config.
func WithClientKeyFile(clientKeyFile string) func(*Config) {
	return func(cfg *Config) {
		cfg.ClientKeyFile = clientKeyFile
	}
}

// HTTPAddressRegexp is used to match URLs starting with "http://" or "https://", with an optional "www." prefix.
var HTTPAddressRegexp = regexp.MustCompile(`^https?:\/\/(www\.)?([-a-zA-Z0-9@:%._\+~#=]{1,256}(\.|:)[a-zA-Z0-9()]{1,5}|:[0-9]{2,5})(/[-a-zA-Z0-9()@:%_\+.~#?&//=]*)*$`)

func (cfg *Config) prepareAndValidate() error {
	cfg.UserAgent = lang.Check(cfg.UserAgent, defaultUserAgent)
	cfg.RequestTimeout = lang.Check(cfg.RequestTimeout, defaultRequestTimeout)

	if cfg.BaseURL != "" && !HTTPAddressRegexp.MatchString(cfg.BaseURL) {
		return fmt.Errorf("invalid base url address=%s", cfg.BaseURL)
	}
	if cfg.ProxyAddress != "" && !HTTPAddressRegexp.MatchString(cfg.ProxyAddress) {
		return fmt.Errorf("invalid proxy address=%s", cfg.ProxyAddress)
	}
	if cfg.ClientCertFile != "" && cfg.ClientKeyFile == "" {
		return errors.New("client key file is empty")
	}
	if cfg.ClientKeyFile != "" && cfg.ClientCertFile == "" {
		return errors.New("client cert file is empty")
	}
	if cfg.Logger == nil {
		if cfg.Debug {
			cfg.Logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		} else {
			cfg.Logger = noopLogger{}
		}
	}
	if cfg.RestyLogger == nil {
		cfg.RestyLogger = newRestyLogger(cfg.Logger)
	}
	cfg.CircuitBreakerTimeout = lang.Check(cfg.CircuitBreakerTimeout, defaultCircuitBreakerTimeout)
	cfg.CircuitBreakerFailures = lang.Check(cfg.CircuitBreakerFailures, defaultCircuitBreakerFailures)

	return nil
}

// ResponseMapForTest is a map that contains functions that will be used to generate responses for tests.
type ResponseMapForTest map[string]func(context.Context, *http.Request) (any, error)

// GetConfigForTest returns a new Config with the server address set to a test server that will be closed when the given context is closed.
// The requests counter will be increased every time a request is made to the test server.
// The responses will be generated by the functions in the response map.
func GetConfigForTest(ctx context.Context, requestCounter *atomic.Int64, responseMap ResponseMapForTest) Config {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestCounter.Add(1)

		rw.Header().Set("Content-Type", "application/json")

		var out any
		if f, ok := responseMap[req.URL.Path]; ok {
			var err error
			if out, err = f(ctx, req); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if out != nil {
			body, err := json.Marshal(out)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err = rw.Write(body); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		}

	}))
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return Config{
		BaseURL:  srv.URL,
		Insecure: true,
	}
}
