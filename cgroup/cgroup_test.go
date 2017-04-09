package cgroup

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/bogdanovich/siberite/queue"
)

var dir = "./test_data"
var sourceQueueName = "source"
var sourceQueuePath = dir + "/" + sourceQueueName
var storageDBName = "storage"
var storageDBPath = dir + "/" + storageDBName
var cgName = "test"
var options = queue.Options{}
var err error

func TestMain(m *testing.M) {
	os.RemoveAll(dir)
	result := m.Run()
	os.RemoveAll(dir)
	os.Exit(result)
}

func setupConsumerGroup(t *testing.T, name string,
	numItems int) (*ConsumerGroup, error) {

	storage, err := leveldb.OpenFile(storageDBPath, &opt.Options{})
	assert.NoError(t, err)
	source, err := queue.Open(sourceQueueName, dir, &options)
	assert.NoError(t, err)

	for i := 0; i < numItems+1; i++ {
		source.Enqueue([]byte(strconv.Itoa(i)))
	}

	source.GetNext()
	return NewConsumerGroup(name, source, storage)
}

func cleanupConsumerGroup(cg *ConsumerGroup) {
	cg.source.Close()
	cg.storage.Close()
	os.RemoveAll(dir)
}

func Test_ConsumerGroup_initialize(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 10)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cg.cursor)
	assert.Equal(t, "_c:"+cgName, string(cg.cursorKey))
	assert.EqualValues(t, 0, cg.failedReads.Head())
	assert.EqualValues(t, 0, cg.failedReads.Tail())
	assert.EqualValues(t, 0, cg.failedReads.Length())
	cleanupConsumerGroup(cg)

	cg, err = setupConsumerGroup(t, "a-1", 10)
	defer cleanupConsumerGroup(cg)
	assert.EqualError(t, err, "cgroup: name is not alphanumeric")
}

func Test_ConsumerGroup_GetNextAndPutBack(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 5)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)

	value, err := cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "1", string(value))

	// get 3 items from source and check next
	// returned item from the counsumer group
	cg.Source().GetNext()
	cg.Source().GetNext()
	value, err = cg.Source().GetNext()
	assert.Equal(t, "3", string(value))

	value, err = cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "4", string(value))

	value, err = cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "5", string(value))

	assert.True(t, cg.failedReads.IsEmpty())
	err = cg.PutBack([]byte("4"))
	assert.NoError(t, err)
	assert.False(t, cg.failedReads.IsEmpty())

	cg.PutBack([]byte("5"))

	value, err = cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "4", string(value))

	value, err = cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "5", string(value))

	value, err = cg.GetNext()
	assert.EqualError(t, err, "queue: ID is out of bounds")
}

func Test_ConsumerGroup_Peek(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 10)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)

	for i := 0; i < 2; i++ {
		value, err := cg.Peek()
		assert.NoError(t, err)
		assert.Equal(t, "1", string(value))
	}
	// get 3 items from source and check next
	// returned item from the counsumer group
	cg.Source().GetNext()
	cg.Source().GetNext()
	value, err := cg.Source().GetNext()
	assert.Equal(t, "3", string(value))

	for i := 0; i < 2; i++ {
		value, err := cg.Peek()
		assert.NoError(t, err)
		assert.Equal(t, "4", string(value))
	}

	assert.True(t, cg.failedReads.IsEmpty())
	err = cg.PutBack([]byte("2"))
	assert.NoError(t, err)
	assert.False(t, cg.failedReads.IsEmpty())

	cg.PutBack([]byte("3"))

	for i := 0; i < 2; i++ {
		value, err := cg.Peek()
		assert.NoError(t, err)
		assert.Equal(t, "2", string(value))
	}
}

func Test_ConsumerGroup_Length(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 10)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)

	value, err := cg.GetNext()
	assert.NoError(t, err)
	assert.Equal(t, "1", string(value))

	// get 3 items from source and check next
	// returned item from the counsumer group
	cg.Source().GetNext()
	cg.Source().GetNext()
	value, err = cg.Source().GetNext()
	assert.Equal(t, "3", string(value))

	assert.EqualValues(t, cg.Source().Length(), cg.Length())

	err = cg.PutBack([]byte("1"))
	assert.NoError(t, err)
	assert.EqualValues(t, cg.Source().Length()+1, cg.Length())
}

func Test_ConsumerGroup_IsEmpty(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 0)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)
	assert.True(t, cg.IsEmpty())
	cg.source.Enqueue([]byte("1"))
	assert.False(t, cg.IsEmpty())
	value, err := cg.GetNext()
	assert.NoError(t, err)
	assert.True(t, cg.IsEmpty())
	err = cg.PutBack(value)
	assert.NoError(t, err)
	assert.False(t, cg.IsEmpty())
}

func Test_ConsumerGroup_Flush(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 10)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)
	assert.EqualValues(t, 10, cg.Length())

	value, err := cg.GetNext()
	assert.NoError(t, err)

	cg.PutBack(value)
	assert.EqualValues(t, 10, cg.Length())

	cg.Flush()
	assert.EqualValues(t, 10, cg.Length())
}

func Test_ConsumerGroup_Delete(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 10)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)
	assert.EqualValues(t, 10, cg.Length())

	value, err := cg.GetNext()
	assert.NoError(t, err)

	cg.PutBack(value)
	assert.EqualValues(t, 10, cg.Length())

	cg.Delete()
	assert.EqualValues(t, 10, cg.Length())
}

func Test_ConsumerGroup_Stats(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 0)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)
	stats := cg.Stats()
	assert.EqualValues(t, 0, stats.OpenReads)
	stats.UpdateOpenReads(1)
	assert.EqualValues(t, 1, stats.OpenReads)
}

func Test_ConsumerGroup_loadAndSaveCursor(t *testing.T) {
	cg, err := setupConsumerGroup(t, cgName, 10)
	defer cleanupConsumerGroup(cg)
	assert.NoError(t, err)

	assert.EqualValues(t, 1, cg.cursor)
	cg.cursor++
	// restores 0 because wasn't saved
	err = cg.loadCursor()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cg.cursor)

	cg.cursor++
	err = cg.updateCursor(cg.cursor)
	assert.NoError(t, err)
	err = cg.loadCursor()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, cg.cursor)
}
