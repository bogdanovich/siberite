package controller

import (
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
			return NewError(commonError, err)
		}

		err = q.DeleteConsumerGroup(cmd.ConsumerGroup)
	} else {
		err = c.repo.DeleteQueue(cmd.QueueName)
	}

	if err != nil {
		log.Printf("Command %s: %s ", cmd, err.Error())
		return NewError(commonError, err)
	}
	fmt.Fprint(c.rw.Writer, endMessage)
	return c.rw.Writer.Flush()
}
