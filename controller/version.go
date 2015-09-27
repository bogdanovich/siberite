package controller

import "fmt"

// Command: VERSION
func (self *Controller) Version() error {
	fmt.Fprintf(self.rw.Writer, "VERSION "+self.repo.Stats.Version+"\r\n")
	self.rw.Writer.Flush()
	return nil
}
