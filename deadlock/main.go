package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	var a, b int
	aMux, bMux := &sync.Mutex{}, &sync.Mutex{}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		aMux.Lock()
		time.Sleep(1 * time.Second)
		a = 2
		bMux.Lock()
		b = 3
		bMux.Unlock()
		aMux.Unlock()
	}()

	go func() {
		defer wg.Done()

		bMux.Lock()
		time.Sleep(1 * time.Second)
		a = 2
		aMux.Lock()
		b = 3
		aMux.Unlock()
		bMux.Unlock()
	}()

	wg.Wait()
	log.Println(a, b)
}
