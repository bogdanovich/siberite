package controller

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
)

// Set handles SET command
// Command: SET <queue> <not_impl> <not_impl> <bytes>
// <data block>
// Response: STORED
func (c *Controller) Set(input []string) error {
	cmd, err := parseSetCommand(input)
	if err != nil {
		return NewError("CLIENT_ERROR", err)
	}

	dataBlock, err := c.readDataBlock(cmd.DataSize)
	if err != nil {
		return NewError("CLIENT_ERROR", err)
	}

	if cmd.FanoutQueues == nil {
		err = c.storeDataBlock(cmd.QueueName, dataBlock)
		if err != nil {
			log.Println(cmd, err)
			return NewError("ERROR", err)
		}
	} else {
		for _, queueName := range cmd.FanoutQueues {
			err = c.storeDataBlock(queueName, dataBlock)
			if err != nil {
				log.Println(cmd, err)
				return NewError("ERROR", err)
			}
		}
	}

	fmt.Fprint(c.rw.Writer, "STORED\r\n")
	c.rw.Writer.Flush()
	atomic.AddUint64(&c.repo.Stats.CmdSet, 1)
	return nil
}

func (c *Controller) readDataBlock(totalBytes int) ([]byte, error) {
	expectedBytes := totalBytes + 2
	dataBlock := make([]byte, expectedBytes)
	_, err := io.ReadFull(c.rw.Reader, dataBlock)
	if err != nil {
		return nil, err
	}

	if string(dataBlock[totalBytes:]) != "\r\n" {
		return nil, errors.New("bad data chunk")
	}

	return dataBlock[:totalBytes], nil
}

func (c *Controller) storeDataBlock(queueName string, dataBlock []byte) error {
	q, err := c.repo.GetQueue(queueName)
	if err != nil {
		return err
	}
	return q.Enqueue([]byte(dataBlock))
}

func parseSetCommand(input []string) (*Command, error) {
	if len(input) < 5 || len(input) > 6 {
		return nil, errors.New("Invalid input")
	}

	totalBytes, err := strconv.Atoi(input[4])
	if err != nil {
		return nil, errors.New("Invalid <bytes> number")
	}

	cmd := &Command{Name: input[0], QueueName: input[1], DataSize: totalBytes}

	if strings.Contains(cmd.QueueName, "+") {
		cmd.FanoutQueues = strings.Split(cmd.QueueName, "+")
		cmd.QueueName = cmd.FanoutQueues[0]
	}
	return cmd, nil
}
