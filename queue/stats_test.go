package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UpdateOpenReads(t *testing.T) {
	stats := &Stats{}
	stats.UpdateOpenReads(1)
	assert.EqualValues(t, 1, stats.OpenReads)
	stats.UpdateOpenReads(-1)
	assert.EqualValues(t, 0, stats.OpenReads)
}
