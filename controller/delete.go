package controller

import (
	"errors"
	"fmt"
	"log"
)

// Command: delete <queue>
// Response:
// END
func (self *Controller) Delete(input []string) error {
	cmd := &Command{Name: input[0], QueueName: input[1]}
	err := self.repo.DeleteQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't delete queue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	fmt.Fprint(self.rw.Writer, "END\r\n")
	self.rw.Writer.Flush()
	return nil
}
