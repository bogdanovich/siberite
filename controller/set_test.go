package controller

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Set(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	command := []string{"set", "test", "0", "0", "10"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0123567890\r\n")

	err = controller.Set(command)
	assert.NoError(t, err)
	assert.Equal(t, "STORED\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"set", "test", "0", "1"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0\r\n")

	err = controller.Set(command)
	assert.EqualError(t, err, "CLIENT_ERROR Invalid command")

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"set", "test", "0", "0", "invalid"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0123567890\r\n")

	err = controller.Set(command)
	assert.EqualError(t, err, "CLIENT_ERROR Invalid <bytes> number")

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"set", "test", "0", "0", "10"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "01235678901234567890\r\n")

	err = controller.Set(command)
	assert.Equal(t, "CLIENT_ERROR bad data chunk", err.Error())
}

func Test_Controller_SetFanout(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	queueNames := []string{"test", "fanout_test", "1", "2"}

	command := []string{"set", strings.Join(queueNames, "+"), "0", "0", "10"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0123456789\r\n")

	err = controller.Set(command)
	assert.NoError(t, err)
	assert.Equal(t, "STORED\r\n", mockTCPConn.WriteBuffer.String())

	for _, queueName := range queueNames {
		q, err := repo.GetQueue(queueName)
		assert.NoError(t, err)

		assert.EqualValues(t, 1, q.Length())

		value, err := q.GetNext()
		assert.NoError(t, err)
		assert.Equal(t, "0123456789", string(value))
		assert.True(t, q.IsEmpty())
	}
}
