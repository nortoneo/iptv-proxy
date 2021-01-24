package proxy

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/nortoneo/iptv-proxy/internal/config"
)

var onceClient sync.Once
var httpClient *http.Client

// GetClient returns initialized http client
func GetClient() *http.Client {
	onceClient.Do(func() {
		initClient()
		log.Println("Client initialized")
	})

	return httpClient
}

func initClient() {
	c := config.GetConfig()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: (&net.Dialer{
			Timeout:   c.ClientDialTimeout,
			KeepAlive: c.ClientDialKeepAlive,
		}).Dial,
		TLSHandshakeTimeout:   c.ClientTLSHandshakeTimeout,
		ResponseHeaderTimeout: c.ClientResponseHeaderTimeout,
		ExpectContinueTimeout: c.ClientExpectContinueTimeout,
	}
	httpClient = &http.Client{
		Timeout:   c.ClientTimeout,
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
