package controller

import (
	"testing"

	"github.com/bogdanovich/siberite/repository"
	"github.com/stretchr/testify/assert"
)

// Initialize queue 'test' with 1 item
// get test = value
// get test = empty
// get test/close = empty
// get test/abort = empty
func Test_Get(t *testing.T) {
	repo, err := repository.Initialize(dir)
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
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "VALUE test 0 10\r\n0123456789\r\nEND\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// When queue is empty
	// get test = empty
	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "END\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// When queue is empty
	// get test/close = empty
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "END\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// When queue is empty
	// get test/abort = empty
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "END\r\n")
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
	repo, err := repository.Initialize(dir)
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
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "VALUE test 0 1\r\n1\r\nEND\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// get test = error
	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.Equal(t, err.Error(), "CLIENT_ERROR Close current item first")

	mockTCPConn.WriteBuffer.Reset()

	// get test/close = value
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "END\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// get test/open = value
	command = []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "VALUE test 0 1\r\n2\r\nEND\r\n")

	// get test/open = error
	command = []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Equal(t, err.Error(), "CLIENT_ERROR Close current item first")

	mockTCPConn.WriteBuffer.Reset()

	// get test/abort = value
	command = []string{"get", "test/abort"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "END\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// get test/open = value
	command = []string{"get", "test/open"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "VALUE test 0 1\r\n2\r\nEND\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// get test/peek = value
	command = []string{"get", "test/peek"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "VALUE test 0 1\r\n3\r\nEND\r\n")

	mockTCPConn.WriteBuffer.Reset()

	// get test/close = value
	command = []string{"get", "test/close"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "END\r\n")
}

// Initialize test queue with 2 items
// get test/open = value
// FinishSession (disconnect)
// NewSession
// get test = same value
func Test_GetOpen_Disconnect(t *testing.T) {
	repo, err := repository.Initialize(dir)
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
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "VALUE test 0 1\r\n1\r\nEND\r\n")

	mockTCPConn.WriteBuffer.Reset()

	controller.FinishSession()

	mockTCPConn = NewMockTCPConn()
	controller = NewSession(mockTCPConn, repo)

	// get test = same value
	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "VALUE test 0 1\r\n1\r\nEND\r\n")
}
