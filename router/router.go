package router

import (
	"bufio"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bogdanovich/siberite/controller"
	"github.com/bogdanovich/siberite/repository"
)

func Dispatch(conn *net.TCPConn, repo *repository.QueueRepository, ch chan struct{}, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	controller := controller.NewSession(rw, repo)
	defer controller.FinishSession()

	for {
		select {
		case <-ch:
			log.Println("Disconnecting", conn.RemoteAddr())
			return
		default:
		}
		conn.SetDeadline(time.Now().Add(1e9))
		message, err := rw.Reader.ReadString('\n')
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			log.Println(conn.RemoteAddr(), err)
			return
		}
		conn.SetDeadline(time.Time{})
		err = dispatch(message, controller)
		if err != nil {
			return
		}
	}
}

func dispatch(message string, controller *controller.Controller) error {
	command := strings.Split(strings.Trim(message, " \r\n"), " ")
	var err error
	switch command[0] {
	case "get":
		err = controller.Get(command)
	case "set":
		err = controller.Set(command)
	case "version":
		err = controller.Version()
	case "stats":
		err = controller.Stats()
	case "delete":
		err = controller.Delete(command)
	case "flush":
		err = controller.Flush(command)
	case "flush_all":
		err = controller.FlushAll(command)
	default:
		return controller.UnknownCommand()
	}

	if err != nil {
		controller.SendError(err.Error())
		return err
	}
	return nil
}
