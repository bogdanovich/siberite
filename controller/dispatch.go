package controller

import (
	"strings"
	"time"
)

func (self *Controller) Dispatch() error {
	var err error
	self.conn.SetDeadline(time.Now().Add(3e9))
	message, err := self.ReadFirstMessage()
	if err != nil {
		return err
	}

	self.conn.SetDeadline(time.Time{})
	command := strings.Split(strings.Trim(message, " \r\n"), " ")
	command[0] = strings.ToLower(command[0])

	switch command[0] {
	case "get":
		err = self.Get(command)
	case "set":
		err = self.Set(command)
	case "version":
		err = self.Version()
	case "stats":
		err = self.Stats()
	case "delete":
		err = self.Delete(command)
	case "flush":
		err = self.Flush(command)
	case "flush_all":
		err = self.FlushAll(command)
	default:
		return self.UnknownCommand()
	}

	if err != nil {
		self.SendError(err.Error())
		return err
	}
	return nil
}
