package main

import (
	"log"

	"github.com/nortoneo/iptv-proxy/internal/config"
	"github.com/nortoneo/iptv-proxy/internal/proxy"
)

func main() {
	c := config.GetConfig()
	log.Printf(
		"App started\nApp config: %+v\nServer config: %+v\nClient config: %+v\nPlaylists: %+v",
		c.App, c.Server, c.Client, c.Lists)

	proxy.InitServer()
}
