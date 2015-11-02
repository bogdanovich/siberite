package cgroup

import (
	"os"

	"github.com/bogdanovich/siberite/queue"
)

// make sure CGQueue implements Consumer interface
var _ queue.Consumer = (*CGQueue)(nil)

// CGQueue represents queue with multiple consumer groups
type CGQueue struct {
	*queue.Queue
	*CGManager
	dataDir string
}

// CGQueueOpen opens a queue with multiple consumer groups
func CGQueueOpen(name string, dataDir string) (*CGQueue, error) {
	dir := dataDir + "/" + name

	sourceQueue, err := queue.Open(name, dir, &queue.Options{})
	if err != nil {
		return nil, err
	}

	cgManager, err := NewCGManager(dir+"/consumers", sourceQueue)
	if err != nil {
		return nil, err
	}

	return &CGQueue{Queue: sourceQueue, CGManager: cgManager, dataDir: dir}, err
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

// Path returns queue data directory path
func (q *CGQueue) Path() string {
	return q.dataDir
}
