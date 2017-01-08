package filehost

import (
	"sync"
	"time"
)

var (
	mu  sync.Mutex
	ips = map[string]chan struct{}{}
)

func ratelimited(ip string) bool {
	mu.Lock()
	defer mu.Unlock()
	ch, ok := ips[ip]
	if !ok {
		c := make(chan struct{}, 50)
		for i := 0; i < 50; i++ {
			c <- struct{}{}
		}
		ips[ip] = c
		return false
	}
	select {
	case <-ch:
		return false
	default:
		return true
	}
}

func rateAdd(ip string) {
	mu.Lock()
	defer mu.Unlock()
	ch, ok := ips[ip]
	if !ok {
		return
	}
	select {
	case ch <- struct{}{}:
	default:
	}
}

func ratelimit() {
	for _ = range time.Tick(time.Second * 5) {
		mu.Lock()
		for key, ch := range ips {
			select {
			case ch <- struct{}{}:
			default:
				delete(ips, key)
			}
		}
		mu.Unlock()
	}
}
