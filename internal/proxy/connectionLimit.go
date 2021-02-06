package proxy

import (
	"sync"

	"github.com/nortoneo/iptv-proxy/internal/config"
)

var listSema = make(map[string]chan struct{})
var initConSemaOnce sync.Once

func getListSema(listName string) chan struct{} {
	// init semaphores once
	initConSemaOnce.Do(func() {
		for k, l := range config.GetConfig().Lists {
			listSema[k] = make(chan struct{}, l.MaxConnections)
		}
	})

	if sema, ok := listSema[listName]; ok {
		return sema
	}

	panic("No sempaphore for list " + listName)
}

func lockListConnection(listName string) {
	sema := getListSema(listName)
	sema <- struct{}{}
}

func unlockListConnection(listName string) {
	sema := getListSema(listName)
	<-sema
}
