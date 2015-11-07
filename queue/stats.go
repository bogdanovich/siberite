package queue

import "sync/atomic"

// Stats contains queue level stats
type Stats struct {
	OpenReads int64
}

// UpdateOpenReads increments OpenReads stats item
func (s *Stats) UpdateOpenReads(value int64) {
	atomic.AddInt64(&s.OpenReads, value)
}
