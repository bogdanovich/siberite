package cgroup

import (
	"os"
	"strconv"
	"testing"

	"github.com/bogdanovich/siberite/queue"
	"github.com/stretchr/testify/assert"
)

var cgQueueName = "cgqueue"

func setupCGQueue(t *testing.T, numItems int) (*CGQueue, error) {
	q, err := CGQueueOpen(cgQueueName, dir)
	assert.NoError(t, err)

	for i := 0; i < numItems+1; i++ {
		q.Enqueue([]byte(strconv.Itoa(i)))
	}

	q.GetNext()
	return q, nil
}

func cleanupCGQueue(q *CGQueue) {
	q.Close()
	os.RemoveAll(dir)
}

func Test_CGQueueOpen(t *testing.T) {
	q, err := setupCGQueue(t, 10)
	defer cleanupCGQueue(q)
	assert.NoError(t, err)
	assert.EqualValues(t, 10, q.Length())
}

func Test_Queue(t *testing.T) {
	q, err := setupCGQueue(t, 10)
	defer cleanupCGQueue(q)
	assert.NoError(t, err)
	assert.EqualValues(t, 10, q.Length())

	value, err := q.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "1", string(value))

	value, err = q.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "2", string(value))

	err = q.PutBack([]byte("1"))
	assert.NoError(t, err)

	value, err = q.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "1", string(value))

}

func Test_ConsumerGroups(t *testing.T) {
	q, err := setupCGQueue(t, 4)
	defer cleanupCGQueue(q)
	assert.NoError(t, err)
	assert.EqualValues(t, 4, q.Length())

	value, err := q.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "1", string(value))

	var names = []string{"work", "check", "consumer", "1"}
	for _, name := range names {
		cg, err := q.ConsumerGroup(name)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, cg.Length())
		assert.False(t, cg.IsEmpty())

		value, err := cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "2", string(value))

		value, err = cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "3", string(value))

		assert.EqualValues(t, 1, cg.Length())
		assert.EqualValues(t, 3, cg.Source().Length())
	}

	for item := range q.ConsumerGroupIterator() {
		assert.NotNil(t, item.Val)
		cg := item.Val.(queue.Consumer)
		assert.EqualValues(t, 1, cg.Length())
		err := cg.PutBack([]byte("2"))
		assert.NoError(t, err)

		value, err := cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "2", string(value))

		value, err = cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "4", string(value))
		assert.EqualValues(t, 0, cg.Length())

		value, err = cg.GetNext()
		assert.EqualError(t, err, "queue: ID is out of bounds")
		assert.True(t, cg.IsEmpty())
		assert.EqualValues(t, 0, cg.Length())
	}

}

func Test_Path(t *testing.T) {
	q, err := setupCGQueue(t, 10)
	defer cleanupCGQueue(q)
	assert.NoError(t, err)
	assert.Equal(t, "./test_data/cgqueue", q.Path())
}

func Test_Drop(t *testing.T) {
	q, err := setupCGQueue(t, 10)
	assert.NoError(t, err)
	q.Drop()
	_, err = os.Stat(q.Path())
	assert.NotNil(t, err, "Path should not exist")
}
