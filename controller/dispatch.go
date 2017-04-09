package controller

import (
	"strings"
	"time"
)

// Dispatch routes client commands to their respective handlers
func (c *Controller) Dispatch() error {
	c.conn.SetDeadline(time.Now().Add(3e9))
	message, err := c.ReadFirstMessage()
	if err != nil {
		return err
	}

	c.conn.SetDeadline(time.Time{})
	command := strings.Split(strings.Trim(message, " \r\n"), " ")
	command[0] = strings.ToLower(command[0])

	switch command[0] {
	case "get", "gets":
		err = c.Get(command)
	case "set":
		err = c.Set(command)
	case "version":
		err = c.Version()
	case "stats":
		err = c.Stats()
	case "delete":
		err = c.Delete(command)
	case "flush":
		err = c.Flush(command)
	case "flush_all":
		err = c.FlushAll()
	case "quit":
		return ErrClientQuit
	default:
		return c.UnknownCommand()
	}

	if err != nil {
		c.SendError(err)
		return err
	}
	return nil
}
