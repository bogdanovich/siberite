package queue

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dir = "./test_data"
var name = "test"
var options = Options{}
var optionsWithKeyPrefix = Options{KeyPrefix: []byte("queue_key:")}
var err error

func TestMain(m *testing.M) {
	result := m.Run()
	err = os.RemoveAll(dir)
	os.Exit(result)
}

func Test_Open(t *testing.T) {
	q, err := Open(name, dir, &options)
	assert.Nil(t, err)
	testOpen(t, q)
	q.Drop()

	// with KeyPrefix
	q, err = Open("with_prefix", dir, &optionsWithKeyPrefix)
	assert.Nil(t, err)
	testOpen(t, q)
	q.Drop()

	invalidQueueName := "%@#*(&($%@#"
	q, err = Open(invalidQueueName, dir, &options)
	assert.EqualError(t, err, "queue: name is not alphanumeric")
	q.Drop()

	invalidQueueName = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	q, err = Open(invalidQueueName, dir, &options)
	assert.EqualError(t, err, "queue: name is too long")
	q.Drop()
}

func testOpen(t *testing.T, q *Queue) {
	assert.Equal(t, uint64(0), q.Head(), "Invalid initial queue state")
	assert.Equal(t, uint64(0), q.Tail(), "Invalid initial queue state")
	assert.Equal(t, uint64(0), q.Length(), "Invalid initial queue state")
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
}

func testHeadAndTail(t *testing.T, q *Queue) {
	queueLength := 5

	for i := 1; i <= queueLength; i++ {
		_ = q.Enqueue([]byte("1"))
		assert.Equal(t, uint64(0), q.Head())
		assert.Equal(t, uint64(i), q.Tail())
	}

	for i := 1; i <= queueLength; i++ {
		_, _ = q.Dequeue()
		assert.Equal(t, uint64(i), q.Head())
		assert.Equal(t, uint64(queueLength), q.Tail())
	}
}

func Test_Peek(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testPeek(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testPeek(t, q)
	q.Drop()
}

func testPeek(t *testing.T, q *Queue) {
	inputValue := "1"
	err = q.Enqueue([]byte(inputValue))
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		item, err := q.Peek()
		assert.Nil(t, err)
		assert.Equal(t, inputValue, string(item.Value), "Invalid value")
	}
}

func Test_EnqueueDequeueLength(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testEnqueueDequeueLength(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testEnqueueDequeueLength(t, q)
	q.Drop()
}

func testEnqueueDequeueLength(t *testing.T, q *Queue) {
	assert.NotEqual(t, 0, q.Length(), "Invalid initial length")

	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		values[i] = strconv.Itoa(i)
	}

	for i := 0; i < len(values); i++ {
		assert.Equal(t, uint64(i), q.Length())
		q.Enqueue([]byte(values[i]))
	}

	for i := 0; i < len(values); i++ {
		item, err := q.Dequeue()
		assert.Nil(t, err)
		assert.Equal(t, values[i], string(item.Value), "Invalid value")
	}
}

func Test_GetNext(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testGetNext(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testGetNext(t, q)
	q.Drop()
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

	assert.Equal(t, uint64(2), q.Head())

	err = q.PutBack(value)
	assert.NoError(t, err)

	assert.Equal(t, uint64(1), q.Head())

	// Check that we get the same item with the next GetNext
	value, _ = q.GetNext()
	assert.Equal(t, "2", string(value))
}

func Test_GetItemByID(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testGetItemByID(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testGetItemByID(t, q)
	q.Drop()
}

func testGetItemByID(t *testing.T, q *Queue) {
	values := []string{"1", "2", "3", "4"}
	for i := 0; i < len(values); i++ {
		q.Enqueue([]byte(values[i]))
	}

	item, err := q.GetItemByID(q.Head() + 1)
	assert.Nil(t, err)
	assert.Equal(t, "1", string(item.Value))

	item, err = q.GetItemByID(q.Head() + 3)
	assert.Nil(t, err)
	assert.Equal(t, "3", string(item.Value))

	item, err = q.GetItemByID(q.Head() + 5)
	assert.Equal(t, "queue: out of bounds", err.Error())
}

func Test_GetItemByOffset(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testGetItemByOffset(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testGetItemByOffset(t, q)
	q.Drop()
}

func testGetItemByOffset(t *testing.T, q *Queue) {
	values := []string{"1", "2", "3", "4"}
	for i := 0; i < len(values); i++ {
		q.Enqueue([]byte(values[i]))
	}

	item, err := q.GetItemByOffset(0)
	assert.Nil(t, err)
	assert.Equal(t, "1", string(item.Value))

	item, err = q.GetItemByOffset(2)
	assert.Nil(t, err)
	assert.Equal(t, "3", string(item.Value))

	item, err = q.GetItemByOffset(5)
	assert.Equal(t, "queue: out of bounds", err.Error())
}

func Test_Prepend(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testPrepend(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testPrepend(t, q)
	q.Drop()
}

func testPrepend(t *testing.T, q *Queue) {
	values := []string{"1", "2", "3", "4", "5"}
	for i := 0; i < len(values); i++ {
		q.Enqueue([]byte(values[i]))
	}

	item, _ := q.Dequeue()
	q.Dequeue()

	assert.Equal(t, uint64(2), q.Head())

	err = q.Prepend(item)
	assert.Nil(t, err)

	assert.Equal(t, uint64(1), q.Head())

	// Check that we get the same item with the next Dequeue
	item, _ = q.Dequeue()
	assert.Equal(t, "1", string(item.Value))
}

func Test_Length(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testLength(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testLength(t, q)
	q.Drop()
}

func testLength(t *testing.T, q *Queue) {
	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		values[i] = strconv.Itoa(i)
	}

	for i := 0; i < len(values); i++ {
		assert.Equal(t, uint64(i), q.Length())
		q.Enqueue([]byte(values[i]))
	}

	for i := len(values); i < 0; i-- {
		_, err := q.Dequeue()
		assert.Nil(t, err)
		assert.Equal(t, uint64(i), q.Length())
	}
}

func Test_AddOpenTransactions(t *testing.T) {
	q, _ := Open(name, dir, &options)
	testAddOpenTransactions(t, q)
	q.Drop()

	q, _ = Open(name, dir, &optionsWithKeyPrefix)
	testAddOpenTransactions(t, q)
	q.Drop()
}

func testAddOpenTransactions(t *testing.T, q *Queue) {
	q.AddOpenTransactions(1)
	assert.Equal(t, int64(1), q.Stats.OpenTransactions)
	q.AddOpenTransactions(-1)
	assert.Equal(t, int64(0), q.Stats.OpenTransactions)
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

	assert.Equal(t, uint64(3), q.Length())

	expectedLength := q.Length()
	expectedHead := q.Head()
	expectedTail := q.Tail()

	// Reopen queue and test if it initializes itself properly
	q.Close()
	q, err = Open(name, dir, opts)
	if err != nil {
		t.Error("Queue initialization error: ", err)
	}

	assert.Equal(t, uint64(expectedLength), q.Length())
	assert.Equal(t, uint64(expectedHead), q.Head())
	assert.Equal(t, uint64(expectedTail), q.Tail())
}

func Test_queuePath(t *testing.T) {
	q, _ := Open("test_queue", dir, &options)
	defer q.Drop()
	assert.Equal(t, "./test_data/test_queue", q.Path())
}
