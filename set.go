package cliex

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/maxbolgarin/abstract"
	"github.com/maxbolgarin/lang"
)

// HTTPSet is a set of HTTP clients. It is used to send requests to multiple HTTP clients.
// It also handles broken clients - clients that return errors during requests.
type HTTPSet struct {
	clients   []*HTTP
	broken    *abstract.SafeSet[int]
	log       Logger
	useBroken bool
}

// NewSet returns a new HTTPSet with provided clients.
// You can add client using Add method.
func NewSet(clis ...*HTTP) *HTTPSet {
	return &HTTPSet{
		log:     noopLogger{},
		broken:  abstract.NewSafeSet[int](),
		clients: clis,
	}
}

// NewSet returns a new HTTPSet with clients inited with the given configs.
func NewSetFromConfigs(cfgs ...Config) (*HTTPSet, error) {
	out := NewSet()
	if err := out.Add(cfgs...); err != nil {
		return nil, err
	}
	return out, nil
}

// WithLogger sets the Logger field of the HTTPSet.
// Logger is used in panic recovering during requests.
func (c *HTTPSet) WithLogger(logger Logger) *HTTPSet {
	c.log = logger
	return c
}

// Add adds a new HTTP client to the set.
func (c *HTTPSet) Add(cfgs ...Config) error {
	if len(cfgs) == 0 {
		return nil
	}
	if c.clients == nil {
		c.clients = make([]*HTTP, 0, len(cfgs))
	}

	for i, cfg := range cfgs {
		cli, err := NewWithConfig(cfg)
		if err != nil {
			return fmt.Errorf("client %d: %w", i, err)
		}
		c.clients = append(c.clients, cli)
	}

	return nil
}

// UseBroken returns a new HTTPSet with the same clients but with the UseBroken flag set.
// When you call Request on this set, only broken clients will be used.
// Client with successful request will be deleted from broken list.
func (c *HTTPSet) UseBroken() (*HTTPSet, bool) {
	if c.broken.Len() == 0 {
		return nil, false
	}

	out := &HTTPSet{
		clients:   c.clients,
		broken:    c.broken,
		useBroken: true,
	}

	return out, true
}

// GetBroken returns the list of broken clients.
func (c *HTTPSet) GetBroken() []int {
	if c.broken.Len() == 0 {
		return nil
	}
	return c.broken.Values()
}

// DeleteBroken deletes the given client from list of broken clients.
func (c *HTTPSet) DeleteBroken(indxs ...int) {
	for _, i := range indxs {
		c.broken.Delete(i)
	}
}

// Client returns the client at the given index.
func (c *HTTPSet) Client(i int) *HTTP {
	return lang.Index(c.clients, i)
}

// Request makes a request to the given URL with the given options and returns a list of responses.
// If useBroken is false, only working clients will be used.
// If useBroken is true, only broken clients will be used.
func (c *HTTPSet) Request(ctx context.Context, url string, opts RequestOpts) ([]*resty.Response, error) {
	var (
		fs    = make([]*abstract.Future[*resty.Response], len(c.clients))
		resps = make([]*resty.Response, 0, len(c.clients))

		errs []error
	)

	for i, http := range c.clients {
		if c.useBroken && !c.broken.Has(i) {
			continue // useBroken: send only in broken
		}
		if !c.useBroken && c.broken.Has(i) {
			continue // !useBroken: send only in working
		}
		fs[i] = abstract.NewFuture(ctx, c.log, func(ctx context.Context) (*resty.Response, error) {
			return http.Request(ctx, url, opts)
		})
	}

	for i, f := range fs {
		if f == nil {
			continue
		}
		resp, err := f.Get(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("client %d: %w", i, err))
			c.broken.Add(i)
		} else {
			c.broken.Delete(i)
			resps = append(resps, resp)
		}
	}

	return resps, errors.Join(errs...)
}

// Req makes a request to the given URL with the given options and returns a list of responses.
func (c *HTTPSet) Req(ctx context.Context, method string, url string, requestAndResponseBody ...any) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: method,
		Body:   lang.First(requestAndResponseBody),
		Result: lang.Index(requestAndResponseBody, 1)})
}

// Get makes a GET request to the given URL and returns a list of responses.
func (c *HTTPSet) Get(ctx context.Context, url string, responseBody ...any) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Body: lang.First(responseBody)})
}

// GetQ makes a GET request to the given URL with the given query and returns a list of responses.
func (c *HTTPSet) GetQ(ctx context.Context, url string, responseBody any, queryPairs ...string) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Body:  responseBody,
		Query: lang.PairsToMap(queryPairs)})
}

// Post makes a POST request to the given URL with the given request body and returns a list of responses.
func (c *HTTPSet) Post(ctx context.Context, url string, requestBody any, responseBody ...any) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPost,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// PostQ makes a POST request to the given URL with the given request body and query and returns a list of responses.
func (c *HTTPSet) PostQ(ctx context.Context, url string, requestBody any, responseBody any, queryPairs ...string) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPost,
		Body:   requestBody,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Put makes a PUT request to the given URL with the given request body and returns a list of responses.
func (c *HTTPSet) Put(ctx context.Context, url string, requestBody any, responseBody ...any) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPut,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// PutQ makes a PUT request to the given URL with the given request body and query and returns a list of responses.
func (c *HTTPSet) PutQ(ctx context.Context, url string, requestBody any, responseBody any, queryPairs ...string) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPut,
		Body:   requestBody,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Patch makes a PATCH request to the given URL with the given request body and returns a list of responses.
func (c *HTTPSet) Patch(ctx context.Context, url string, requestBody any, responseBody ...any) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPatch,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// PatchQ makes a PATCH request to the given URL with the given request body and query and returns a list of responses.
func (c *HTTPSet) PatchQ(ctx context.Context, url string, requestBody any, responseBody any, queryPairs ...string) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodPatch,
		Body:   requestBody,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}

// Delete makes a DELETE request to the given URL with the given request body and returns a list of responses.
func (c *HTTPSet) Delete(ctx context.Context, url string, requestBody any, responseBody ...any) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodDelete,
		Body:   requestBody,
		Result: lang.First(responseBody)})
}

// DeleteQ makes a DELETE request to the given URL with the given request body and query and returns a list of responses.
func (c *HTTPSet) DeleteQ(ctx context.Context, url string, requestBody any, responseBody any, queryPairs ...string) ([]*resty.Response, error) {
	return c.Request(ctx, url, RequestOpts{
		Method: http.MethodDelete,
		Body:   requestBody,
		Result: responseBody,
		Query:  lang.PairsToMap(queryPairs)})
}
