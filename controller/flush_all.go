package controller

import (
	"errors"
	"fmt"
	"log"
)

// FlushAll handles FLUSH_ALL command.
// Command: FLUSH_ALL
// Response: Flushed all queues
func (c *Controller) FlushAll() error {
	err := c.repo.FlushAllQueues()
	if err != nil {
		log.Printf("Can't flush all queues: %s", err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	fmt.Fprint(c.rw.Writer, "Flushed all queues.\r\n")
	c.rw.Writer.Flush()
	return nil
}
