package controller

import (
	"errors"
	"fmt"
	"log"
)

// Command: FLUSH_ALL
// Response:
// END
func (self *Controller) FlushAll() error {
	err := self.repo.FlushAllQueues()
	if err != nil {
		log.Printf("Can't flush all queues: %s", err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	fmt.Fprint(self.rw.Writer, "Flushed all queues.\r\n")
	self.rw.Writer.Flush()
	return nil
}
