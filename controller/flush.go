package controller

import (
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
		log.Printf(err.Error())
		return NewError("ERROR", err)
	}

	err = q.Flush()
	if err != nil {
		return NewError("ERROR", err)
	}

	fmt.Fprint(c.rw.Writer, "END\r\n")
	c.rw.Writer.Flush()
	return nil
}
