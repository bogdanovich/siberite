package queue

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dir = "./test_data"
var name = "test"
var err error

func TestMain(m *testing.M) {
	result := m.Run()
	err = os.RemoveAll(dir)
	os.Exit(result)
}

func Test_Open(t *testing.T) {
	q, err := Open(name, dir)
	defer q.Drop()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), q.Head(), "Invalid initial queue state")
	assert.Equal(t, uint64(0), q.Tail(), "Invalid initial queue state")
	assert.Equal(t, uint64(0), q.Length(), "Invalid initial queue state")

	invalidQueueName := "%@#*(&($%@#"
	q2, err := Open(invalidQueueName, dir)
	defer q2.Drop()
	assert.Equal(t, "Queue name is not alphanumeric", err.Error())

	invalidQueueName = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	q3, err := Open(invalidQueueName, dir)
	defer q3.Drop()
	assert.Equal(t, "Queue name is too long", err.Error())
}

func Test_Drop(t *testing.T) {
	q, _ := Open(name, dir)
	q.Drop()
	_, err = os.Stat(q.Path())
	assert.NotNil(t, err, "Path should not exist")
}

func Test_HeadTail(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

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
	q, _ := Open(name, dir)
	defer q.Drop()

	inputValue := "1"
	err = q.Enqueue([]byte(inputValue))
	assert.Nil(t, err)

	for i := 0; i < 3; i++ {
		item, err := q.Peek()
		assert.Nil(t, err)
		assert.Equal(t, inputValue, string(item.Value), "Invalid value")
	}
}

func Test_EnqueueDequeueLength(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

	if q.Length() != 0 {
		t.Error("Invalid initial length")
	}

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

func Test_GetItemByID(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

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
	assert.Equal(t, "Id should be between head and tail", err.Error())
}

func Test_GetItemByOffset(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

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
	assert.Equal(t, "Id should be between head and tail", err.Error())
}

func Test_Prepend(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

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
	q, _ := Open(name, dir)
	defer q.Drop()

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
	q, _ := Open(name, dir)
	defer q.Drop()

	q.AddOpenTransactions(1)
	assert.Equal(t, int64(1), q.Stats.OpenTransactions)
	q.AddOpenTransactions(-1)
	assert.Equal(t, int64(0), q.Stats.OpenTransactions)
}

func Test_initialize(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

	q.Enqueue([]byte("1"))
	q.Enqueue([]byte("2"))
	q.Enqueue([]byte("3"))

	assert.Equal(t, uint64(3), q.Length())

	expectedLength := q.Length()
	expectedHead := q.Head()
	expectedTail := q.Tail()

	// Reopen queue and test if it initializes itself properly
	q.Close()
	q, err = Open(name, dir)
	if err != nil {
		t.Error("Queue initialization error: ", err)
	}

	assert.Equal(t, uint64(expectedLength), q.Length())
	assert.Equal(t, uint64(expectedHead), q.Head())
	assert.Equal(t, uint64(expectedTail), q.Tail())
}

func Test_queuePath(t *testing.T) {
	q, _ := Open("test_queue", dir)
	defer q.Drop()
	assert.Equal(t, "./test_data/test_queue", q.Path())
}
