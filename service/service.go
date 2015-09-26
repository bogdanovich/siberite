package server

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/bogdanovich/siberite/controller"
	"github.com/bogdanovich/siberite/repository"
)

type Service struct {
	dataDir string
	repo    *repository.QueueRepository
	ch      chan struct{}
	wg      *sync.WaitGroup
}

func New(dataDir string) *Service {
	s := &Service{
		dataDir: dataDir,
		repo:    &repository.QueueRepository{},
		ch:      make(chan struct{}),
		wg:      &sync.WaitGroup{},
	}
	s.wg.Add(1)
	return s
}

func (self *Service) Serve(listener *net.TCPListener) {
	defer self.wg.Done()

	log.Println("initializing...")
	var err error
	self.repo, err = repository.Initialize(self.dataDir)
	log.Println("data directory: ", self.dataDir)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-self.ch:
			log.Println("stopping listening on", listener.Addr())
			listener.Close()
			return
		default:
		}
		listener.SetDeadline(time.Now().Add(1e9))
		conn, err := listener.AcceptTCP()
		if nil != err {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
		}
		log.Println(conn.RemoteAddr(), "connected")
		self.wg.Add(1)
		go self.HandleConnection(conn)
	}
}

func (self *Service) Stop() {
	log.Println("stopping server and finishing work...")
	close(self.ch)
	self.wg.Wait()
}

func (self *Service) HandleConnection(conn *net.TCPConn) {
	defer conn.Close()
	defer self.wg.Done()

	controller := controller.NewSession(conn, self.repo)
	defer controller.FinishSession()

	for {
		select {
		case <-self.ch:
			log.Println("Disconnecting", conn.RemoteAddr())
			return
		default:
		}
		err := controller.Dispatch()
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			log.Println(conn.RemoteAddr(), err)
			return
		}
	}
}

func (self *Service) Version() string {
	return repository.Version
}
