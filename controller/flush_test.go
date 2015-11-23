package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Flush(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 5)
	defer cleanupControllerTest(repo)

	// Returns first item
	command := []string{"get", "test.cgroup"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"flush", "test.cgroup"}
	err = controller.Flush(command)
	assert.Nil(t, err)
	response, err := mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", response)

	mockTCPConn.WriteBuffer.Reset()

	// Still returns first item (because it was flushed before)
	command = []string{"get", "test.cgroup"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"flush", "test"}
	err = controller.Flush(command)
	assert.Nil(t, err)
	response, err = mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", response)

	command = []string{"FLUSH", "test"}
	err = controller.Flush(command)
	assert.Nil(t, err)
	response, err = mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", response)
}
