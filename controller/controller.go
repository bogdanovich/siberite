package controller

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/bogdanovich/siberite/queue"
	"github.com/bogdanovich/siberite/repository"
)

type Controller struct {
	conn           *net.TCPConn
	rw             *bufio.ReadWriter
	repo           *repository.QueueRepository
	currentItem    *queue.Item
	currentCommand *Command
}

type Command struct {
	Name       string
	QueueName  string
	SubCommand string
	DataSize   int
}

func NewSession(conn *net.TCPConn, repo *repository.QueueRepository) *Controller {
	atomic.AddUint64(&repo.Stats.TotalConnections, 1)
	atomic.AddUint64(&repo.Stats.CurrentConnections, 1)
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return &Controller{conn, rw, repo, nil, nil}
}

func (self *Controller) FinishSession() {
	if self.currentItem != nil {
		self.abort(self.currentCommand)
	}
	atomic.AddUint64(&self.repo.Stats.CurrentConnections, ^uint64(0))
}

func (self *Controller) ReadFirstMessage() (string, error) {
	return self.rw.Reader.ReadString('\n')
}

func (self *Controller) setCurrentState(cmd *Command, item *queue.Item) {
	self.currentCommand = cmd
	self.currentItem = item
}

func (self *Controller) UnknownCommand() error {
	errorMessage := "ERROR Unknown command"
	self.SendError(errorMessage)
	return errors.New(errorMessage)
}

func (self *Controller) SendError(errorMessage string) {
	fmt.Fprintf(self.rw.Writer, "%s\r\n", errorMessage)
	self.rw.Writer.Flush()
}
