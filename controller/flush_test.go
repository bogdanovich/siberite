package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Flush(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 1)
	defer cleanupControllerTest(repo)

	command := []string{"flush", "test"}
	err = controller.Flush(command)
	assert.Nil(t, err)
	response, err := mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", response)

	command = []string{"FLUSH", "test"}
	err = controller.Flush(command)
	assert.Nil(t, err)
	response, err = mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", response)
}
