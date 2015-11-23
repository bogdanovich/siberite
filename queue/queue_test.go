package queue

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var dir = "./test_data"
var name = "test"
var sharedDBName = "test_shared"
var sharedDBPath = dir + "/" + sharedDBName
var options = Options{}
var optionsWithKeyPrefix = Options{KeyPrefix: []byte("queue_key:")}
var err error

type testQueue func(*Queue)

func withSharedQueues(t *testing.T, fn testQueue) {
	db, err := leveldb.OpenFile(sharedDBPath, &opt.Options{})
	prefixes := []string{"prefix1:", "prefix2:", "prefix3:"}
	queues := make(map[int]*Queue)
	for i, keyPrefix := range prefixes {
		queues[i], err = OpenShared(name, keyPrefix, db)
		assert.NoError(t, err)
		fn(queues[i])
	}
	db.Close()
	os.RemoveAll(dir)
}

func TestMain(m *testing.M) {
	os.RemoveAll(dir)
	result := m.Run()
	os.RemoveAll(dir)
	os.Exit(result)
}

func Test_ValidQueueName(t *testing.T) {
	invalidQueueName := "%@#*(&($%@#"
	q, err := Open(invalidQueueName, dir, &options)
	assert.EqualError(t, err, "queue: name is not alphanumeric")
	q.Drop()

	validQueueNames := []string{"test-name-1", "test_name_2"}
	for _, queueName := range validQueueNames {
		q, err = Open(queueName, dir, &options)
		assert.NoError(t, err)
		q.Drop()
	}
}

func Test_Open(t *testing.T) {
	invalidQueueName := "%@#*(&($%@#"
	q, err := Open(invalidQueueName, dir, &options)
	assert.EqualError(t, err, "queue: name is not alphanumeric")
	q.Drop()

	invalidQueueName = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	q, err = Open(invalidQueueName, dir, &options)
	assert.EqualError(t, err, "queue: name is too long")
	q.Drop()

	q, err = Open(name, dir, &options)
	assert.NoError(t, err)
	testOpen(t, q)
	q.Drop()

	// with KeyPrefix
	q, err = Open("with_prefix", dir, &optionsWithKeyPrefix)
	assert.NoError(t, err)
	testOpen(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testOpen(t, q)
	})

}

func testOpen(t *testing.T, q *Queue) {
	assert.EqualValues(t, 0, q.Head(), "Invalid initial queue state")
	assert.EqualValues(t, 0, q.Tail(), "Invalid initial queue state")
	assert.EqualValues(t, 0, q.Length(), "Invalid initial queue state")
}

func Test_Drop(t *testing.T) {
	q, _ := Open(name, dir, &options)
	q.Drop()
	_, err = os.Stat(q.Path())
	assert.NotNil(t, err, "Path should not exist")
}

func Test_HeadAndTail(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testHeadAndTail(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testHeadAndTail(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testHeadAndTail(t, q)
	})
}

func testHeadAndTail(t *testing.T, q *Queue) {
	queueLength := 5

	for i := 1; i <= queueLength; i++ {
		_ = q.Enqueue([]byte("1"))
		assert.EqualValues(t, 0, q.Head())
		assert.EqualValues(t, i, q.Tail())
	}

	for i := 1; i <= queueLength; i++ {
		_, _ = q.GetNext()
		assert.EqualValues(t, i, q.Head())
		assert.EqualValues(t, queueLength, q.Tail())
	}
}

func Test_Peek(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testPeek(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testPeek(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testPeek(t, q)
	})
}

func testPeek(t *testing.T, q *Queue) {
	inputValue := "1"
	err = q.Enqueue([]byte(inputValue))
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		value, err := q.Peek()
		assert.Nil(t, err)
		assert.Equal(t, inputValue, string(value), "Invalid value")
	}
}

func Test_EnqueueDequeueLength(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testEnqueueDequeueLength(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testEnqueueDequeueLength(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testEnqueueDequeueLength(t, q)
	})
}

func testEnqueueDequeueLength(t *testing.T, q *Queue) {
	assert.NotEqual(t, 0, q.Length(), "Invalid initial length")

	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		values[i] = strconv.Itoa(i)
	}

	for i := 0; i < len(values); i++ {
		assert.EqualValues(t, i, q.Length())
		q.Enqueue([]byte(values[i]))
	}

	for i := 0; i < len(values); i++ {
		value, err := q.GetNext()
		assert.Nil(t, err)
		assert.Equal(t, values[i], string(value), "Invalid value")
	}
}

func Test_GetNext(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testGetNext(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testGetNext(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testGetNext(t, q)
	})
}

func testGetNext(t *testing.T, q *Queue) {
	values := []string{"1", "2"}
	for i := 0; i < len(values); i++ {
		q.Enqueue([]byte(values[i]))
	}

	value, err := q.GetNext()
	assert.Nil(t, err)
	assert.Equal(t, "1", string(value))

	value, err = q.GetNext()
	assert.Nil(t, err)
	assert.Equal(t, "2", string(value))

	value, err = q.GetNext()
	assert.EqualError(t, err, "queue: is empty")
}

func Test_PutBack(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testPutBack(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testPutBack(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testPutBack(t, q)
	})
}

func testPutBack(t *testing.T, q *Queue) {
	values := []string{"1", "2"}
	for i := 0; i < len(values); i++ {
		q.Enqueue([]byte(values[i]))
	}

	err = q.PutBack([]byte("0"))
	assert.EqualError(t, err, "queue: head can not be less then zero")

	// get 1
	q.GetNext()
	// get 2
	value, err := q.GetNext()
	assert.NoError(t, err)

	assert.EqualValues(t, 2, q.Head())

	err = q.PutBack(value)
	assert.NoError(t, err)

	assert.EqualValues(t, 1, q.Head())

	// Check that we get the same item with the next GetNext
	value, _ = q.GetNext()
	assert.Equal(t, "2", string(value))
}

func Test_ReadItemByID(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testReadItemByID(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testReadItemByID(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testReadItemByID(t, q)
	})
}

func testReadItemByID(t *testing.T, q *Queue) {
	values := []string{"1", "2", "3", "4"}
	for i := 0; i < len(values); i++ {
		q.Enqueue([]byte(values[i]))
	}

	item, err := q.ReadItemByID(q.Head() + 1)
	assert.Nil(t, err)
	assert.EqualValues(t, q.Head()+1, item.ID)
	assert.Equal(t, "1", string(item.Value))

	item, err = q.ReadItemByID(q.Head() + 3)
	assert.Nil(t, err)
	assert.EqualValues(t, q.Head()+3, item.ID)
	assert.Equal(t, "3", string(item.Value))

	item, err = q.ReadItemByID(q.Head() + 5)
	assert.Equal(t, "queue: ID is out of bounds", err.Error())
}

func Test_ReadItemByOffset(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testReadItemByOffset(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testReadItemByOffset(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testReadItemByOffset(t, q)
	})
}

func testReadItemByOffset(t *testing.T, q *Queue) {
	values := []string{"1", "2", "3", "4"}
	for i := 0; i < len(values); i++ {
		q.Enqueue([]byte(values[i]))
	}

	item, err := q.ReadItemByOffset(0)
	assert.Nil(t, err)
	assert.Equal(t, q.Head()+1, item.ID)
	assert.Equal(t, "1", string(item.Value))

	item, err = q.ReadItemByOffset(2)
	assert.Nil(t, err)
	assert.Equal(t, q.Head()+1+2, item.ID)
	assert.Equal(t, "3", string(item.Value))

	item, err = q.ReadItemByOffset(5)
	assert.Equal(t, "queue: ID is out of bounds", err.Error())
}

func Test_LengthAndIsEmpty(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testLengthAndIsEmpty(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testLengthAndIsEmpty(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testLengthAndIsEmpty(t, q)
	})
}

func testLengthAndIsEmpty(t *testing.T, q *Queue) {
	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		values[i] = strconv.Itoa(i)
	}
	assert.True(t, q.IsEmpty())

	for i := 0; i < len(values); i++ {
		assert.EqualValues(t, i, q.Length())
		q.Enqueue([]byte(values[i]))
	}

	for i := len(values); i > 0; i-- {
		assert.EqualValues(t, i, q.Length())
		_, err := q.GetNext()
		assert.Nil(t, err)
	}
	assert.True(t, q.IsEmpty())
}

func Test_DeleteAll(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testDeleteAll(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testDeleteAll(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testDeleteAll(t, q)
	})
}

func testDeleteAll(t *testing.T, q *Queue) {
	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		values[i] = strconv.Itoa(i)
	}

	for i := 0; i < len(values); i++ {
		assert.EqualValues(t, i, q.Length())
		q.Enqueue([]byte(values[i]))
	}

	assert.EqualValues(t, 10, q.Length())
	err = q.DeleteAll()
	assert.NoError(t, err)
	assert.True(t, q.IsEmpty())
}

func Test_Flush(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testFlush(t, q, false)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testFlush(t, q, false)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testFlush(t, q, true)
	})
}

func testFlush(t *testing.T, q *Queue, isShared bool) {
	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		values[i] = strconv.Itoa(i)
	}

	for i := 0; i < len(values); i++ {
		assert.EqualValues(t, i, q.Length())
		q.Enqueue([]byte(values[i]))
	}

	assert.EqualValues(t, 10, q.Length())
	err = q.Flush()
	if !isShared {
		assert.NoError(t, err)
		assert.EqualValues(t, 0, q.Length())
		assert.True(t, q.IsEmpty())
	} else {
		assert.EqualError(t, err, "queue: can't flush shared queue")
		assert.EqualValues(t, 10, q.Length())
		assert.False(t, q.IsEmpty())
	}

}

func Test_Stats(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testStats(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testStats(t, q)
	q.Drop()

	withSharedQueues(t, func(q *Queue) {
		testStats(t, q)
	})
}

func testStats(t *testing.T, q *Queue) {
	stats := q.Stats()
	assert.EqualValues(t, 0, stats.OpenReads)
	stats.UpdateOpenReads(1)
	assert.EqualValues(t, 1, stats.OpenReads)
}

func Test_initialize(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testInitialize(t, q, &options)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testInitialize(t, q, &optionsWithKeyPrefix)
	q.Drop()
}

func testInitialize(t *testing.T, q *Queue, opts *Options) {
	q.Enqueue([]byte("1"))
	q.Enqueue([]byte("2"))
	q.Enqueue([]byte("3"))

	assert.EqualValues(t, 3, q.Length())

	expectedLength := q.Length()
	expectedHead := q.Head()
	expectedTail := q.Tail()

	// Reopen queue and test if it initializes itself properly
	q.Close()
	q, err = Open(name, dir, opts)
	if err != nil {
		t.Error("Queue initialization error: ", err)
	}

	assert.EqualValues(t, expectedLength, q.Length())
	assert.EqualValues(t, expectedHead, q.Head())
	assert.EqualValues(t, expectedTail, q.Tail())
}

func Test_initializeShared(t *testing.T) {
	// store some data
	db, err := leveldb.OpenFile(sharedDBPath, &opt.Options{})
	prefixes := []string{"prefix1:", "prefix2:", "prefix3:"}
	queues := make(map[int]*Queue)
	for i, keyPrefix := range prefixes {
		queues[i], err = OpenShared(name, keyPrefix, db)
		assert.NoError(t, err)

		queues[i].Enqueue([]byte("1"))
		queues[i].Enqueue([]byte("2"))
		queues[i].Enqueue([]byte("3"))
		queues[i].Enqueue([]byte("4"))

		// read one value
		queues[i].GetNext()

		assert.EqualValues(t, 3, queues[i].Length())
	}
	db.Close()

	// initialize again
	db, err = leveldb.OpenFile(sharedDBPath, &opt.Options{})
	prefixes = []string{"prefix1:", "prefix2:", "prefix3:"}
	queues = make(map[int]*Queue)
	for i, keyPrefix := range prefixes {
		queues[i], err = OpenShared(name, keyPrefix, db)
		assert.NoError(t, err)

		assert.EqualValues(t, 3, queues[i].Length())
		assert.EqualValues(t, 1, queues[i].Head())
		assert.EqualValues(t, 4, queues[i].Tail())
	}
	db.Close()

}

func Test_queuePath(t *testing.T) {
	q, _ := Open("test_queue", dir, &options)
	defer q.Drop()
	assert.Equal(t, "./test_data/test_queue", q.Path())
}
