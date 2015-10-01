package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/bogdanovich/siberite/queue"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(t, err)

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
	assert.Nil(t, err)

	assert.Equal(t, 3, repo.Count(), "Invalid repo count after initialization")

	for i := 0; i < len(queueNames); i++ {
		q, _ = repo.GetQueue(queueNames[i])
		assert.Equal(t, uint64(1), q.Head(), "Invalid queue initialization")
		assert.Equal(t, uint64(totalItems), q.Tail(), "Invalid queue initialization")
		assert.Equal(t, uint64(totalItems-1), q.Length(), "Invalid queue initialization")
	}
	repo.DeleteAllQueues()
}

func Test_DeleteQueue(t *testing.T) {
	repo, err := Initialize(dir)
	defer repo.DeleteAllQueues()

	assert.Nil(t, err)

	q, _ := repo.GetQueue(name)
	queuePath := q.Path()

	_, err = os.Stat(queuePath)
	assert.Nil(t, err)

	err = repo.DeleteQueue(name)
	assert.Nil(t, err)

	_, err = os.Stat(queuePath)
	assert.NotNil(t, err, "Queue data should not exist")
}

func Test_FullStats(t *testing.T) {
	repo, _ := Initialize(dir)
	defer repo.DeleteAllQueues()

	repo.GetQueue("test1")
	repo.GetQueue("test2")

	statItemKeys := []string{
		"uptime", "time", "version", "curr_connections", "total_connections",
		"cmd_get", "cmd_set", "queue_test2_items", "queue_test2_open_transactions",
		"queue_test1_items", "queue_test1_open_transactions",
	}

	for i, statItem := range repo.FullStats() {
		assert.Equal(t, statItem.Key, statItemKeys[i], "Invalid stats output")
	}
}

func Test_Count(t *testing.T) {
	repo, _ := Initialize(dir)
	defer repo.DeleteAllQueues()

	repo.GetQueue("test1")
	repo.GetQueue("test2")
	assert.Equal(t, 2, repo.Count())
}
