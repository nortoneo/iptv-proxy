package proxy

import (
	"errors"
	"sync"
	"time"

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

	return listSema[listName]
}

func lockListConnection(listName string) error {
	sema := getListSema(listName)
	lockTimeout := config.GetConfig().Server.WaitForConnectionSlotTimeout
	for {
		select {
		case sema <- struct{}{}:
			return nil
		case <-time.After(lockTimeout):
			return errors.New("Connection lock timeout")
		}
	}
}

func unlockListConnection(listName string) {
	sema := getListSema(listName)
	<-sema
}
