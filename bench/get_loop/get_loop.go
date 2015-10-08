package main

import (
	"flag"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/kklis/gomemcache"
)

var queueHost = flag.String("host", "localhost", "queue host")
var queuePort = flag.Int("port", 22133, "queue port")
var queueName = flag.String("queue", "test", "queue name")
var maxGoroutines = flag.Int("concurrency", 4, "max concurrent lookups")

func loop(queueName string, done chan struct{}, stop chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	memc, err := gomemcache.Connect(*queueHost, *queuePort)
	defer memc.Close()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Get loop started")
	for {
		select {
		case <-stop:
			log.Println("Stopped")
			return
		default:
			for i := 0; i < 10000; i++ {
				_, _, err := memc.Get(queueName)
				if err != nil {
					log.Println(err)
					done <- struct{}{}
					return
				}
			}
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	var wg sync.WaitGroup
	for {
		done := make(chan struct{}, *maxGoroutines)
		stop := make(chan struct{})
		for i := 0; i < *maxGoroutines; i++ {
			wg.Add(1)
			go loop(*queueName, done, stop, &wg)
		}
		<-done
		log.Println("Stopping other goroutines..")
		close(stop)
		wg.Wait()
		log.Println("Waiting for 60 seconds...")
		time.Sleep(60 * time.Second)
		log.Println("Continue with reads...")
	}

}
