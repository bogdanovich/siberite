package controller

import "fmt"

func (self *Controller) Stats() error {

	for _, item := range self.repo.FullStats() {
		fmt.Fprintf(self.rw.Writer, "STAT %s %s\r\n", item.Key, item.Value)
	}
	fmt.Fprintf(self.rw.Writer, "END\r\n")
	self.rw.Writer.Flush()
	return nil
}
