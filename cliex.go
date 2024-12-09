package cliex

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/maxbolgarin/abstract"
	"github.com/maxbolgarin/lang"
	"github.com/sony/gobreaker/v2"
)

// HTTP is the resty wrapper for easy use.
type HTTP struct {
	cli *resty.Client
	cbs *abstract.SafeMap[string, *gobreaker.CircuitBreaker[*resty.Response]]
	log Logger

	cbCfg    gobreaker.Settings
	enableCB bool
}

// New returns a new HTTP client weith applied With* options to Config.
func New(optsFuncs ...func(*Config)) (*HTTP, error) {
	var cfg Config
	for _, optsFunc := range optsFuncs {
		optsFunc(&cfg)
	}
	return NewWithConfig(cfg)
}

// MustNew returns a new HTTP client inited with provided config.
// Panics if an error occurs.
func MustNew(optsFuncs ...func(*Config)) *HTTP {
	cli, err := New(optsFuncs...)
	if err != nil {
		panic(err)
	}
	return cli
}

// NewWithConfig returns a new HTTP client inited with provided config.
func NewWithConfig(cfg Config) (*HTTP, error) {
	if err := cfg.prepareAndValidate(); err != nil {
		return nil, err
	}

	cli := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetLogger(cfg.RestyLogger).
		SetHeader("User-Agent", cfg.UserAgent).
		SetTimeout(cfg.RequestTimeout).
		SetJSONMarshaler(jsoniter.ConfigCompatibleWithStandardLibrary.Marshal).
		SetJSONUnmarshaler(jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: cfg.Insecure}).
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(20)).
		SetAllowGetMethodPayload(true).
		SetDebug(cfg.Debug).
		OnAfterResponse(errorHandler)

	if cfg.AuthToken != "" {
		cli.SetHeader("Authorization", cfg.AuthToken)
	}

	if cfg.ProxyAddress != "" {
		cli.SetProxy(cfg.ProxyAddress)
	}

	if len(cfg.CAFiles) > 0 {
		for _, caFile := range cfg.CAFiles {
			cli.SetRootCertificate(caFile)
		}
	}

	if cfg.ClientCertFile != "" && cfg.ClientKeyFile != "" {
		cert1, err := tls.LoadX509KeyPair(cfg.ClientCertFile, cfg.ClientKeyFile)
		if err != nil {
			return nil, err
		}
		cli.SetCertificates(cert1)
	}

	out := &HTTP{
		cli: cli,
		cbs: abstract.NewSafeMap[string, *gobreaker.CircuitBreaker[*resty.Response]](),
		log: cfg.Logger,
		cbCfg: gobreaker.Settings{
			Name:    "HTTP Circuit Breaker",
			Timeout: cfg.CircuitBreakerTimeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= cfg.CircuitBreakerFailures
			},
		},
		enableCB: cfg.CircuitBreaker,
	}

	return out, nil
}

// C returns the resty client.
func (c *HTTP) C() *resty.Client {
	return c.cli
}

// R returns the resty request with applied context.
func (c *HTTP) R(ctx context.Context) *resty.Request {
	return c.cli.R().SetContext(ctx)
}

// Request makes HTTP request with the given options to the BaseURL + URL and returns response.
// It also applies circuit breaker if enabled.
func (c *HTTP) Request(ctx context.Context, url string, opts RequestOpts) (*resty.Response, error) {
	if !c.enableCB {
		return c.request(ctx, url, opts)
	}
	cb, ok := c.cbs.Lookup(url)
	if !ok {
		cb = gobreaker.NewCircuitBreaker[*resty.Response](c.cbCfg)
		c.cbs.Set(url, cb)
	}
	return cb.Execute(func() (*resty.Response, error) {
		return c.request(ctx, url, opts)
	})
}

func (c *HTTP) request(ctx context.Context, url string, opts RequestOpts) (*resty.Response, error) {
	req := c.R(ctx).SetBody(opts.Body).SetResult(opts.Result).SetAuthToken(opts.AuthToken).
		SetHeaders(opts.Headers).SetQueryParams(opts.Query).SetCookies(opts.Cookies).
		ForceContentType(opts.ForceContentType).SetFormData(opts.FormData)
	if opts.BasicAuthUser != "" && opts.BasicAuthPass != "" {
		req.SetBasicAuth(opts.BasicAuthUser, opts.BasicAuthPass)
	}
	if opts.EnableTrace {
		req.EnableTrace()
	}
	if opts.Files != nil {
		req.SetFiles(opts.Files)
	}
	if opts.OutputPath != "" {
		req.SetOutput(opts.OutputPath)
	}
	opts.RequestName = lang.If(opts.RequestName != "", opts.RequestName+" ", "")

	sender := getSender(req, opts.Method)
	url = c.prepareURL(url)

	resp, err := sender(url)
	switch {
	case err == nil:
		return resp, nil
	case (opts.RetryCount == 0 && !opts.InfiniteRetry) || (opts.RetryOnlyServerErrors && !IsServerError(err)):
		return nil, fmt.Errorf("failed %srequest: %w", opts.RequestName, err)
	}

	// Start retry

	opts.RetryCount = lang.If(opts.InfiniteRetry, math.MaxInt, opts.RetryCount)
	opts.RetryWaitTime = lang.Check(opts.RetryWaitTime, defaultWaitTime)
	opts.RetryMaxWaitTime = lang.Check(opts.RetryMaxWaitTime, defaultMaxWaitTime)

	if !opts.NoLogRetryError {
		msg := "failed " + opts.RequestName + "request, "
		if opts.InfiniteRetry {
			msg += "infinite retry"
		} else {
			msg += strconv.Itoa(opts.RetryCount) + " retries"
		}
		c.log.Error(msg, "error", err, "address", c.cli.BaseURL+url)
	}

	var errs []error
	for retry := 1; retry < opts.RetryCount; retry++ {
		sleepTime := getSleepTime(retry, opts.RetryWaitTime, opts.RetryMaxWaitTime)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("request canceled, got errors: %w", errors.Join(errs...))

		case <-time.After(sleepTime):
		}

		resp, err = sender(url)
		if err != nil {
			if !opts.NoLogRetryError {
				c.log.Warn("failed "+opts.RequestName+"request after retry", "error", err, "n", retry, "address", c.cli.BaseURL+url)
			}
			errs = append(errs, err)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("failed %srequest after retries, got errors: %w", opts.RequestName, errors.Join(errs...))
}

// Req performs request with method to the BaseURL +  URL and returns response
func (c *HTTP) Req(ctx context.Context, method string, url string, requestAndResponseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: method,
		Body:   lang.First(requestAndResponseBody),
		Result: lang.Index(requestAndResponseBody, 1)})
}

// Get performs GET request to the BaseURL +  URL and returns response
func (c *HTTP) Get(ctx context.Context, url string, responseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Result: lang.First(responseBody)})
}

// GetQ performs GET request to the BaseURL +  URL with query and returns response
func (c *HTTP) GetQ(ctx context.Context, url string, responseBody any, queryPairs ...string) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Post performs POST request to the BaseURL +  URL and returns response
func (c *HTTP) Post(ctx context.Context, url string, requestBody any, responseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPost,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// PostQ performs POST request to the BaseURL +  URL with query and returns response
func (c *HTTP) PostQ(ctx context.Context, url string, requestBody any, responseBody any, queryPairs ...string) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPost,
		Body:   requestBody,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Put performs PUT request to the BaseURL +  URL and returns response
func (c *HTTP) Put(ctx context.Context, url string, requestBody any, responseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPut,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// PutQ performs PUT request to the BaseURL +  URL with query and returns response
func (c *HTTP) PutQ(ctx context.Context, url string, requestBody any, responseBody any, queryPairs ...string) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPut,
		Body:   requestBody,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Patch performs PATCH request to the BaseURL +  URL and returns response
func (c *HTTP) Patch(ctx context.Context, url string, requestBody any, responseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPatch,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// PatchQ performs PATCH request to the BaseURL +  URL with query and returns response
func (c *HTTP) PatchQ(ctx context.Context, url string, requestBody any, responseBody any, queryPairs ...string) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPatch,
		Body:   requestBody,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Delete performs DELETE request to the BaseURL +  URL and returns response
func (c *HTTP) Delete(ctx context.Context, url string, responseBody ...any) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodDelete,
		Result: lang.First(responseBody)})
}

// DeleteQ performs DELETE request to the BaseURL +  URL with query and returns response
func (c *HTTP) DeleteQ(ctx context.Context, url string, responseBody any, queryPairs ...string) (*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodDelete,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
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

	apiErr, ok := ErrorMapping[r.StatusCode()]
	if !ok {
		apiErr = fmt.Errorf("code %d", r.StatusCode())
	}

	var errBody ServerErrorResponse
	if err := json.Unmarshal(r.Body(), &errBody); err == nil {
		errMsg := getErrorMessage(errBody)
		if errBody.Code != 0 {
			apiErr = lang.Check(ErrorMapping[errBody.Code], apiErr)
		}
		if errMsg != "" {
			return fmt.Errorf("%w: %s", apiErr, errMsg)
		}
	}

	if body := string(r.Body()); body != "" {
		return fmt.Errorf("%w: %s", apiErr, maxLen(body, 100))
	}

	return apiErr
}

func maxLen(a string, b int) string {
	if len(a) > b {
		return a[:b]
	}
	return a
}

func getSleepTime(retry int, min, max time.Duration) time.Duration {
	sleepTime := float64(min) * math.Pow(2, float64(retry))
	sleepTime = rand.Float64()*(sleepTime-float64(min)) + float64(min)
	if sleepTime > float64(max) {
		sleepTime = float64(max)
	}
	return time.Duration(sleepTime)
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

func getErrorMessage(r ServerErrorResponse) string {
	if r.Message != "" {
		return r.Message
	}
	if r.Error != "" {
		return r.Error
	}
	if r.Details != "" {
		return r.Details
	}
	if r.Text != "" {
		return r.Text
	}
	if r.Msg != "" {
		return r.Msg
	}
	if r.Err != "" {
		return r.Err
	}
	return ""
}

func IsServerError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "code 5")
}
