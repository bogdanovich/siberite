package controller

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bogdanovich/siberite/queue"
	"github.com/bogdanovich/siberite/repository"
)

const (
	commonError   = "ERROR"
	clientError   = "CLIENT_ERROR"
	storedMessage = "STORED\r\n"
	endMessage    = "END\r\n"

	// consumer group separator
	cgSeparator = "."
)

var (
	// ErrUnknownCommand is returned when command was not recognized
	ErrUnknownCommand = &Error{commonError, "Unknown command"}

	// ErrInvalidCommand means command wasn't parsed correcty
	ErrInvalidCommand = &Error{clientError, "Invalid command"}

	// ErrCloseCurrentItemFirst is returned when client attemted
	// to read next item before closing the current one
	ErrCloseCurrentItemFirst = &Error{clientError, "Close current item first"}

	// ErrBadDataChunk is returned when data provided by client has different size
	ErrBadDataChunk = &Error{clientError, "bad data chunk"}

	// ErrInvalidDataSize is returned when data size field is not a number
	ErrInvalidDataSize = &Error{clientError, "Invalid <bytes> number"}

	// ErrClientQuit is returned when client sends 'quit' command (not an error)
	ErrClientQuit = &Error{commonError, "Quit command received"}
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
	dataBuffer     []byte
	currentValue   []byte
	currentCommand *Command
}

// Command represents a client command
type Command struct {
	Name          string
	QueueName     string
	SubCommand    string
	ConsumerGroup string
	FanoutQueues  []string
	DataSize      int
}

// NewSession creates and initializes new controller
func NewSession(conn Conn, repo *repository.QueueRepository) *Controller {
	atomic.AddUint64(&repo.Stats.TotalConnections, 1)
	atomic.AddUint64(&repo.Stats.CurrentConnections, 1)
	return &Controller{
		conn:       conn,
		rw:         bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		repo:       repo,
		dataBuffer: make([]byte, 1024),
	}
}

// FinishSession aborts unfinished transaction
func (c *Controller) FinishSession() {
	if c.currentValue != nil {
		c.abort()
	}
	atomic.AddUint64(&c.repo.Stats.CurrentConnections, ^uint64(0))
}

// ReadFirstMessage reads initial message from connection buffer
func (c *Controller) ReadFirstMessage() (string, error) {
	return c.rw.Reader.ReadString('\n')
}

// UnknownCommand reports an error
func (c *Controller) UnknownCommand() error {
	c.SendError(ErrUnknownCommand)
	return ErrUnknownCommand
}

//SendError sends an error message to the client
func (c *Controller) SendError(err error) {

	if e, ok := err.(*Error); ok {
		fmt.Fprintf(c.rw.Writer, "%s\r\n", e.Error())
	} else {
		fmt.Fprintf(c.rw.Writer, "%s %s\r\n", commonError, err.Error())
	}

	c.rw.Writer.Flush()
}

// Save current unconfirmed item
func (c *Controller) setCurrentState(cmd *Command, currentValue []byte) {
	c.currentCommand = cmd
	c.currentValue = currentValue
}

func (c *Controller) getConsumer(cmd *Command) (queue.Consumer, error) {
	if cmd.ConsumerGroup == "" {
		return c.repo.GetQueue(cmd.QueueName)
	}

	q, err := c.repo.GetQueue(cmd.QueueName)
	if err == nil {
		return q.ConsumerGroup(cmd.ConsumerGroup)
	}
	return q, err
}

func parseCommand(input []string) *Command {
	cmd := &Command{Name: input[0], QueueName: input[1], SubCommand: ""}
	tokens := make([]string, 3)
	if strings.Contains(cmd.QueueName, cgSeparator) {
		tokens = strings.SplitN(cmd.QueueName, cgSeparator, 3)
		cmd.QueueName = tokens[0]
		cmd.ConsumerGroup = tokens[1]
	}
	return cmd
}
