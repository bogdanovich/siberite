package controller

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bogdanovich/siberite/repository"
)

var dir = "./test_data"
var name = "test"
var err error

type mockTCPConn struct {
	WriteBuffer bytes.Buffer
	ReadBuffer  bytes.Buffer
}

func newMockTCPConn() *mockTCPConn {
	conn := &mockTCPConn{}
	return conn
}

func (conn *mockTCPConn) Read(b []byte) (int, error) {
	return conn.ReadBuffer.Read(b)
}

func (conn *mockTCPConn) Write(b []byte) (int, error) {
	return conn.WriteBuffer.Write(b)
}

func (conn *mockTCPConn) SetDeadline(t time.Time) error {
	return nil
}

func setupControllerTest(t *testing.T, qSize int) (*repository.QueueRepository,
	*Controller, *mockTCPConn) {

	repo, err := repository.NewRepository(dir)
	assert.NoError(t, err)

	mockTCPConn := newMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	q, err := repo.GetQueue("test")
	assert.NoError(t, err)

	for i := 0; i < qSize; i++ {
		q.Enqueue([]byte(strconv.Itoa(i)))
	}

	return repo, controller, mockTCPConn
}

func cleanupControllerTest(repo *repository.QueueRepository) {
	repo.DeleteAllQueues()
}

func TestMain(m *testing.M) {
	_ = os.RemoveAll(dir)
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Println(err)
	}
	result := m.Run()
	err = os.RemoveAll(dir)
	os.Exit(result)
}

func Test_Controller_NewSession_FinishSession(t *testing.T) {
	repo, c, _ := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	assert.Equal(t, uint64(1), repo.Stats.CurrentConnections)
	assert.Equal(t, uint64(1), repo.Stats.TotalConnections)

	c.FinishSession()
	assert.Equal(t, uint64(0), repo.Stats.CurrentConnections)
}

func Test_Controller_ReadFirstMessage(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	fmt.Fprintf(&mockTCPConn.ReadBuffer, "GET work\r\n")
	message, err := controller.ReadFirstMessage()
	assert.Nil(t, err)
	assert.Equal(t, "GET work\r\n", message)

	fmt.Fprintf(&mockTCPConn.ReadBuffer, "SET work 0 0 10\r\n0123456789\r\n")
	message, err = controller.ReadFirstMessage()
	assert.Nil(t, err)
	assert.Equal(t, "SET work 0 0 10\r\n", message)
}

func Test_Controller_UnknownCommand(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	err = controller.UnknownCommand()
	assert.EqualError(t, err, "ERROR Unknown command")
	assert.Equal(t, "ERROR Unknown command\r\n", mockTCPConn.WriteBuffer.String())

}

func Test_Controller_SendError(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	controller.SendError(errors.New("Test error message"))
	assert.Equal(t, "ERROR Test error message\r\n", mockTCPConn.WriteBuffer.String())
}

func Test_Controller_parseCommand(t *testing.T) {
	testCases := map[string]Command{
		"work":      Command{},
		"work.cg":   Command{ConsumerGroup: "cg"},
		"work.cg.1": Command{ConsumerGroup: "cg"},
	}

	for input, command := range testCases {
		cmd := parseGetCommand([]string{"get", input})
		assert.Equal(t, "get", cmd.Name, input)
		assert.Equal(t, "work", cmd.QueueName, input)
		assert.Equal(t, command.SubCommand, cmd.SubCommand, input)
		assert.Equal(t, command.ConsumerGroup, cmd.ConsumerGroup, input)
	}
}
