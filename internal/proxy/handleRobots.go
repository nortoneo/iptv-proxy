package proxy

import (
	"net/http"
)

func handleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	content := "User-agent: *\nDisallow: /"
	w.Write([]byte(content))
}
