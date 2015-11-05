package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Version(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	err = controller.Version()
	assert.Nil(t, err)
	assert.Equal(t, "VERSION "+repo.Stats.Version+"\r\n", mockTCPConn.WriteBuffer.String())
}
