package cgroup

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bogdanovich/siberite/queue"
)

var storagePath = dir + "/consumers"

func setupCGManager(t *testing.T, numItems int) (*CGManager, error) {
	source, err := queue.Open(sourceQueueName, dir, &options)
	assert.NoError(t, err)

	for i := 0; i < numItems+1; i++ {
		source.Enqueue([]byte(strconv.Itoa(i)))
	}

	// dequeue first element so head is not 0
	source.GetNext()
	return NewCGManager(storagePath, source)
}

func cleanupCGManager(m *CGManager) {
	m.Close()
	m.source.Close()
	os.RemoveAll(dir)
}

func Test_CGManager_NewCGManager(t *testing.T) {
	m, err := setupCGManager(t, 10)
	defer cleanupCGManager(m)
	assert.NoError(t, err)
	assert.False(t, m.source.IsEmpty())
	assert.EqualValues(t, 10, m.source.Length())
}

func Test_CGManager_initialize(t *testing.T) {
	m, err := setupCGManager(t, 10)
	defer cleanupCGManager(m)
	assert.NoError(t, err)
	assert.EqualValues(t, 10, m.source.Length())

	// Create 5 consumer groups and read some values from each
	for i := 1; i < 5; i++ {
		cg, err := m.ConsumerGroup(strconv.Itoa(i))
		assert.NoError(t, err)
		for j := 1; j <= i; j++ {
			cg.GetNext()
		}
	}

	// Re-open consumer group manager again and check for initialized values
	m.Close()
	m, err = NewCGManager(storagePath, m.source)
	assert.NoError(t, err)

	for item := range m.ConsumerGroupIterator() {
		assert.NotNil(t, item.Val)
		cg := item.Val.(*ConsumerGroup)
		i, _ := strconv.Atoi(item.Key)
		assert.EqualValues(t, 10-i, cg.Length())
		i++
	}
}

func Test_CGManager_ConsumerGroupAndIteration(t *testing.T) {
	m, err := setupCGManager(t, 10)
	defer cleanupCGManager(m)
	assert.NoError(t, err)

	var names = []string{"work", "check", "consumer", "1"}
	for _, name := range names {
		cg, err := m.ConsumerGroup(name)
		assert.NoError(t, err)
		assert.EqualValues(t, 10, cg.Length())
		assert.False(t, cg.IsEmpty())

		value, err := cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "1", string(value))

		value, err = cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "2", string(value))

		assert.EqualValues(t, 8, cg.Length())
		assert.EqualValues(t, 10, cg.Source().Length())
	}

	for item := range m.ConsumerGroupIterator() {
		assert.NotNil(t, item.Val)
		cg := item.Val.(queue.Consumer)
		assert.EqualValues(t, 8, cg.Length())
		value, err := cg.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "3", string(value))
	}

}

func Test_CGManager_DeleteConsumerGroup(t *testing.T) {
	m, err := setupCGManager(t, 10)
	defer cleanupCGManager(m)
	assert.NoError(t, err)

	cg, err := m.ConsumerGroup("test_cgroup")
	assert.NoError(t, err)
	assert.EqualValues(t, 10, cg.Length())
	assert.False(t, cg.IsEmpty())

	value, err := cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "1", string(value))

	value, err = cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "2", string(value))

	err = m.DeleteConsumerGroup("test_cgroup")
	assert.NoError(t, err)

	for _ = range m.ConsumerGroupIterator() {
		assert.Fail(t, "should not have any consumer groups")
	}

	cg, err = m.ConsumerGroup("test_cgroup")
	assert.NoError(t, err)
	assert.EqualValues(t, 10, cg.Length())
}
