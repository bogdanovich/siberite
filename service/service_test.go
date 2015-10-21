package service

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dir = "./test_data"
var hostAndPort = "127.0.0.1:22140"
var err error

func TestMain(m *testing.M) {
	os.Mkdir(dir, 0777)
	result := m.Run()
	os.RemoveAll(dir)
	os.Exit(result)
}

func Test_StartGetVersionAndStop(t *testing.T) {
	s := New(dir)

	laddr, err := net.ResolveTCPAddr("tcp", hostAndPort)
	if nil != err {
		log.Fatalln(err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if nil != err {
		log.Fatalln(err)
	}
	log.Println("listening on", listener.Addr())

	go s.Serve(listener)
	defer s.Stop()

	conn, err := net.Dial("tcp", hostAndPort)
	assert.Nil(t, err)

	fmt.Fprintf(conn, "version\r\n")
	answer, err := bufio.NewReader(conn).ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("VERSION %s\r\n", s.Version()), answer)
}
