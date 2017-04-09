package controller

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Stats(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 3)
	defer cleanupControllerTest(repo)

	q, err := repo.GetQueue("test")
	assert.NoError(t, err)

	cg, err := q.ConsumerGroup("cg1")
	assert.NoError(t, err)
	cg.GetNext()

	err = controller.Stats()
	statsResponse := "STAT uptime 0\r\n" +
		fmt.Sprintf("STAT time %d\r\n", time.Now().Unix()) +
		"STAT version " + repo.Stats.Version + "\r\n" +
		"STAT curr_connections 1\r\n" +
		"STAT total_connections 1\r\n" +
		"STAT cmd_get 0\r\n" +
		"STAT cmd_set 0\r\n" +
		fmt.Sprintf("STAT queue_test_items %d\r\n", 3) +
		"STAT queue_test_open_transactions 0\r\n" +
		fmt.Sprintf("STAT queue_test.cg1_items %d\r\n", 2) +
		"STAT queue_test.cg1_open_transactions 0\r\n" +
		"END\r\n"
	assert.Nil(t, err)
	assert.Equal(t, statsResponse, mockTCPConn.WriteBuffer.String())
}
