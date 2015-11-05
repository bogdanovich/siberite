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
	cmd := parseCommand(input)

	var err error
	if cmd.ConsumerGroup != "" {
		q, err := c.repo.GetQueue(cmd.QueueName)
		if err != nil {
			return err
		}

		err = q.DeleteConsumerGroup(cmd.ConsumerGroup)
	} else {
		err = c.repo.DeleteQueue(cmd.QueueName)
	}

	if err != nil {
		log.Printf("Can't delete %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	fmt.Fprint(c.rw.Writer, "END\r\n")
	c.rw.Writer.Flush()
	return nil
}
