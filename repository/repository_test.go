package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/bogdanovich/siberite/queue"
)

var dir = "./test_data"
var name = "test"
var err error

func TestMain(m *testing.M) {
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Println(err)
	}
	result := m.Run()
	err = os.RemoveAll(dir)
	os.Exit(result)
}

func Test_Initialize(t *testing.T) {
	repo, err := Initialize(dir)
	if err != nil {
		t.Error(err)
	}

	// Create 3 queues and push some data
	queueNames := []string{"test1", "test2", "test3"}
	var q *queue.Queue
	totalItems := 3
	for i := 0; i < len(queueNames); i++ {
		q, _ = repo.GetQueue(queueNames[i])
		for j := 0; j < totalItems; j++ {
			q.Enqueue([]byte("value"))
		}
		// Get one element out
		_, _ = q.Dequeue()
	}

	// Close all queues and destroy repo
	repo.CloseAllQueues()
	repo = nil

	// Initialize repo again and check loaded queues
	repo, err = Initialize(dir)
	if err != nil {
		t.Error(err)
	}

	if repo.Count() != 3 {
		t.Errorf("Invalid repo count after initialization (%d instead of %d", repo.Count(), len(queueNames))
	}

	for i := 0; i < len(queueNames); i++ {
		q, _ = repo.GetQueue(queueNames[i])
		if q.Head() != 1 || q.Tail() != uint64(totalItems) || q.Length() != uint64(totalItems-1) {
			t.Errorf("Invalid queue initialization")
		}
	}
	repo.DeleteAllQueues()
}

func Test_DeleteQueue(t *testing.T) {
	repo, err := Initialize(dir)
	defer repo.DeleteAllQueues()

	if err != nil {
		t.Error(err)
	}

	q, _ := repo.GetQueue(name)
	queuePath := q.Path()
	if _, err := os.Stat(queuePath); os.IsNotExist(err) {
		t.Error("Queue data dir does not exist: ", err)
	}

	err = repo.DeleteQueue(name)
	if err != nil {
		t.Error("Delete queue error: ", err)
	}

	if _, err := os.Stat(queuePath); err == nil {
		t.Error("Queue data should not exist")
	}
}

func Test_FullStats(t *testing.T) {
	repo, _ := Initialize(dir)
	defer repo.DeleteAllQueues()

	repo.GetQueue("test1")
	repo.GetQueue("test2")

	statItemKeys := []string{
		"uptime", "time", "version", "curr_connections", "total_connections",
		"cmd_get", "cmd_set", "queue_test2_items", "queue_test1_items",
	}

	for i, statItem := range repo.FullStats() {
		if statItemKeys[i] != statItem.Key {
			t.Error("Invalid stats output")
		}
	}
}
