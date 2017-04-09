package service

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var dir = "./test_data"
var hostAndPort = "127.0.0.1:22140"
var err error

func TestMain(m *testing.M) {
	os.RemoveAll(dir)
	os.Mkdir(dir, 0777)
	result := m.Run()
	os.RemoveAll(dir)
	os.Exit(result)
}

func Test_StartGetVersionAndStop(t *testing.T) {
	s := New(dir)

	laddr, err := net.ResolveTCPAddr("tcp", hostAndPort)
	if err != nil {
		log.Fatalln(err)
	}

	go s.Serve(laddr)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", hostAndPort)
	assert.Nil(t, err)

	fmt.Fprintf(conn, "version\r\n")
	answer, err := bufio.NewReader(conn).ReadString('\n')
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("VERSION %s\r\n", s.Version()), answer)
}
