package model

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type V2rayClient interface {
	SetOutbound(*Vmess) error
	DeleteOutbound()
	Ping(string) *PingStat
	Close()
}

type PingStat struct {
	V   *Vmess
	Dur time.Duration
	Err error
}

func getHttpClient(host string, port uint32) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse(fmt.Sprintf("%s:%d", host, port))
			},
		},
	}
}

func ping(c *http.Client, host string) *PingStat {
	start := time.Now()
	rsp, err := c.Head(host)
	dur := time.Since(start)

	defer func() {
		if rsp == nil {
			return
		}
		_, _ = io.Copy(io.Discard, rsp.Body)
		_ = rsp.Body.Close()
	}()

	return &PingStat{
		Dur: dur,
		Err: err,
	}
}
