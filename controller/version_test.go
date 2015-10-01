package controller

import (
	"testing"

	"github.com/bogdanovich/siberite/repository"
	"github.com/stretchr/testify/assert"
)

func Test_Version(t *testing.T) {
	repo, err := repository.Initialize(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)
	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	err = controller.Version()
	assert.Nil(t, err)
	assert.Equal(t, "VERSION "+repo.Stats.Version+"\r\n", mockTCPConn.WriteBuffer.String())
}
