package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	service "github.com/bogdanovich/siberite/service"
)

var DataDir = flag.String("data", "./data", "Path to data directory")
var HostAndPort = flag.String("listen", "0.0.0.0:22133", "IP and port to listen")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	laddr, err := net.ResolveTCPAddr("tcp", *HostAndPort)
	if nil != err {
		log.Fatalln(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if nil != err {
		log.Fatalln(err)
	}
	log.Println("listening on", listener.Addr())

	// Make a new service and send it into the background.
	service := service.New(*DataDir)
	go service.Serve(listener)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	service.Stop()
}
