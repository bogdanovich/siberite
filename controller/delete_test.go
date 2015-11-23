package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Delete(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 1)
	defer cleanupControllerTest(repo)

	// Create consumer group and get first item
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

	// Still returns first item (because it was deleted before)
	command = []string{"get", "test.cgroup"}
	err = controller.Get(command)
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"delete", "test"}
	err = controller.Delete(command)
	assert.NoError(t, err)
	response, err = mockTCPConn.WriteBuffer.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, "END\r\n", response)

	command = []string{"DELETE", "test"}
	err = controller.Delete(command)
	assert.NoError(t, err)
	response, err = mockTCPConn.WriteBuffer.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, "END\r\n", response)
}
