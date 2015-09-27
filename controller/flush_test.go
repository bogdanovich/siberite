package controller

import (
	"testing"

	"github.com/bogdanovich/siberite/repository"
	"github.com/stretchr/testify/assert"
)

func Test_Flush(t *testing.T) {
	repo, err := repository.Initialize(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)
	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("1"))

	command := []string{"flush", "test"}
	err = controller.Flush(command)
	assert.Nil(t, err)
	response, err := mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, response, "END\r\n")

	command = []string{"FLUSH", "test"}
	err = controller.Flush(command)
	assert.Nil(t, err)
	response, err = mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, response, "END\r\n")
}
