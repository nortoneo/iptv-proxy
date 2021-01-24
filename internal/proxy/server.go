package proxy

import (
	"net/http"

	"github.com/nortoneo/iptv-proxy/internal/config"
	"github.com/nortoneo/iptv-proxy/internal/urlconvert"

	"github.com/gorilla/mux"
)

// InitServer starting http server
func InitServer() {
	r := mux.NewRouter()
	r.HandleFunc("/list/{key}", handleListRequest).Name("list")
	r.HandleFunc("/"+urlconvert.GetProxyRoutePrefix()+"{encUrl}"+urlconvert.GetProxyRoutePathSeparator()+"{additionalPath:.*}", handleProxyRequest).Name("proxy")
	r.HandleFunc("/robots.txt", handleRobots).Name("robots")

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
