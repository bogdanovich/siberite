package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bogdanovich/siberite/cgroup"
)

// Version represents siberite version
const Version = "siberite-0.6.4"

// QueueRepository represents a repository of queues
type QueueRepository struct {
	sync.RWMutex
	storage  map[string]*cgroup.CGQueue
	DataPath string
	Stats    *Stats
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

// NewRepository and open all queues in the data directory
func NewRepository(dataDir string) (*QueueRepository, error) {
	dataPath, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	stats := &Stats{Version, time.Now().Unix(), 0, 0, 0, 0}
	repo := QueueRepository{storage: make(map[string]*cgroup.CGQueue), DataPath: dataPath, Stats: stats}
	return &repo, repo.initialize()
}

// GetQueue returns existing queue from repository,
// creates a new one if it doesn't exist
func (repo *QueueRepository) GetQueue(key string) (*cgroup.CGQueue, error) {
	if q, ok := repo.get(key); ok {
		return q, nil
	}

	repo.Lock()
	defer repo.Unlock()

	// now that we have acquired the lock, recheck to see if someone else
	// already managed to create the queue while we were waiting on the lock
	if q, ok := repo.storage[key]; ok {
		return q, nil
	}

	// ok, we are the first - create the queue
	q, err := cgroup.CGQueueOpen(key, repo.DataPath)
	if err != nil {
		return nil, err
	}
	repo.storage[key] = q
	return q, nil
}

// DeleteQueue deletes a queue from the repository
func (repo *QueueRepository) DeleteQueue(key string) error {
	repo.Lock()
	defer repo.Unlock()
	if q, ok := repo.storage[key]; ok {
		q.Drop()
		delete(repo.storage, key)
	}
	return nil
}

// DeleteAllQueues deletes all queues from the repo
func (repo *QueueRepository) DeleteAllQueues() error {
	for key := range repo.snapshot() {
		if err := repo.DeleteQueue(key); err != nil {
			return err
		}
	}
	return nil
}

// FlushAllQueues removes all items from all the queues
func (repo *QueueRepository) FlushAllQueues() error {
	for _, q := range repo.snapshot() {
		if err := q.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// CloseAllQueues closes all queues
func (repo *QueueRepository) CloseAllQueues() error {
	for _, q := range repo.snapshot() {
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
	stats = append(stats, StatItem{"curr_connections", fmt.Sprintf("%d", atomic.LoadUint64(&repo.Stats.CurrentConnections))})
	stats = append(stats, StatItem{"total_connections", fmt.Sprintf("%d", atomic.LoadUint64(&repo.Stats.TotalConnections))})
	stats = append(stats, StatItem{"cmd_get", fmt.Sprintf("%d", atomic.LoadUint64(&repo.Stats.CmdGet))})
	stats = append(stats, StatItem{"cmd_set", fmt.Sprintf("%d", atomic.LoadUint64(&repo.Stats.CmdSet))})

	for _, q := range repo.snapshot() {
		stats = append(stats, StatItem{"queue_" + q.Name + "_items", fmt.Sprintf("%d", q.Length())})
		stats = append(stats, StatItem{"queue_" + q.Name + "_open_transactions", fmt.Sprintf("%d", q.Stats().OpenReadsValue())})
		for _, cg := range q.ConsumerGroups() {
			stats = append(stats, StatItem{"queue_" + q.Name + "." + cg.Name + "_items", fmt.Sprintf("%d", cg.Length())})
			stats = append(stats, StatItem{"queue_" + q.Name + "." + cg.Name + "_open_transactions", fmt.Sprintf("%d", cg.Stats().OpenReadsValue())})
		}
	}
	return stats
}

// Count returns a total number of queues
func (repo *QueueRepository) Count() int {
	repo.RLock()
	defer repo.RUnlock()
	return len(repo.storage)
}

func (repo *QueueRepository) initialize() error {
	dirs, err := os.ReadDir(repo.DataPath)
	if err != nil {
		return fmt.Errorf("error opening data directory (%s): %s",
			repo.DataPath, err.Error())
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			// queue init
			_, err := repo.GetQueue(dir.Name())
			if err != nil {
				return fmt.Errorf("queue %s: %w", dir.Name(), err)
			}
		}
	}
	return nil
}

func (repo *QueueRepository) get(key string) (*cgroup.CGQueue, bool) {
	repo.RLock()
	defer repo.RUnlock()
	val, ok := repo.storage[key]
	return val, ok
}

func (repo *QueueRepository) snapshot() map[string]*cgroup.CGQueue {
	repo.RLock()
	defer repo.RUnlock()
	queues := make(map[string]*cgroup.CGQueue, len(repo.storage))
	for key, q := range repo.storage {
		queues[key] = q
	}
	return queues
}
