package main

import (
	"log"

	"github.com/nortoneo/iptv-proxy/internal/config"
	"github.com/nortoneo/iptv-proxy/internal/proxy"
)

func main() {
	log.Printf("%+v\n", config.GetConfig())
	proxy.InitServer()
}
