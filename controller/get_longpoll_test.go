package controller

import (
	"bufio"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Get_LongPoll_Timeout(t *testing.T) {
    repo, controller, mockTCPConn := setupControllerTest(t, 0)
    defer cleanupControllerTest(repo)

    start := time.Now()
    err = controller.Get([]string{"get", "test/t=30"})
    dur := time.Since(start)
    assert.Nil(t, err)
    assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())
    // should wait at least ~30ms (allow slack)
    if dur < 25*time.Millisecond {
        t.Fatalf("expected wait >=25ms, got %v", dur)
    }
}

func Test_Controller_Get_LongPoll_Notify(t *testing.T) {
    repo, controller, mockTCPConn := setupControllerTest(t, 0)
    defer cleanupControllerTest(repo)

    // Enqueue after a short delay while GET waits
    done := make(chan struct{})
    go func() {
        defer close(done)
        _ = controller.Get([]string{"get", "test/t=200"})
    }()

    time.Sleep(50 * time.Millisecond)
    q, err := repo.GetQueue("test")
    assert.NoError(t, err)
    q.Enqueue([]byte(strconv.Itoa(1)))

    select {
    case <-done:
        // ok
    case <-time.After(500 * time.Millisecond):
        t.Fatalf("GET did not return after enqueue")
    }

    // Verify VALUE response
    rd := bufio.NewReader(&mockTCPConn.WriteBuffer)
    line, _ := rd.ReadString('\n')
    assert.Equal(t, "VALUE test 0 1\r\n", line)
    body, _ := rd.ReadString('\n')
    assert.Equal(t, fmt.Sprintf("%d\r\n", 1), body)
}


