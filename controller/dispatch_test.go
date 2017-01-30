package controller

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Dispatch(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 0)
	defer cleanupControllerTest(repo)

	// Command: set test 0 0 1
	// 1
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "set test 0 0 1\r\n1\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "STORED\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: SET test 0 0 2
	// 20
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "SET test 0 0 2\r\n")
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "20\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "STORED\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: SET test 0 0 2
	// ab
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "SET test 0 0 2\r\n")
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "ab\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "STORED\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: set test 0 0 10
	// 123
	// 12
	// 1
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "set test 0 0 10\r\n")
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "123\r\n")
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "12\r\n")
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "1\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "STORED\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: get test
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "get test\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: get test/open
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "get test/open\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 2\r\n20\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: GET test/abort
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "GET test/abort\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: get test/open
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "get test/open\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 2\r\n20\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: get test/close
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "get test/close\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: gets test/close/open
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "gets test/close/open\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "VALUE test 0 2\r\nab\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: version
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "version\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "VERSION "+repo.Stats.Version+"\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: STATS
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "STATS\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	response, _ := mockTCPConn.WriteBuffer.ReadString('\n')
	fmt.Println(response)
	assert.True(t, strings.HasPrefix(response, "STAT uptime"))

	mockTCPConn.WriteBuffer.Reset()

	// Command: flush test
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "flush test\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: DELETE test
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "DELETE test\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	// Command: flush_all
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "flush_all\r\n")
	err = controller.Dispatch()
	assert.Nil(t, err)
	assert.Equal(t, "Flushed all queues.\r\n", mockTCPConn.WriteBuffer.String())

	// Command: quit
	fmt.Fprintf(&mockTCPConn.ReadBuffer, "quit\r\n")
	err = controller.Dispatch()
	assert.Error(t, err, "Quit command received")

	mockTCPConn.WriteBuffer.Reset()
}
