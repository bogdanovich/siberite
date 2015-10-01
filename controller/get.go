package controller

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync/atomic"
)

var timeoutRegexp = regexp.MustCompile(`(t\=\d+)\/?`)

// Command: GET <queue>
// Response:
// VALUE <queue> 0 <bytes>
// <data block>
// END
func (self *Controller) Get(input []string) error {
	cmd := parseGetCommand(input)

	switch cmd.SubCommand {
	case "", "open":
		return self.get(cmd)
	case "close":
		return self.getClose(cmd)
	case "abort":
		return self.getAbort(cmd)
	case "peek":
		return self.peek(cmd)
	}

	return errors.New("ERROR " + "Invalid command")
}

func (self *Controller) get(cmd *Command) error {
	if self.currentItem != nil {
		return errors.New("CLIENT_ERROR " + "Close current item first")
	}

	q, err := self.repo.GetQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't GetQueue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	item, _ := q.Dequeue()
	if len(item.Value) > 0 {
		fmt.Fprintf(self.rw.Writer, "VALUE %s 0 %d\r\n", cmd.QueueName, len(item.Value))
		fmt.Fprintf(self.rw.Writer, "%s\r\n", item.Value)
	}
	fmt.Fprint(self.rw.Writer, "END\r\n")
	if cmd.SubCommand == "open" && len(item.Value) > 0 {
		self.setCurrentState(cmd, item)
		q.AddOpenTransactions(1)
	}
	self.rw.Writer.Flush()
	atomic.AddUint64(&self.repo.Stats.CmdGet, 1)
	return nil
}

func (self *Controller) getClose(cmd *Command) error {
	q, err := self.repo.GetQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't GetQueue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	if self.currentItem != nil {
		q.AddOpenTransactions(-1)
		self.setCurrentState(nil, nil)
	}

	fmt.Fprint(self.rw.Writer, "END\r\n")
	self.rw.Writer.Flush()
	return nil
}

func (self *Controller) getAbort(cmd *Command) error {
	self.abort(cmd)
	fmt.Fprint(self.rw.Writer, "END\r\n")
	self.rw.Writer.Flush()
	return nil
}

func (self *Controller) abort(cmd *Command) error {
	if self.currentItem != nil {
		q, err := self.repo.GetQueue(cmd.QueueName)
		if err != nil {
			log.Printf("Can't GetQueue %s: %s", cmd.QueueName, err.Error())
			return errors.New("SERVER_ERROR " + err.Error())
		}
		err = q.Prepend(self.currentItem)
		if err != nil {
			return errors.New("SERVER_ERROR " + err.Error())
		}
		if self.currentItem != nil {
			q.AddOpenTransactions(-1)
			self.setCurrentState(nil, nil)
		}
	}
	return nil
}

func (self *Controller) peek(cmd *Command) error {
	q, err := self.repo.GetQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't GetQueue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}
	item, _ := q.Peek()
	if len(item.Value) > 0 {
		fmt.Fprintf(self.rw.Writer, "VALUE %s 0 %d\r\n", cmd.QueueName, len(item.Value))
		fmt.Fprintf(self.rw.Writer, "%s\r\n", item.Value)
	}
	fmt.Fprint(self.rw.Writer, "END\r\n")
	self.rw.Writer.Flush()
	atomic.AddUint64(&self.repo.Stats.CmdGet, 1)
	return nil
}

func parseGetCommand(input []string) *Command {
	cmd := &Command{Name: input[0], QueueName: input[1], SubCommand: ""}
	if strings.Contains(input[1], "t=") {
		input[1] = timeoutRegexp.ReplaceAllString(input[1], "")
	}
	if strings.Contains(input[1], "/") {
		tokens := strings.SplitN(input[1], "/", 2)
		cmd.QueueName = tokens[0]
		cmd.SubCommand = strings.Trim(tokens[1], "/")
	}
	return cmd
}
