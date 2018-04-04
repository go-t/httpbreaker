package httpbreaker

import (
	"net/http"

	"github.com/sony/gobreaker"
)

type Client struct {
	*http.Client
	*gobreaker.CircuitBreaker
}

func NewClient(cfg gobreaker.Settings) *Client {
	return Wrap(&http.Client{}, cfg)
}

func Wrap(client *http.Client, cfg gobreaker.Settings) *Client {
	cb := gobreaker.NewCircuitBreaker(cfg)
	tr := client.Transport
	if tr == nil {
		tr = http.DefaultTransport
	}
	client.Transport = &breaker{
		tr: tr,
		cb: cb,
	}
	return &Client{Client: client, CircuitBreaker: cb}
}

type breaker struct {
	tr http.RoundTripper
	cb *gobreaker.CircuitBreaker
}

func (br breaker) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := br.cb.Execute(func() (interface{}, error) {
		return br.tr.RoundTrip(req)
	})
	if err != nil {
		return nil, err
	} else {
		return res.(*http.Response), nil
	}
}
