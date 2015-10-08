package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/dustin/randbo"
	"github.com/kklis/gomemcache"
)

var queueHost = flag.String("host", "localhost", "queue host")
var queuePort = flag.Int("port", 22133, "queue port")
var queueName = flag.String("queue", "db_bench", "queue name")
var numGoroutines = flag.Int("concurrency", 1, "max concurrent lookups")
var numQueues = flag.Int("queues", 1, "number of simultaneously used queues")
var numSets = flag.Int("sets", 0, "number of set commands")
var numGets = flag.Int("gets", 0, "number of get commands")
var itemSize = flag.Int("item_size", 64, "item size")

type dataSource struct {
	buf []byte
	io.Reader
}

func (source *dataSource) GetData() []byte {
	source.Read(source.buf)
	return source.buf
}

func getQueueName() string {
	if *numQueues > 1 {
		return fmt.Sprintf("%s%d", *queueName, rand.Intn(*numQueues))
	}
	return *queueName
}

func set(memc *gomemcache.Memcache, source *dataSource) error {
	return memc.Set(getQueueName(), source.GetData(), 0, 0)
}

func get(memc *gomemcache.Memcache) error {
	_, _, err := memc.Get(getQueueName())
	return err
}

func worker(queueName string, done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	memc, err := gomemcache.Connect(*queueHost, *queuePort)
	defer memc.Close()
	if err != nil {
		log.Println(err)
		return
	}

	setsRemaning := *numSets
	getsRemaning := *numGets
	getSetRatio := float32(*numGets) / float32(*numSets)
	dataSource := &dataSource{make([]byte, *itemSize), randbo.New()}

	for {
		select {
		case <-done:
			log.Println("Done")
			return
		default:
			for i := 0; i < 10000; i++ {
				if getsRemaning > 0 {
					if setsRemaning < 1 {
						err = get(memc)
						getsRemaning--
					} else if float32(getsRemaning)/float32(setsRemaning) > getSetRatio {
						err = get(memc)
						getsRemaning--
					} else {
						err = set(memc, dataSource)
						setsRemaning--
					}
				} else if setsRemaning > 0 {
					err = set(memc, dataSource)
					setsRemaning--
				} else {
					return
				}
				if err != nil && err.Error() != "memcache: not found" {
					log.Println(err)
					memc, err = gomemcache.Connect(*queueHost, *queuePort)
					if err != nil {
						return
					}
				}
			}
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	done := make(chan struct{})

	startTime := time.Now()
	for i := 0; i < *numGoroutines; i++ {
		wg.Add(1)
		go worker(*queueName, done, &wg)
	}

	wg.Wait()

	totalGets := *numGets * *numGoroutines
	totalSets := *numSets * *numGoroutines

	duration := time.Since(startTime)
	fmt.Println("Number of concurrent clients:", *numGoroutines)
	fmt.Println("Number of queues:", *numQueues)
	fmt.Println("Total gets:", totalGets)
	fmt.Println("Total sets:", totalSets)
	fmt.Println("Time taken for tests:", duration.Seconds(), "seconds")
	fmt.Println("Bytes read:", *itemSize*totalGets/1024, "Kbytes")
	fmt.Printf("Read rate: %f Kbytes/sec\r\n", float64(*itemSize*totalGets)/duration.Seconds()/1024)
	fmt.Println("Bytes written:", *itemSize*totalSets/1024, "Kbytes")
	fmt.Printf("Write rate: %f Kbytes/sec\r\n", float64(*itemSize*totalSets)/duration.Seconds()/1024)
	fmt.Printf("Requests per second: %f #/sec\r\n", float64(totalGets+totalSets)/duration.Seconds())
	fmt.Printf("Time per request: %f us (mean)\r\n", float64(duration.Nanoseconds())/float64(totalGets+totalSets))
}
