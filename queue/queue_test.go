package queue

import (
	"fmt"
	"os"
	"strconv"
	"testing"
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
	if err != nil {
		t.Error(err)
	}
	if q.Head() != 0 || q.Tail() != 0 || q.Length() != 0 {
		t.Error("Invalid initial queue state")
	}

	invalidQueueName := "%@#*(&($%@#"
	q2, err := Open(invalidQueueName, dir)
	defer q2.Drop()
	if err.Error() != "Queue name is not alphanumeric" {
		t.Error("Expected invalid queue name error")
	}

	invalidQueueName = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	q3, err := Open(invalidQueueName, dir)
	defer q3.Drop()
	if err.Error() != "Queue name is too long" {
		t.Error("Expected invalid queue name error")
	}
}

func Test_Drop(t *testing.T) {
	q, _ := Open(name, dir)
	q.Drop()
	if _, err := os.Stat(q.Path()); err == nil {
		t.Error("Path should not exist")
	}
}

func Test_HeadTail(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

	queueLength := 5

	for i := 1; i <= queueLength; i++ {
		_ = q.Enqueue([]byte("1"))
		if q.Head() != 0 || q.Tail() != uint64(i) {
			t.Errorf("Unexpected head or tail (0,%d) (%d, %d)", i, q.Head(), q.Tail())
		}
	}

	for i := 1; i <= queueLength; i++ {
		_, _ = q.Dequeue()
		if q.Head() != uint64(i) || q.Tail() != uint64(queueLength) {
			t.Errorf("Unexpected head or tail (%d,%d) (%d, %d)", i, queueLength, q.Head(), q.Tail())
		}
	}

}

func Test_Peek(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

	inputValue := "1"
	err = q.Enqueue([]byte(inputValue))
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		item, err := q.Peek()
		if err != nil {
			t.Error(err)
		}
		if string(item.Value) != inputValue {
			t.Error("Got invalid value")
		}
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
		if count := q.Length(); i != int(count) {
			t.Error("Invalid number of items in queue")
		}
		q.Enqueue([]byte(values[i]))
	}

	for i := 0; i < len(values); i++ {
		item, err := q.Dequeue()
		if err != nil {
			t.Error(err)
		}
		if string(item.Value) != values[i] {
			t.Error("Got invalid value")
		}
	}

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

	if q.Head() != 2 {
		t.Error("Expected head = %d, actual = %d", 2, q.Head())
	}

	err = q.Prepend(item)
	if err != nil {
		t.Error(err)
	}

	if q.Head() != 1 {
		t.Error("Expected head = %d, actual = %d", 1, q.Head())
	}

	// Check that we get the same item with the next Dequeue
	item, _ = q.Dequeue()
	if string(item.Value) != "1" {
		t.Error("Expected item.Value = 1, actual: %s", string(item.Value))
	}
}

func Test_Length(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

	values := make([]string, 10)
	for i := 0; i < 10; i++ {
		values[i] = strconv.Itoa(i)
	}

	for i := 0; i < len(values); i++ {
		if q.Length() != uint64(i) {
			t.Error("Length %d, expected %d", q.Length(), i)
		}
		q.Enqueue([]byte(values[i]))
	}

	for i := len(values); i < 0; i-- {
		_, err := q.Dequeue()
		if err != nil {
			t.Error(err)
		}
		if q.Length() != uint64(i) {
			t.Error("Length %d, expected %d", q.Length(), i)
		}
	}
}

func Test_initialize(t *testing.T) {
	q, _ := Open(name, dir)
	defer q.Drop()

	q.Enqueue([]byte("1"))
	q.Enqueue([]byte("2"))
	q.Enqueue([]byte("3"))

	if q.Length() != 3 {
		t.Error("Invalid lendth")
	}

	expectedLength := q.Length()
	expectedHead := q.Head()
	expectedTail := q.Tail()

	// Reopen queue and test if it initializes itself properly
	q.Close()
	q, err = Open(name, dir)
	if err != nil {
		t.Error("Queue initialization error: ", err)
	}

	actualLength := q.Length()
	if actualLength != 3 {
		t.Errorf("Length expected: %d, actual: %d", expectedLength, actualLength)
	}

	if q.Head() != expectedHead {
		t.Errorf("Head expected: %d, actual: %d", expectedHead, q.Head)
	}

	if q.Tail() != expectedTail {
		t.Errorf("Tail expected: %d, actual: %d", expectedTail, q.Tail)
	}
}

func Example_queuePath() {
	q, _ := Open("test_queue", dir)
	defer q.Drop()
	fmt.Println(q.Path())
	// Output: ./test_data/test_queue
}
