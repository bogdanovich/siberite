package controller

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/bogdanovich/siberite/repository"
)

// Conn represents a connection interface
type Conn interface {
	io.Reader
	io.Writer
	SetDeadline(t time.Time) error
}

// Controller represents a connection controller
type Controller struct {
	conn           Conn
	rw             *bufio.ReadWriter
	repo           *repository.QueueRepository
	currentValue   []byte
	currentCommand *Command
}

// Command represents a client command
type Command struct {
	Name       string
	QueueName  string
	SubCommand string
	DataSize   int
}

// NewSession creates and initializes new controller
func NewSession(conn Conn, repo *repository.QueueRepository) *Controller {
	atomic.AddUint64(&repo.Stats.TotalConnections, 1)
	atomic.AddUint64(&repo.Stats.CurrentConnections, 1)
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return &Controller{conn, rw, repo, nil, nil}
}

// FinishSession aborts unfinished transaction
func (c *Controller) FinishSession() {
	if c.currentValue != nil {
		c.abort(c.currentCommand)
	}
	atomic.AddUint64(&c.repo.Stats.CurrentConnections, ^uint64(0))
}

// ReadFirstMessage reads initial message from connection buffer
func (c *Controller) ReadFirstMessage() (string, error) {
	return c.rw.Reader.ReadString('\n')
}

// UnknownCommand reports an error
func (c *Controller) UnknownCommand() error {
	errorMessage := "ERROR Unknown command"
	c.SendError(errorMessage)
	return errors.New(errorMessage)
}

//SendError sends an error message to the client
func (c *Controller) SendError(errorMessage string) {
	fmt.Fprintf(c.rw.Writer, "%s\r\n", errorMessage)
	c.rw.Writer.Flush()
}

// Save current unconfirmed item
func (c *Controller) setCurrentState(cmd *Command, currentValue []byte) {
	c.currentCommand = cmd
	c.currentValue = currentValue
}
