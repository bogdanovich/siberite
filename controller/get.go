package controller

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync/atomic"
)

var timeoutRegexp = regexp.MustCompile(`(t\=\d+)\/?`)

// Get handles GET command
// Command: GET <queue>
// Response:
// VALUE <queue> 0 <bytes>
// <data block>
// END
func (c *Controller) Get(input []string) error {
	var err error
	cmd := parseGetCommand(input)

	switch cmd.SubCommand {
	case "", "open":
		err = c.get(cmd)
	case "close":
		err = c.getClose(cmd)
	case "close/open":
		if err = c.getClose(cmd); err == nil {
			err = c.get(cmd)
		}
	case "abort":
		err = c.abort()
	case "peek":
		err = c.peek(cmd)
	default:
		err = ErrInvalidCommand
	}

	if err != nil {
		return err
	}
	fmt.Fprint(c.rw.Writer, endMessage)
	c.rw.Writer.Flush()
	return nil
}

func (c *Controller) get(cmd *Command) error {
	if c.currentValue != nil {
		return ErrCloseCurrentItemFirst
	}

	q, err := c.getConsumer(cmd)
	if err != nil {
		log.Println(cmd, err)
		return NewError(commonError, err)
	}
	value, _ := q.GetNext()
	if len(value) > 0 {
		fmt.Fprintf(c.rw.Writer, "VALUE %s 0 %d\r\n", cmd.QueueName, len(value))
		fmt.Fprintf(c.rw.Writer, "%s\r\n", value)
	}
	if strings.Contains(cmd.SubCommand, "open") && len(value) > 0 {
		c.setCurrentState(cmd, value)
		q.Stats().UpdateOpenReads(1)
	}
	atomic.AddUint64(&c.repo.Stats.CmdGet, 1)
	return nil
}

func (c *Controller) getClose(cmd *Command) error {
	q, err := c.getConsumer(cmd)
	if err != nil {
		log.Println(cmd, err)
		return NewError(commonError, err)
	}
	if c.currentValue != nil {
		q.Stats().UpdateOpenReads(-1)
		c.setCurrentState(nil, nil)
	}

	return nil
}

func (c *Controller) abort() error {
	if c.currentValue != nil {
		q, err := c.getConsumer(c.currentCommand)
		if err != nil {
			log.Println(c.currentCommand, err)
			return NewError(commonError, err)
		}
		err = q.PutBack(c.currentValue)
		if err != nil {
			return NewError(commonError, err)
		}
		if c.currentValue != nil {
			q.Stats().UpdateOpenReads(-1)
			c.setCurrentState(nil, nil)
		}
	}
	return nil
}

func (c *Controller) peek(cmd *Command) error {
	q, err := c.getConsumer(cmd)
	if err != nil {
		log.Println(cmd, err)
		return NewError(commonError, err)
	}
	value, _ := q.Peek()
	if len(value) > 0 {
		fmt.Fprintf(c.rw.Writer, "VALUE %s 0 %d\r\n", cmd.QueueName, len(value))
		fmt.Fprintf(c.rw.Writer, "%s\r\n", value)
	}
	atomic.AddUint64(&c.repo.Stats.CmdGet, 1)
	return nil
}

func parseGetCommand(input []string) *Command {
	cmd := &Command{Name: input[0], QueueName: input[1], SubCommand: ""}
	if strings.Contains(input[1], "t=") {
		input[1] = timeoutRegexp.ReplaceAllString(input[1], "")
	}
	tokens := make([]string, 3)
	if strings.Contains(input[1], "/") {
		tokens = strings.SplitN(input[1], "/", 2)
		cmd.QueueName = tokens[0]
		cmd.SubCommand = strings.Trim(tokens[1], "/")
	}
	if strings.Contains(cmd.QueueName, cgSeparator) {
		tokens = strings.SplitN(cmd.QueueName, cgSeparator, 3)
		cmd.QueueName = tokens[0]
		cmd.ConsumerGroup = tokens[1]
	}
	return cmd
}
