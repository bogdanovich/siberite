package controller

import "fmt"

// Stats handles STATS command
func (c *Controller) Stats() error {
	for _, item := range c.repo.FullStats() {
		fmt.Fprintf(c.rw.Writer, "STAT %s %s\r\n", item.Key, item.Value)
	}
	fmt.Fprintf(c.rw.Writer, endMessage)
	c.rw.Writer.Flush()
	return nil
}
