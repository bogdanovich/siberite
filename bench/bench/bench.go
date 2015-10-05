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

var QueueHost = flag.String("host", "localhost", "queue host")
var QueuePort = flag.Int("port", 22133, "queue port")
var QueueName = flag.String("queue", "db_bench", "queue name")
var NumGoroutines = flag.Int("concurrency", 1, "max concurrent lookups")
var NumQueues = flag.Int("queues", 1, "number of simultaneously used queues")
var NumSets = flag.Int("sets", 0, "number of set commands")
var NumGets = flag.Int("gets", 0, "number of get commands")
var ItemSize = flag.Int("item_size", 64, "item size")

type dataSource struct {
	buf []byte
	io.Reader
}

func (self *dataSource) GetData() []byte {
	self.Read(self.buf)
	return self.buf
}

func getQueueName() string {
	if *NumQueues > 1 {
		return fmt.Sprintf("%s%d", *QueueName, rand.Intn(*NumQueues))
	}
	return *QueueName
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
	memc, err := gomemcache.Connect(*QueueHost, *QueuePort)
	defer memc.Close()
	if err != nil {
		log.Println(err)
		return
	}

	setsRemaning := *NumSets
	getsRemaning := *NumGets
	getSetRatio := float32(*NumGets) / float32(*NumSets)
	dataSource := &dataSource{make([]byte, *ItemSize), randbo.New()}

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
						getsRemaning -= 1
					} else if float32(getsRemaning)/float32(setsRemaning) > getSetRatio {
						err = get(memc)
						getsRemaning -= 1
					} else {
						err = set(memc, dataSource)
						setsRemaning -= 1
					}
				} else if setsRemaning > 0 {
					err = set(memc, dataSource)
					setsRemaning -= 1
				} else {
					return
				}
				if err != nil && err.Error() != "memcache: not found" {
					log.Println(err)
					memc, err = gomemcache.Connect(*QueueHost, *QueuePort)
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
	for i := 0; i < *NumGoroutines; i++ {
		wg.Add(1)
		go worker(*QueueName, done, &wg)
	}

	wg.Wait()

	totalGets := *NumGets * *NumGoroutines
	totalSets := *NumSets * *NumGoroutines

	duration := time.Since(startTime)
	fmt.Println("Number of concurrent clients:", *NumGoroutines)
	fmt.Println("Number of queues:", *NumQueues)
	fmt.Println("Total gets:", totalGets)
	fmt.Println("Total sets:", totalSets)
	fmt.Println("Time taken for tests:", duration.Seconds(), "seconds")
	fmt.Println("Bytes read:", *ItemSize*totalGets/1024, "Kbytes")
	fmt.Printf("Read rate: %f Kbytes/sec\r\n", float64(*ItemSize*totalGets)/duration.Seconds()/1024)
	fmt.Println("Bytes written:", *ItemSize*totalSets/1024, "Kbytes")
	fmt.Printf("Write rate: %f Kbytes/sec\r\n", float64(*ItemSize*totalSets)/duration.Seconds()/1024)
	fmt.Printf("Requests per second: %f #/sec\r\n", float64(totalGets+totalSets)/duration.Seconds())
	fmt.Printf("Time per request: %f us (mean)\r\n", float64(duration.Nanoseconds())/float64(totalGets+totalSets))
}
