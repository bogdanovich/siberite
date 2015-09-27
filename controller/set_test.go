package controller

import (
	"fmt"
	"testing"

	"github.com/bogdanovich/siberite/repository"
	"github.com/stretchr/testify/assert"
)

func Test_Set(t *testing.T) {
	repo, err := repository.Initialize(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)

	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	command := []string{"set", "test", "0", "0", "10"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0123567890\r\n")

	err = controller.Set(command)
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), "STORED\r\n")

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"set", "test", "0", "1"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0\r\n")

	err = controller.Set(command)
	assert.Equal(t, err.Error(), "ERROR Invalid input")

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"set", "test", "0", "0", "invalid"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "0123567890\r\n")

	err = controller.Set(command)
	assert.Equal(t, err.Error(), "ERROR Invalid <bytes> number")

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"set", "test", "0", "0", "10"}
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "01235678901234567890\r\n")

	err = controller.Set(command)
	assert.Equal(t, err.Error(), "CLIENT_ERROR bad data chunk")
}
