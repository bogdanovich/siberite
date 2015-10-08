package controller

import "fmt"

// Version handles VERSION command
func (c *Controller) Version() error {
	fmt.Fprintf(c.rw.Writer, "VERSION "+c.repo.Stats.Version+"\r\n")
	c.rw.Writer.Flush()
	return nil
}
