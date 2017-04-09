package controller

import (
	"fmt"
	"log"
)

// FlushAll handles FLUSH_ALL command.
// Command: FLUSH_ALL
// Response: Flushed all queues
func (c *Controller) FlushAll() error {
	if err := c.repo.FlushAllQueues(); err != nil {
		log.Printf("Can't flush all queues: %s", err.Error())
		return NewError(commonError, err)
	}
	fmt.Fprint(c.rw.Writer, "Flushed all queues.\r\n")
	return c.rw.Writer.Flush()
}
