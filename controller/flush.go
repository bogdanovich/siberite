package controller

import (
	"errors"
	"fmt"
	"log"
)

// Flush handles FLUSH command
// Command: FLUSH <queue>
// Response:
// END
func (c *Controller) Flush(input []string) error {
	cmd := &Command{Name: input[0], QueueName: input[1]}
	err := c.repo.FlushQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't flush queue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	fmt.Fprint(c.rw.Writer, "END\r\n")
	c.rw.Writer.Flush()
	return nil
}
