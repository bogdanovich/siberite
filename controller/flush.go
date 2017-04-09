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
		return NewError(commonError, err)
	}

	if err = q.Flush(); err != nil {
		return NewError(commonError, err)
	}

	fmt.Fprint(c.rw.Writer, endMessage)
	return c.rw.Writer.Flush()
}
