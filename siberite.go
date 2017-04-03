package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	service "github.com/bogdanovich/siberite/service"
)

var (
	dataDir     = flag.String("data", "./data", "path to data directory")
	hostAndPort = flag.String("listen", "0.0.0.0:22133", "ip and port to listen")
	pidPath     = flag.String("pid", "", "path to PID file to use")
	versionFlag = flag.Bool("version", false, "prints current version")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	service := service.New(*dataDir)

	if *versionFlag {
		fmt.Println(service.Version())
		os.Exit(0)
	}

	// Write a PID file if its requested
	if len(*pidPath) > 0 {
		err := ioutil.WriteFile(*pidPath, []byte(strconv.Itoa(os.Getpid())), 0644)
		if nil != err {
			log.Fatalln(err)
		}
		defer os.Remove(*pidPath)
	}
	
	laddr, err := net.ResolveTCPAddr("tcp", *hostAndPort)
	if nil != err {
		log.Fatalln(err)
	}

	go service.Serve(laddr)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the service gracefully.
	service.Stop()
}
