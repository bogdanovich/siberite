package repository

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/bogdanovich/siberite/cgroup"
	"github.com/streamrail/concurrent-map"
)

// Version represents siberite version
const Version = "siberite-0.5"

// QueueRepository represents a repository of queues
type QueueRepository struct {
	storage  cmap.ConcurrentMap
	DataPath string
	Stats    *Stats
	sync.Mutex
}

// Stats keeps service stat fields
type Stats struct {
	Version            string
	StartTime          int64
	CurrentConnections uint64
	TotalConnections   uint64
	CmdGet             uint64
	CmdSet             uint64
}

// StatItem - a single stats item
type StatItem struct {
	Key   string
	Value string
}

var err error

// NewRepository and open all queues in the data directory
func NewRepository(dataDir string) (*QueueRepository, error) {
	dataPath, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	stats := &Stats{Version, time.Now().Unix(), 0, 0, 0, 0}
	repo := QueueRepository{storage: cmap.New(), DataPath: dataPath, Stats: stats}
	return &repo, repo.initialize()
}

// GetQueue returns existing queue from repository,
// creates a new one if it doesn't exist
func (repo *QueueRepository) GetQueue(key string) (*cgroup.CGQueue, error) {
	q, ok := repo.get(key)
	if !ok {
		repo.Lock()
		defer repo.Unlock()
		if q, ok = repo.get(key); !ok {
			q, err = cgroup.CGQueueOpen(key, repo.DataPath)
			if err != nil {
				return nil, err
			}
			repo.storage.Set(key, q)
		}
	}
	return q, nil
}

// DeleteQueue deletes a queue from the repository
func (repo *QueueRepository) DeleteQueue(key string) error {
	if q, ok := repo.get(key); ok {
		q.Drop()
		repo.storage.Remove(key)
	}
	return nil
}

// DeleteAllQueues deletes all queues from the repo
func (repo *QueueRepository) DeleteAllQueues() error {
	var err error
	for pair := range repo.storage.IterBuffered() {
		err = repo.DeleteQueue(pair.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

// FlushQueue removes all items from queue
func (repo *QueueRepository) FlushQueue(key string) error {
	err := repo.DeleteQueue(key)
	if err != nil {
		return err
	}
	// initialize new queue
	_, err = repo.GetQueue(key)
	return err
}

// FlushAllQueues removes all items from all the queues
func (repo *QueueRepository) FlushAllQueues() error {
	var err error
	for pair := range repo.storage.IterBuffered() {
		err = repo.FlushQueue(pair.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

// CloseAllQueues closes all queues
func (repo *QueueRepository) CloseAllQueues() error {
	var err error
	var q *cgroup.CGQueue
	for pair := range repo.storage.IterBuffered() {
		q, err = repo.GetQueue(pair.Key)
		if err != nil {
			return err
		}
		q.Close()
	}
	return nil
}

// FullStats gets repository stats
func (repo *QueueRepository) FullStats() []StatItem {
	stats := []StatItem{}
	currentTime := time.Now().Unix()
	stats = append(stats, StatItem{"uptime", fmt.Sprintf("%d", currentTime-repo.Stats.StartTime)})
	stats = append(stats, StatItem{"time", fmt.Sprintf("%d", currentTime)})
	stats = append(stats, StatItem{"version", fmt.Sprintf("%s", repo.Stats.Version)})
	stats = append(stats, StatItem{"curr_connections", fmt.Sprintf("%d", repo.Stats.CurrentConnections)})
	stats = append(stats, StatItem{"total_connections", fmt.Sprintf("%d", repo.Stats.TotalConnections)})
	stats = append(stats, StatItem{"cmd_get", fmt.Sprintf("%d", repo.Stats.CmdGet)})
	stats = append(stats, StatItem{"cmd_set", fmt.Sprintf("%d", repo.Stats.CmdSet)})
	var q *cgroup.CGQueue
	var cg *cgroup.ConsumerGroup
	for pair := range repo.storage.IterBuffered() {
		q = pair.Val.(*cgroup.CGQueue)
		stats = append(stats, StatItem{"queue_" + q.Name + "_items", fmt.Sprintf("%d", q.Length())})
		stats = append(stats, StatItem{"queue_" + q.Name + "_open_transactions", fmt.Sprintf("%d", q.Stats().OpenReads)})
		for pair := range q.ConsumerGroupIterator() {
			cg = pair.Val.(*cgroup.ConsumerGroup)
			stats = append(stats, StatItem{"queue_" + q.Name + ":" + cg.Name + "_items", fmt.Sprintf("%d", cg.Length())})
			stats = append(stats, StatItem{"queue_" + q.Name + ":" + cg.Name + "_open_transactions", fmt.Sprintf("%d", cg.Stats().OpenReads)})
		}
	}
	return stats
}

// Count returns a total number of queues
func (repo *QueueRepository) Count() int {
	return repo.storage.Count()
}

func (repo *QueueRepository) initialize() error {
	dirs, err := ioutil.ReadDir(repo.DataPath)
	if err != nil {
		return fmt.Errorf("error opening data directory (%s): %s", repo.DataPath, err.Error())
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			// queue initization
			q, err := repo.GetQueue(dir.Name())
			if err != nil {
				log.Fatalf("queue %s...%s", dir.Name(), err.Error())
			}
			log.Printf("queue \"%s\": size %d, head %d, tail %d", dir.Name(), q.Length(), q.Head(), q.Tail())
		}
	}
	return nil
}

func (repo *QueueRepository) get(key string) (*cgroup.CGQueue, bool) {
	val, ok := repo.storage.Get(key)
	if ok {
		return val.(*cgroup.CGQueue), ok
	}
	return nil, ok
}
