package yammer

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

type RateLimitedClient struct {
	http.Client
	ticker *time.Ticker
	nreqs int
	period time.Duration
	throttle chan time.Time
}

func (r *RateLimitedClient) Do(req *http.Request) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Do(req)
}

func (r *RateLimitedClient) Get(url string) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Get(url)
}

func (r *RateLimitedClient) Head(url string) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Head(url)
}

func (r *RateLimitedClient) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Post(url, contentType, body)
}

func (r *RateLimitedClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.PostForm(url, data)
}

func NewRateLimitedClient(nreqs int, period time.Duration) *RateLimitedClient {
	r := RateLimitedClient{
		Client: http.Client{},
		nreqs: nreqs,
		period: period,
		throttle: make(chan time.Time, nreqs),
	}

	r.ticker = time.NewTicker(r.period)
	n := r.nreqs
	for n > 0 {
		r.throttle <- time.Now()
		n--
	}

	go func() {
		for {
			select {
			case <-r.ticker.C:
				n := r.nreqs
				for n > 0 {
					select {
					case r.throttle <- time.Now():
					default:
					n--
					}
				}
			}
		}
	}()

	return &r
}

type Client struct {
	baseURL     string
	bearerToken string
	connection  *http.Client
}

func New(bearerToken string) *Client {
	return &Client{
		baseURL:     "https://www.yammer.com",
		bearerToken: bearerToken,
		connection:  &http.Client{},
	}
}
