package controller

import (
	"fmt"
	"testing"
	"time"

	"github.com/bogdanovich/siberite/repository"
	"github.com/stretchr/testify/assert"
)

func Test_Stats(t *testing.T) {
	repo, err := repository.Initialize(dir)
	defer repo.CloseAllQueues()
	assert.Nil(t, err)

	mockTCPConn := NewMockTCPConn()
	controller := NewSession(mockTCPConn, repo)

	q, err := repo.GetQueue("test")
	assert.Nil(t, err)

	q.Enqueue([]byte("1"))

	err = controller.Stats()
	statsResponse := "STAT uptime 0\r\n" +
		fmt.Sprintf("STAT time %d\r\n", time.Now().Unix()) +
		"STAT version " + repo.Stats.Version + "\r\n" +
		"STAT curr_connections 1\r\n" +
		"STAT total_connections 1\r\n" +
		"STAT cmd_get 0\r\n" +
		"STAT cmd_set 0\r\n" +
		fmt.Sprintf("STAT queue_test_items %d\r\n", q.Length()) +
		"STAT queue_test_open_transactions 0\r\n" +
		"END\r\n"
	assert.Nil(t, err)
	assert.Equal(t, mockTCPConn.WriteBuffer.String(), statsResponse)
}
