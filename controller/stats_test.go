package controller

import (
	"strconv"
	"strings"
	"testing"

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
	assert.Nil(t, err)

	lines := strings.Split(strings.TrimSuffix(mockTCPConn.WriteBuffer.String(), "\r\n"), "\r\n")
	assert.Len(t, lines, 12)

	uptimeFields := strings.Fields(lines[0])
	assert.Equal(t, []string{"STAT", "uptime"}, uptimeFields[:2])
	_, err = strconv.ParseUint(uptimeFields[2], 10, 64)
	assert.NoError(t, err)

	timeFields := strings.Fields(lines[1])
	assert.Equal(t, []string{"STAT", "time"}, timeFields[:2])
	_, err = strconv.ParseUint(timeFields[2], 10, 64)
	assert.NoError(t, err)

	assert.Equal(t, "STAT version "+repo.Stats.Version, lines[2])
	assert.Equal(t, "STAT curr_connections 1", lines[3])
	assert.Equal(t, "STAT total_connections 1", lines[4])
	assert.Equal(t, "STAT cmd_get 0", lines[5])
	assert.Equal(t, "STAT cmd_set 0", lines[6])
	assert.Equal(t, "STAT queue_test_items 3", lines[7])
	assert.Equal(t, "STAT queue_test_open_transactions 0", lines[8])
	assert.Equal(t, "STAT queue_test.cg1_items 2", lines[9])
	assert.Equal(t, "STAT queue_test.cg1_open_transactions 0", lines[10])
	assert.Equal(t, "END", lines[11])
}
