package config

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

func newHTTPClient(requestTimeout time.Duration, useCookie, disableKeepAlive, disableCompression bool, proxyFunc func(*http.Request) (*url.URL, error)) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 10 ^ 9,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     10 * time.Second,
		DisableCompression:  disableCompression,
		DisableKeepAlives:   disableKeepAlive,
	}
	if proxyFunc == nil {
		transport.Proxy = http.ProxyFromEnvironment
	}
	client := http.Client{
		Timeout:   requestTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if useCookie {
		jar, _ := cookiejar.New(nil)
		client.Jar = jar
	}
	return &client
}
