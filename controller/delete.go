package controller

import (
	"errors"
	"fmt"
	"log"
)

// Delete handles DELETE command
// Command: DELETE <queue>
// Response:
// END
func (c *Controller) Delete(input []string) error {
	cmd := &Command{Name: input[0], QueueName: input[1]}
	err := c.repo.DeleteQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't delete queue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	fmt.Fprint(c.rw.Writer, "END\r\n")
	c.rw.Writer.Flush()
	return nil
}
