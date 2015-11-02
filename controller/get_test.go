package controller

import (
	"testing"

	"github.com/bogdanovich/siberite/repository"
	"github.com/stretchr/testify/assert"
)

func Test_parseGetCommand(t *testing.T) {
	testCases := map[string]Command{
		"work":                            Command{},
		"work/open":                       Command{SubCommand: "open"},
		"work/close":                      Command{SubCommand: "close"},
		"work/abort":                      Command{SubCommand: "abort"},
		"work/peek":                       Command{SubCommand: "peek"},
		"work/t=10":                       Command{},
		"work/t=10/t=100/t=22":            Command{},
		"work/t=10/open":                  Command{SubCommand: "open"},
		"work/open/t=10":                  Command{SubCommand: "open"},
		"work/close/open/t=10":            Command{SubCommand: "close/open"},
		"work/close/t=10/open/abort":      Command{SubCommand: "close/open/abort"},
		"work:cg":                         Command{ConsumerGroup: "cg"},
		"work:consumer/open":              Command{SubCommand: "open", ConsumerGroup: "consumer"},
		"work:1/close":                    Command{SubCommand: "close", ConsumerGroup: "1"},
		"work:000/abort":                  Command{SubCommand: "abort", ConsumerGroup: "000"},
		"work:1:2/peek":                   Command{SubCommand: "peek", ConsumerGroup: "1"},
		"work:consumergroup/t=10":         Command{ConsumerGroup: "consumergroup"},
		"work:test:cg/t=10/t=100/t=22":    Command{ConsumerGroup: "test"},
		"work:1cg/t=10/open":              Command{SubCommand: "open", ConsumerGroup: "1cg"},
		"work:c/open/t=10":                Command{SubCommand: "open", ConsumerGroup: "c"},
		"work:0/close/open/t=10":          Command{SubCommand: "close/open", ConsumerGroup: "0"},
		"work:group/close/t=1/open/abort": Command{SubCommand: "close/open/abort", ConsumerGroup: "group"},
	}

	for input, command := range testCases {
		cmd := parseGetCommand([]string{"get", input})
		assert.Equal(t, "get", cmd.Name, input)
		assert.Equal(t, "work", cmd.QueueName, input)
		assert.Equal(t, command.SubCommand, cmd.SubCommand, input)
		assert.Equal(t, command.ConsumerGroup, cmd.ConsumerGroup, input)
	}
}

// Initialize queue 'test' with 1 item
// get test = value
// get test = empty
// get test/close = empty
// get test/abort = empty
func Test_Get(t *testing.T) {
	repo, err := repository.NewRepository(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)

	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("0123456789"))

	// When queue has items
	// get test = value
	command := []string{"get", "test"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 10\r\n0123456789\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// When queue is empty
	// get test = empty
	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// When queue is empty
	// get test/close = empty
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// When queue is empty
	// get test/abort = empty
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())
}

// Initialize test queue with 4 items
// get test/open = value
// get test = error
// get test/close = empty
// get test/open = value
// get test/open = error
// get test/abort = empty
// get test/open = value
// get test/peek = next value
// get test/close = empty
func Test_GetOpen(t *testing.T) {
	repo, err := repository.NewRepository(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)

	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("1"))
	q.Enqueue([]byte("2"))
	q.Enqueue([]byte("3"))
	q.Enqueue([]byte("4"))

	// get test/open = value
	command := []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test = error
	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.Equal(t, "CLIENT_ERROR Close current item first", err.Error())

	mockTCPConn.WriteBuffer.Reset()

	// get test/close = value
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test/open = value
	command = []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n2\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	// get test/open = error
	command = []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Equal(t, err.Error(), "CLIENT_ERROR Close current item first")

	mockTCPConn.WriteBuffer.Reset()

	// get test/abort = value
	command = []string{"get", "test/abort"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test/open = value
	command = []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n2\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test/peek = value
	command = []string{"get", "test/peek"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n3\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test/close = value
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())
}

// Initialize test queue with 2 items
// get test/open = value
// FinishSession (disconnect)
// NewSession
// get test = same value
func Test_GetOpen_Disconnect(t *testing.T) {
	repo, err := repository.NewRepository(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)

	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	repo.FlushQueue("test")
	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("1"))
	q.Enqueue([]byte("2"))

	// get test/open = value
	command := []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	controller.FinishSession()

	mockTCPConn = NewMockTCPConn()
	controller = NewSession(mockTCPConn, repo)

	// get test = same value
	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())
}

// Initialize test queue with 4 items
// get test/close/open = value
// get test = error
// get test/t=10/close/open = value
// get test/close/open/t=1000 = next value
// FinishSession (disconnect)
// get test/close/t=88/open = same value
func Test_GetCloseOpen(t *testing.T) {
	repo, err := repository.NewRepository(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)

	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	repo.FlushQueue("test")
	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("1"))
	q.Enqueue([]byte("2"))
	q.Enqueue([]byte("3"))
	q.Enqueue([]byte("4"))

	// get test/close/open = 1
	command := []string{"get", "test/close/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test = error
	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.Equal(t, err.Error(), "CLIENT_ERROR Close current item first")

	mockTCPConn.WriteBuffer.Reset()

	// get test/abort = value
	command = []string{"get", "test/abort"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test/t=10/close/open = value
	command = []string{"get", "test/t=10/close/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// get test/close/open/t=1000 = next value
	command = []string{"get", "test/close/open/t=1000"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n2\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	controller.FinishSession()

	mockTCPConn = NewMockTCPConn()
	controller = NewSession(mockTCPConn, repo)

	// get test = same value
	command = []string{"get", "test/t=88/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n2\r\nEND\r\n", mockTCPConn.WriteBuffer.String())
}

// Initialize test queue with 2 items
// gets test/open = value
// gets test = error
// GETS test/t=10/close/open = value
func Test_Gets(t *testing.T) {
	repo, err := repository.NewRepository(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)

	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	repo.FlushQueue("test")
	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("1"))
	q.Enqueue([]byte("2"))

	// gets test/open = 1
	command := []string{"gets", "test"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// GETS test/t=10/close/open = 2
	command = []string{"GETS", "test/t=10/close/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n2\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

}
