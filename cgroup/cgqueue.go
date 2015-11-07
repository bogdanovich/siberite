package cgroup

import (
	"os"
	"sync"

	"github.com/bogdanovich/siberite/queue"
)

// make sure CGQueue implements Consumer interface
var _ queue.Consumer = (*CGQueue)(nil)

// CGQueue represents queue with multiple consumer groups
type CGQueue struct {
	sync.Mutex
	Name    string
	dataDir string
	*queue.Queue
	*CGManager
}

// CGQueueOpen opens a queue with multiple consumer groups
func CGQueueOpen(name string, dataDir string) (*CGQueue, error) {
	q := &CGQueue{Name: name, dataDir: dataDir + "/" + name}
	return q, q.initialize()
}

// Close closes the queue
func (q *CGQueue) Close() {
	q.CGManager.Close()
	q.Queue.Close()
}

// Drop closes the queue and removes it's data directory
func (q *CGQueue) Drop() {
	q.Close()
	os.RemoveAll(q.Path())
}

// Flush drops all queue data
func (q *CGQueue) Flush() error {
	q.Lock()
	defer q.Unlock()
	q.Drop()
	return q.initialize()
}

// Path returns queue data directory path
func (q *CGQueue) Path() string {
	return q.dataDir
}

func (q *CGQueue) initialize() error {
	var err error
	q.Queue, err = queue.Open(q.Name, q.dataDir, &queue.Options{})
	if err != nil {
		return err
	}

	q.CGManager, err = NewCGManager(q.dataDir+"/_.metadata", q.Queue)
	return err
}
