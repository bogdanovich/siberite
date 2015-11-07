package controller

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Set(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	command := []string{"set", "test", "0", "0", "10"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0123567890\r\n")

	err = controller.Set(command)
	assert.Nil(t, err)
	assert.Equal(t, "STORED\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"set", "test", "0", "1"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0\r\n")

	err = controller.Set(command)
	assert.EqualError(t, err, "CLIENT_ERROR Invalid input")

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
