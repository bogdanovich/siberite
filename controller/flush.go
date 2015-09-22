package controller

import (
	"errors"
	"fmt"
	"log"
)

// Command: flush <queue>
// Response:
// END
func (self *Controller) Flush(input []string) error {
	cmd := &Command{Name: input[0], QueueName: input[1]}
	err := self.repo.FlushQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't flush queue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	fmt.Fprint(self.rw.Writer, "END\r\n")
	self.rw.Writer.Flush()
	return nil
}
