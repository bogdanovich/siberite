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

// OpenReadsValue returns the current number of open reads.
func (s *Stats) OpenReadsValue() int64 {
	return atomic.LoadInt64(&s.OpenReads)
}
