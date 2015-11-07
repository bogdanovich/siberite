package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_FlushAll(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 2)
	defer cleanupControllerTest(repo)

	err = controller.FlushAll()
	assert.Nil(t, err)
	response, err := mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "Flushed all queues.\r\n", response)
}
