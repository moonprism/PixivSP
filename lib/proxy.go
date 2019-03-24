package lib

import (
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"time"
)

// setTransport use socket5 proxy
func SetTransport(client *http.Client, addr string) error {

	dialer, err := proxy.SOCKS5("tcp", addr,
		nil,
		&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	)
	if err != nil {
		return err
	}

	client.Transport = &http.Transport{
		Proxy:               nil,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return nil
}