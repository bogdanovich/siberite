package controller

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync/atomic"
)

// Set handles SET command
// Command: SET <queue> <not_impl> <not_impl> <bytes>
// <data block>
// Response: STORED
func (c *Controller) Set(input []string) error {
	if len(input) < 5 || len(input) > 6 {
		return &Error{"CLIENT_ERROR", "Invalid input"}
	}

	totalBytes, err := strconv.Atoi(input[4])
	if err != nil {
		return &Error{"CLIENT_ERROR", "Invalid <bytes> number"}
	}

	cmd := &Command{Name: input[0], QueueName: input[1], DataSize: totalBytes}

	dataBlock, err := c.readDataBlock(cmd.DataSize)
	if err != nil {
		return NewError("CLIENT_ERROR", err)
	}

	q, err := c.repo.GetQueue(cmd.QueueName)
	if err != nil {
		log.Println(cmd, err)
		return NewError("ERROR", err)
	}

	err = q.Enqueue([]byte(dataBlock))
	if err != nil {
		return NewError("ERROR", err)
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
