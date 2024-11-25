package cliex

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/maxbolgarin/abstract"
	"github.com/maxbolgarin/errm"
	"github.com/maxbolgarin/lang"
	"github.com/maxbolgarin/logze"
)

// HTTPSet is a set of HTTP clients.
type HTTPSet struct {
	clients   []*HTTP
	broken    *abstract.SafeSet[int]
	log       logze.Logger
	useBroken bool
}

// NewEmptySet returns a new HTTPSet with no clients.
func NewEmptySet() *HTTPSet {
	return &HTTPSet{
		broken: abstract.NewSafeSet[int](),
	}
}

// NewSet returns a new HTTPSet with the given configurations.
func NewSet(cfgs ...Config) (*HTTPSet, error) {
	out := NewEmptySet()
	if err := out.Add(cfgs...); err != nil {
		return nil, err
	}
	return out, nil
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
			return errm.Wrapf(err, "create %d HTTP client", i+1)
		}
		c.clients = append(c.clients, cli)
	}

	return nil
}

// WithLogger sets the logger to the set.
func (c *HTTPSet) WithLogger(log logze.Logger) {
	c.log = log
	for _, http := range c.clients {
		http.WithLogger(log)
	}
}

// R returns a new resty request with the first client.
func (c *HTTPSet) R(ctx context.Context) *resty.Request {
	return c.clients[0].R(ctx)
}

// Request makes a request to the given URL with the given options and returns a list of responses.
func (c *HTTPSet) Request(ctx context.Context, url string, opts RequestOpts) ([]*resty.Response, error) {
	var (
		errs  = errm.NewSafeList()
		fs    = make([]*abstract.Future[*resty.Response], len(c.clients))
		resps = make([]*resty.Response, 0, len(c.clients))
	)

	for i, http := range c.clients {
		if c.useBroken && !c.broken.Has(i) {
			continue // useBroken: send only in broken
		}
		if !c.useBroken && c.broken.Has(i) {
			continue // !useBroken: send only in working
		}
		fs[i] = abstract.NewFuture(ctx, logze.ConvertToS(c.log), func(ctx context.Context) (*resty.Response, error) {
			return http.Request(ctx, url, opts)
		})
	}

	for i, f := range fs {
		if f == nil {
			continue
		}
		resp, err := f.Get(ctx)
		if err != nil {
			errs.Wrap(err, "request", "client", i)
			c.broken.Add(i)
		} else {
			c.broken.Delete(i)
			resps = append(resps, resp)
		}
	}

	return resps, errs.Err()
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

// Send sends the given request body to the given URL and adds the given query pairs.
func (c *HTTPSet) Send(ctx context.Context, requestBody []byte, queryPairs ...string) error {
	var (
		errs = errm.NewSafeList()
		ws   = make([]*abstract.Waiter, len(c.clients))
	)

	for i, http := range c.clients {
		if c.useBroken && !c.broken.Has(i) {
			continue // useBroken: send only in broken
		}
		if !c.useBroken && c.broken.Has(i) {
			continue // !useBroken: send only in working
		}
		ws[i] = abstract.NewWaiter(ctx, logze.ConvertToS(c.log), func(ctx context.Context) error {
			return http.Send(ctx, requestBody, queryPairs...)
		})
	}

	for i, f := range ws {
		if f == nil {
			continue
		}
		if err := f.Await(ctx); err != nil {
			errs.Wrap(err, "send", "client", i)
			c.broken.Add(i)
		} else {
			c.broken.Delete(i)
		}
	}

	return errs.Err()
}

// UseBroken returns a new HTTPSet with the same clients but with the UseBroken flag set.
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

// DeleteBroken deletes the given clients from the set.
func (c *HTTPSet) DeleteBroken(indxs ...int) {
	for _, i := range indxs {
		c.broken.Delete(i)
	}
}

// Client returns the client at the given index.
func (c *HTTPSet) Client(i int) *HTTP {
	return lang.Index(c.clients, i)
}
