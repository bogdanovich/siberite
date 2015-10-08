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
		return errors.New("ERROR Invalid input")
	}

	totalBytes, err := strconv.Atoi(input[4])
	if err != nil {
		return errors.New("ERROR Invalid <bytes> number")
	}

	cmd := &Command{Name: input[0], QueueName: input[1], DataSize: totalBytes}

	dataBlock, err := c.readDataBlock(cmd.DataSize)
	if err != nil {
		return errors.New("CLIENT_ERROR " + err.Error())
	}

	q, err := c.repo.GetQueue(cmd.QueueName)
	if err != nil {
		log.Printf("Can't GetQueue %s: %s", cmd.QueueName, err.Error())
		return errors.New("SERVER_ERROR " + err.Error())
	}

	err = q.Enqueue([]byte(dataBlock))
	if err != nil {
		return errors.New("SERVER_ERROR " + err.Error())
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
