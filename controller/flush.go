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
	cmd := parseCommand(input)

	q, err := c.getConsumer(cmd)
	if err != nil {
		log.Printf("Can't get consumer %s: %s", cmd, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}

	err = q.Flush()
	if err != nil {
		log.Printf("Flush error %s: %s", cmd, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}

	fmt.Fprint(c.rw.Writer, "END\r\n")
	c.rw.Writer.Flush()
	return nil
}
