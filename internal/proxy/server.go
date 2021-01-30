package proxy

import (
	"net/http"

	"github.com/nortoneo/iptv-proxy/internal/config"

	"github.com/gorilla/mux"
)

// InitServer starting http server
func InitServer() {
	r := mux.NewRouter()
	r.HandleFunc("/list/{name}", handleListRequest).Queries("token", "{token}").Name("list")
	r.HandleFunc("/robots.txt", handleRobots).Name("robots")
	r.NotFoundHandler = http.HandlerFunc(handleProxyRequest)

	c := config.GetConfig()
	srv := &http.Server{
		Handler:      r,
		Addr:         c.ListenAddress,
		WriteTimeout: c.ServWriteTimeout,
		ReadTimeout:  c.ServReadTimeout,
		IdleTimeout:  c.ServIdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
