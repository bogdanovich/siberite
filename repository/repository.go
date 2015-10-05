package repository

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/bogdanovich/siberite/queue"
	"github.com/streamrail/concurrent-map"
)

const Version = "siberite-0.3.1"

type QueueRepository struct {
	storage  cmap.ConcurrentMap
	DataPath string
	Stats    *Stats
	sync.Mutex
}

// Service stats
type Stats struct {
	Version            string
	StartTime          int64
	CurrentConnections uint64
	TotalConnections   uint64
	CmdGet             uint64
	CmdSet             uint64
}

type StatItem struct {
	Key   string
	Value string
}

var err error

// Open and initialize all queues in data directory
func Initialize(dataDir string) (*QueueRepository, error) {
	dataPath, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	stats := &Stats{Version, time.Now().Unix(), 0, 0, 0, 0}
	repo := QueueRepository{storage: cmap.New(), DataPath: dataPath, Stats: stats}
	return &repo, repo.initialize()
}

// Get existing queue from repository or create a new one
func (self *QueueRepository) GetQueue(key string) (*queue.Queue, error) {
	q, ok := self.get(key)
	if !ok {
		self.Lock()
		if q, ok = self.get(key); !ok {
			q, err = queue.Open(key, self.DataPath)
			if err != nil {
				return nil, err
			}
			self.storage.Set(key, q)
		}
		self.Unlock()
	}
	return q, nil
}

// Delete queue from repository
func (self *QueueRepository) DeleteQueue(key string) error {
	if q, ok := self.get(key); ok {
		q.Drop()
		self.storage.Remove(key)
	}
	return nil
}

// Delete all queues from repository
func (self *QueueRepository) DeleteAllQueues() error {
	var err error
	for pair := range self.storage.IterBuffered() {
		err = self.DeleteQueue(pair.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove all items from queue
func (self *QueueRepository) FlushQueue(key string) error {
	err := self.DeleteQueue(key)
	if err != nil {
		return err
	}
	// initialize new queue
	_, err = self.GetQueue(key)
	return err
}

// Remove all items from all queues
func (self *QueueRepository) FlushAllQueues() error {
	var err error
	for pair := range self.storage.IterBuffered() {
		err = self.FlushQueue(pair.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *QueueRepository) CloseAllQueues() error {
	var err error
	var q *queue.Queue
	for pair := range self.storage.IterBuffered() {
		q, err = self.GetQueue(pair.Key)
		if err != nil {
			return err
		}
		q.Close()
	}
	return nil
}

// Get repository stats
func (self *QueueRepository) FullStats() []StatItem {
	stats := []StatItem{}
	currentTime := time.Now().Unix()
	stats = append(stats, StatItem{"uptime", fmt.Sprintf("%d", currentTime-self.Stats.StartTime)})
	stats = append(stats, StatItem{"time", fmt.Sprintf("%d", currentTime)})
	stats = append(stats, StatItem{"version", fmt.Sprintf("%s", self.Stats.Version)})
	stats = append(stats, StatItem{"curr_connections", fmt.Sprintf("%d", self.Stats.CurrentConnections)})
	stats = append(stats, StatItem{"total_connections", fmt.Sprintf("%d", self.Stats.TotalConnections)})
	stats = append(stats, StatItem{"cmd_get", fmt.Sprintf("%d", self.Stats.CmdGet)})
	stats = append(stats, StatItem{"cmd_set", fmt.Sprintf("%d", self.Stats.CmdSet)})
	var q *queue.Queue
	for pair := range self.storage.IterBuffered() {
		q = pair.Val.(*queue.Queue)
		stats = append(stats, StatItem{"queue_" + q.Name + "_items", fmt.Sprintf("%d", q.Length())})
		stats = append(stats, StatItem{"queue_" + q.Name + "_open_transactions", fmt.Sprintf("%d", q.Stats.OpenTransactions)})
	}
	return stats
}

func (self *QueueRepository) Count() int {
	return self.storage.Count()
}

func (self *QueueRepository) initialize() error {
	dirs, err := ioutil.ReadDir(self.DataPath)
	if err != nil {
		return fmt.Errorf("error opening data directory (%s): %s", self.DataPath, err.Error())
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			// queue initization
			q, err := self.GetQueue(dir.Name())
			if err != nil {
				log.Printf("initializing queue %s...%s", dir.Name(), err.Error())
			}
			log.Printf("queue \"%s\": size %d, head %d, tail %d", dir.Name(), q.Length(), q.Head(), q.Tail())
		}
	}
	return nil
}

func (self *QueueRepository) get(key string) (*queue.Queue, bool) {
	val, ok := self.storage.Get(key)
	if ok {
		return val.(*queue.Queue), ok
	}
	return nil, ok
}
