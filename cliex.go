package cliex

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/maxbolgarin/errm"
	"github.com/maxbolgarin/lang"
	"github.com/maxbolgarin/logze"
)

var (
	// ErrBadRequest means 400 status code from server
	ErrBadRequest = errm.New("bad request")
	// ErrUnauthorized means 401 status code from server
	ErrUnauthorized = errm.New("unauthorized")
	// ErrForbidden means 403 status code from server
	ErrForbidden = errm.New("forbidden")
	// ErrNotFound means 404 status code from server
	ErrNotFound = errm.New("not found")
	// ErrInternalServer means 500 status code from server
	ErrInternalServer = errm.New("server error")
	// ErrBadGateway means 502 status code from server
	ErrBadGateway = errm.New("bad gateway")
)

// ServerErrorResponse is the error response from server (try to guess what it is)
type ServerErrorResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RequestOpts is the options for resty client request
type RequestOpts struct {
	// Method is the HTTP method to use
	Method string
	// Headers is the headers of the request
	Headers map[string]string
	// Query is the query string of the request
	Query map[string]string
	// Cookies is the cookies of the request
	Cookies []*http.Cookie
	// BasicAuthUser is the user for basic authentication
	BasicAuthUser string
	// BasicAuthPass is the password for basic authentication
	BasicAuthPass string
	// RequestName is the name of the request, that used for logging
	RequestName string
	// Body is the body of the request
	Body any
	// Result is the variable where the response body will be stored
	Result any
	// RetryCount is the number of times to retry the request
	RetryCount int
	// RetryWaitTime is the duration to wait before retrying the request
	RetryWaitTime time.Duration
	// Infinite is whether to retry the request infinitely
	Infinite bool
	// NoLogRetryError is whether to log the retry error
	NoLogRetryError bool
}

// HTTP is the resty wrapper for easy use.
type HTTP struct {
	cli *resty.Client
	log logze.Logger

	address string
}

// New returns a new HTTP client with default config
func New() *HTTP {
	cli, _ := NewWithConfig(Config{})
	return cli
}

// NewWithBaseURL returns a new HTTP client with BaseURL
func NewWithBaseURL(baseURL string) *HTTP {
	cli, _ := NewWithConfig(Config{BaseURL: baseURL})
	return cli
}

// NewWithAddress returns a new HTTP client with Address
func NewWithAddress(address string) *HTTP {
	cli, _ := NewWithConfig(Config{Address: address})
	return cli
}

// NewWithConfig returns a new HTTP client inited with provided config
func NewWithConfig(cfg Config) (*HTTP, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errm.Wrap(err, "validate config")
	}

	cli := resty.New().
		SetBaseURL(lang.Check(cfg.BaseURL, cfg.Address)).
		SetLogger(newRestyLogger(lang.If(cfg.Logger.NotInited(), logze.Log, cfg.Logger))).
		SetHeader("User-Agent", lang.Check(cfg.UserAgent, defaultUserAgent)).
		SetTimeout(lang.Check(cfg.RequestTimeout, defaultRequestTimeout)).
		SetJSONMarshaler(jsoniter.ConfigCompatibleWithStandardLibrary.Marshal).
		SetJSONUnmarshaler(jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: cfg.Insecure}).
		OnAfterResponse(errorHandler)

	if cfg.AuthToken != "" {
		cli.SetHeader("Authorization", cfg.AuthToken)
	}

	if cfg.ProxyAddress != "" {
		cli.SetProxy(cfg.ProxyAddress)
	}

	out := &HTTP{
		cli:     cli,
		log:     logze.Log,
		address: cfg.Address,
	}

	return out, nil
}

// WithLogger sets the logger to the client
func (c *HTTP) WithLogger(log logze.Logger) {
	c.cli.SetLogger(newRestyLogger(log))
	c.log = log
}

// C returns the resty client
func (c *HTTP) C() *resty.Client {
	return c.cli
}

// R returns the resty request with applied context
func (c *HTTP) R(ctx context.Context) *resty.Request {
	return c.cli.R().SetContext(ctx)
}

// Request makes HTTP request with the given options to the given URL and returns response.
func (c *HTTP) Request(ctx context.Context, url string, opts RequestOpts) (*resty.Response, error) {
	req := c.R(ctx).SetBody(opts.Body).SetResult(opts.Result).
		SetHeaders(opts.Headers).SetQueryParams(opts.Query).SetCookies(opts.Cookies)
	if opts.BasicAuthUser != "" && opts.BasicAuthPass != "" {
		req.SetBasicAuth(opts.BasicAuthUser, opts.BasicAuthPass)
	}
	sender := getSender(req, opts.Method)
	url = c.prepareURL(url)

	resp, err := sender(url)
	switch {
	case err == nil:
		return resp, nil
	case opts.RetryCount == 0 && !opts.Infinite:
		return nil, err
	}

	// Start retry
	opts.RequestName = lang.If(opts.RequestName != "", opts.RequestName+" ", "")
	opts.RetryCount = lang.If(opts.Infinite, 666666, opts.RetryCount)
	opts.RetryWaitTime = lang.Check(opts.RetryWaitTime, defaultWaitTime)
	if !opts.NoLogRetryError {
		c.log.Errorf(err, "cannot make %srequest, start retrying for %d times with %s interval",
			opts.RequestName, opts.RetryCount, opts.RetryWaitTime, "address", c.cli.BaseURL, "path", url)
	}

	errSet := errm.NewSet()

	for i := 0; i < opts.RetryCount; i++ {
		select {
		case <-ctx.Done():
			return nil, errm.Wrap(errSet.Err(), "requests error")

		case <-time.After(opts.RetryWaitTime):
		}

		if opts.Infinite {
			opts.RetryCount++
		}

		resp, err = sender(url)
		if err != nil {
			if !opts.NoLogRetryError {
				c.log.Warnf("cannot make %srequest", opts.RequestName,
					"error", err, "retry", i+1, "address", c.cli.BaseURL, "path", url)
			}
			errSet.Add(err)
			continue
		}

		return resp, nil
	}

	return nil, errm.Wrap(errSet.Err(), "requests errors")
}

// Req performs request with method to the given URL and returns response
func (c *HTTP) Req(ctx context.Context, method string, url string, requestAndResponseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: method,
		Body:   lang.First(requestAndResponseBody),
		Result: lang.Index(requestAndResponseBody, 1)})
}

// Get performs GET request to the given URL and returns response
func (c *HTTP) Get(ctx context.Context, url string, responseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Result: lang.First(responseBody)})
}

// GetQ performs GET request to the given URL with query and returns response
func (c *HTTP) GetQ(ctx context.Context, url string, responseBody any, queryPairs ...string) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Post performs POST request to the given URL and returns response
func (c *HTTP) Post(ctx context.Context, url string, requestBody any, responseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPost,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// Send sends the Post request to the Address.
func (c *HTTP) Send(ctx context.Context, requestBody []byte, queryPairs ...string) error {
	if c.address == "" {
		return errm.New("address cannot be empty")
	}
	_, err := c.Request(ctx, c.address, RequestOpts{
		Method: http.MethodPost,
		Body:   requestBody,
		Query:  lang.PairsToMap(queryPairs),
	})
	return err
}

func (c *HTTP) prepareURL(url string) string {
	if c.cli.BaseURL == "" && !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

func errorHandler(_ *resty.Client, r *resty.Response) error {
	if r.StatusCode() < 400 {
		return nil
	}
	var apiErr error
	switch r.StatusCode() {
	case http.StatusBadRequest:
		apiErr = ErrBadRequest
	case http.StatusUnauthorized:
		apiErr = ErrUnauthorized
	case http.StatusForbidden:
		apiErr = ErrForbidden
	case http.StatusNotFound:
		apiErr = ErrNotFound
	case http.StatusInternalServerError:
		apiErr = ErrInternalServer
	case http.StatusBadGateway:
		apiErr = ErrBadGateway
	}
	var errBody ServerErrorResponse
	if err := json.Unmarshal(r.Body(), &errBody); err == nil {
		if errBody.Message != "" {
			return errm.Wrap(apiErr, "api error", "code", r.StatusCode(), "message", errBody.Message)
		}
	}
	if body := string(r.Body()); body != "" {
		return errm.Wrap(apiErr, "api error", "code", r.StatusCode(), "body", body)
	}
	return errm.Wrap(apiErr, "api error", "code", r.StatusCode())
}

func getSender(r *resty.Request, method string) func(string) (*resty.Response, error) {
	switch method {
	case http.MethodGet, "":
		return r.Get
	case http.MethodHead:
		return r.Head
	case http.MethodPost:
		return r.Post
	case http.MethodPut:
		return r.Put
	case http.MethodPatch:
		return r.Patch
	case http.MethodDelete:
		return r.Delete
	case http.MethodOptions:
		return r.Options
	}
	return r.Get
}

type restyLogger struct {
	l logze.Logger
}

func newRestyLogger(l logze.Logger) restyLogger {
	return restyLogger{l: l}
}

func (l restyLogger) Debugf(format string, v ...any) {
	l.l.Debugf(format, v...)
}

func (l restyLogger) Warnf(format string, v ...any) {
	l.l.Warnf(format, v...)
}

func (l restyLogger) Errorf(format string, v ...any) {
	l.l.Errorf(nil, format, v...)
}
