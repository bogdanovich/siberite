package server

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/bogdanovich/siberite/repository"
	"github.com/bogdanovich/siberite/router"
)

type Service struct {
	dataDir string
	ch      chan struct{}
	wg      *sync.WaitGroup
}

func New(dataDir string) *Service {
	s := &Service{dataDir, make(chan struct{}), &sync.WaitGroup{}}
	s.wg.Add(1)
	return s
}

func (s *Service) Serve(listener *net.TCPListener) {
	defer s.wg.Done()

	log.Println("initializing...")

	repo, err := repository.Initialize(s.dataDir)
	log.Println("data directory: ", s.dataDir)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-s.ch:
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
		s.wg.Add(1)
		go router.Dispatch(conn, repo, s.ch, s.wg)
	}
}

func (s *Service) Stop() {
	log.Println("stopping server and finishing work...")
	close(s.ch)
	s.wg.Wait()
}
