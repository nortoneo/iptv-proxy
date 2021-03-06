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
	reqListName := vars["name"]
	reqToken := vars["token"]
	listURLString, err := config.GetListURL(reqListName)
	log.Println("Using list: " + listURLString)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	token, err := config.GetListToken(reqListName)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if token != reqToken {
		log.Println("Wrong token for list " + reqListName)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = lockListConnection(reqListName)
	if err != nil {
		log.Println("Too many connections for list " + reqListName)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	defer unlockListConnection(reqListName)

	proxiedURLString, err := urlconvert.ConvertURLtoProxyURL(listURLString, config.GetConfig().App.URL, reqListName)
	log.Println("Proxy list: " + proxiedURLString)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, nosnippet")
	w.Header().Set("location", proxiedURLString)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
