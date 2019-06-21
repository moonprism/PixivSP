package lib

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	log "github.com/sirupsen/logrus"
)

type HttpClient struct {
	*http.Client
}

func NewHttpClient() *HttpClient {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	return &HttpClient{
		client,
	}
}

func (c *HttpClient) SetProxy(host string, port string) (err error) {
	dialer, err := proxy.SOCKS5("tcp", host+":"+port,
		nil,
		&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	)
	if err != nil {
		return
	}

	c.Client.Transport = &http.Transport{
		Proxy:               nil,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return
}

func (c *HttpClient) GetResponseDoc(link string) (doc *goquery.Document, err error) {
	resp, err := c.Client.Get(link)
	if err != nil {
		log.Errorf("request url %s failed: %v", link, err)
		return
	}

	defer func() {
		err = resp.Body.Close()
	}()

	// parse html doc
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Errorf("parse html %s failed: %v", link, err)
	}
	return
}

func (c *HttpClient) GetResponseJson(link string, data *interface{}) (err error) {
	resp, err := c.Client.Get(link)
	if err != nil {
		log.Errorf("request url %s failed: %v", link, err)
		return
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Errorf("parse json %s failed: %v", link, err)
		}
	}()


	return
}