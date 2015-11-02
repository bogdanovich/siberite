package cgroup

import (
	"os"
	"strconv"
	"testing"

	"github.com/bogdanovich/siberite/queue"
	"github.com/stretchr/testify/assert"
)

var storagePath = dir + "/consumers"

func setupCGManager(t *testing.T, numItems int) (*CGManager, error) {
	source, err := queue.Open(sourceQueueName, dir, &options)
	assert.NoError(t, err)

	for i := 0; i < numItems; i++ {
		source.Enqueue([]byte(strconv.Itoa(i)))
	}

	// dequeue first element so head is not 0
	// remaing queue size is 9
	source.GetNext()
	return NewCGManager(storagePath, source)
}

func cleanupCGManager(m *CGManager) {
	m.Close()
	m.source.Close()
	os.RemoveAll(dir)
}

func Test_NewCGManager(t *testing.T) {
	m, err := setupCGManager(t, 10)
	defer cleanupCGManager(m)
	assert.NoError(t, err)
	assert.False(t, m.source.IsEmpty())
	assert.EqualValues(t, 9, m.source.Length())
}

func Test_ConsumerGroupAndIteration(t *testing.T) {
	m, err := setupCGManager(t, 10)
	defer cleanupCGManager(m)
	assert.NoError(t, err)

	var names = []string{"work", "check", "consumer", "1"}
	for _, name := range names {
		cg, err := m.ConsumerGroup(name)
		assert.NoError(t, err)
		assert.EqualValues(t, 9, cg.Length())
		assert.False(t, cg.IsEmpty())

		value, err := cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "1", string(value))

		value, err = cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "2", string(value))

		assert.EqualValues(t, 7, cg.Length())
		assert.EqualValues(t, 9, cg.Source().Length())
	}

	for item := range m.ConsumerGroupIterator() {
		assert.NotNil(t, item.Val)
		cg := item.Val.(queue.Consumer)
		assert.EqualValues(t, 7, cg.Length())
		value, err := cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "3", string(value))
	}

}
