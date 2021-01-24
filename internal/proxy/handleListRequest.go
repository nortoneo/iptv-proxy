package proxy

import (
	"log"
	"net/http"

	"github.com/nortoneo/iptv-proxy/internal/config"
	"github.com/nortoneo/iptv-proxy/internal/urlconvert"

	"github.com/gorilla/mux"
)

func handleListRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listKey := vars["key"]
	listURLString, err := config.GetListURL(listKey)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	proxiedURLString, err := urlconvert.ConvertURLtoProxyURL(listURLString, config.GetConfig().AppURL)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, nosnippet")
	w.Header().Set("location", proxiedURLString)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
