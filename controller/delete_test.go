package controller

import (
	"testing"

	"github.com/bogdanovich/siberite/repository"
	"github.com/stretchr/testify/assert"
)

func Test_Delete(t *testing.T) {
	repo, _ := repository.Initialize(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)
	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("1"))

	command := []string{"delete", "test"}
	err = controller.Delete(command)
	assert.Nil(t, err)
	response, err := mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", response)

	command = []string{"DELETE", "test"}
	err = controller.Delete(command)
	assert.Nil(t, err)
	response, err = mockTCPConn.WriteBuffer.ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", response)
}
