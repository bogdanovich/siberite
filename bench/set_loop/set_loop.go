package main

import (
	"flag"
	"log"
	"runtime"
	"strconv"
	"sync"

	"github.com/kklis/gomemcache"
)

var QueueHost = flag.String("host", "localhost", "queue host")
var QueuePort = flag.Int("port", 22133, "queue port")
var QueueName = flag.String("queue", "test", "queue name")
var MaxGoroutines = flag.Int("concurrency", 1, "max concurrent lookups")

func loop(queueName string, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	memc, err := gomemcache.Connect(*QueueHost, *QueuePort)
	defer memc.Close()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Set loop started")
	var value string
	for {
		select {
		case <-done:
			log.Println("Done")
			return
		default:
			for i := 0; i < 10000; i++ {
				value = strconv.Itoa(i)
				memc.Set(queueName, []byte(value), 0, 0)
				if err != nil {
					log.Println(err)
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
	done := make(chan struct{})

	for i := 0; i < *MaxGoroutines; i++ {
		wg.Add(1)
		go loop(*QueueName, done, &wg)
	}

	wg.Wait()
}
