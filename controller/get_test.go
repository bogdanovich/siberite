package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_parseGetCommand(t *testing.T) {
	testCases := map[string]Command{
		"work":                            Command{},
		"work/open":                       Command{SubCommand: "open"},
		"work/close":                      Command{SubCommand: "close"},
		"work/abort":                      Command{SubCommand: "abort"},
		"work/peek":                       Command{SubCommand: "peek"},
		"work/t=10":                       Command{},
		"work/t=10/t=100/t=22":            Command{},
		"work/t=10/open":                  Command{SubCommand: "open"},
		"work/open/t=10":                  Command{SubCommand: "open"},
		"work/close/open/t=10":            Command{SubCommand: "close/open"},
		"work/close/t=10/open/abort":      Command{SubCommand: "close/open/abort"},
		"work.cg":                         Command{ConsumerGroup: "cg"},
		"work.consumer/open":              Command{SubCommand: "open", ConsumerGroup: "consumer"},
		"work.1/close":                    Command{SubCommand: "close", ConsumerGroup: "1"},
		"work.000/abort":                  Command{SubCommand: "abort", ConsumerGroup: "000"},
		"work.1:2/peek":                   Command{SubCommand: "peek", ConsumerGroup: "1:2"},
		"work.consumergroup/t=10":         Command{ConsumerGroup: "consumergroup"},
		"work.test.cg/t=10/t=100/t=22":    Command{ConsumerGroup: "test"},
		"work.1cg/t=10/open":              Command{SubCommand: "open", ConsumerGroup: "1cg"},
		"work.c/open/t=10":                Command{SubCommand: "open", ConsumerGroup: "c"},
		"work.0/close/open/t=10":          Command{SubCommand: "close/open", ConsumerGroup: "0"},
		"work.group/close/t=1/open/abort": Command{SubCommand: "close/open/abort", ConsumerGroup: "group"},
	}

	for input, command := range testCases {
		cmd := parseGetCommand([]string{"get", input})
		assert.Equal(t, "get", cmd.Name, input)
		assert.Equal(t, "work", cmd.QueueName, input)
		assert.Equal(t, command.SubCommand, cmd.SubCommand, input)
		assert.Equal(t, command.ConsumerGroup, cmd.ConsumerGroup, input)
	}
}

func Test_Controller_Get(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 1)
	defer cleanupControllerTest(repo)

	queueNames := []string{"test.1", "test.cgroup", "test"}

	for _, queueName := range queueNames {
		// When queue has items
		// get test = value
		command := []string{"get", queueName}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// When queue is empty
		// get test = empty
		command = []string{"get", queueName}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// When queue is empty
		// get test/close = empty
		command = []string{"get", queueName + "/close"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// When queue is empty
		// get test/abort = empty
		command = []string{"get", queueName + "/close"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()
	}
}

// Initialize queueName with 4 items
// get queueName/open = value
// get test = error
// get queueName/close = empty
// get queueName/open = value
// get queueName/open = error
// get queueName/abort = empty
// get queueName/open = value
// get queueName/peek = next value
// get queueName/close = empty
func Test_Controller_GetOpen(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 4)
	defer cleanupControllerTest(repo)

	queueNames := []string{"test.1", "test.cgroup", "test"}

	for _, queueName := range queueNames {
		// get queueName/open = value
		command := []string{"get", queueName + "/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get test = error
		command = []string{"get", queueName}
		err = controller.Get(command)
		assert.EqualError(t, err, "CLIENT_ERROR Close current item first")

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/close = value
		command = []string{"get", queueName + "/close"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/open = value
		command = []string{"get", queueName + "/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		// get queueName/open = error
		command = []string{"get", queueName + "/open"}
		err = controller.Get(command)
		assert.EqualError(t, err, "CLIENT_ERROR Close current item first")

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/abort = value
		command = []string{"get", queueName + "/abort"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/open = value
		command = []string{"get", queueName + "/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/peek = value
		command = []string{"get", queueName + "/peek"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n2\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/close = value
		command = []string{"get", queueName + "/close"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()
	}

}

// Initialize test queue with 2 items
// get queueName/open = value
// FinishSession (disconnect)
// NewSession
// get queueName = same value
func Test_Controller_GetOpen_Disconnect(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 2)
	defer cleanupControllerTest(repo)

	queueNames := []string{"test.1", "test.cgroup", "test"}

	for _, queueName := range queueNames {
		// get queueName/open = value
		command := []string{"get", queueName + "/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		controller.FinishSession()

		mockTCPConn = newMockTCPConn()
		controller = NewSession(mockTCPConn, repo)

		// get queueName = same value
		command = []string{"get", queueName}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()
	}
}

// Initialize queueName with 4 items
// get queueName/close/open = value
// get queueName = error
// get queueName/t=10/close/open = value
// get queueName/close/open/t=1000 = next value
// FinishSession (disconnect)
// get queueName/close/t=88/open = same value
func Test_Controller_GetCloseOpen(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 4)
	defer cleanupControllerTest(repo)

	queueNames := []string{"test.1", "test.cgroup", "test"}

	for _, queueName := range queueNames {
		// get queueName/close/open = 1
		command := []string{"get", queueName + "/close/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get queueName = error
		command = []string{"get", queueName}
		err = controller.Get(command)
		assert.EqualError(t, err, "CLIENT_ERROR Close current item first")

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/abort = value
		command = []string{"get", queueName + "/abort"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/t=10/close/open = value
		command = []string{"get", queueName + "/t=10/close/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// get queueName/close/open/t=1000 = next value
		command = []string{"get", queueName + "/close/open/t=1000"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		controller.FinishSession()

		mockTCPConn = newMockTCPConn()
		controller = NewSession(mockTCPConn, repo)

		// get queueName = same value
		command = []string{"get", queueName + "/t=88/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()
	}
}

// Initialize queueName with 2 items
// gets test/open = value
// gets test = error
// GETS test/t=10/close/open = value
func Test_Controller_Gets(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 2)
	defer cleanupControllerTest(repo)

	queueNames := []string{"test.1", "test.cgroup", "test"}

	for _, queueName := range queueNames {
		// gets test/open = 1
		command := []string{"gets", queueName}
		err = controller.Get(command)
		assert.NoError(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// GETS test/t=10/close/open = 2
		command = []string{"GETS", queueName + "/t=10/close/open"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		// GETS test/t=10/close = END
		command = []string{"GETS", queueName + "/t=10/close"}
		err = controller.Get(command)
		assert.Nil(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()
	}
}

// Initialize original queue with 3 items
// Read first value using consumer group
// Read first two items from the source queue
// Continue reading using initial consumer group
// It should return third item of original queue
// Read one more item from consumer goup (returns nothing)
// Add one item to the original queue
// Read an item from consumer group (returns previously added value)
func Test_Controller_ConsumerGroup_Get(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 3)
	defer cleanupControllerTest(repo)

	command := []string{"get", "test.cgroup"}
	err = controller.Get(command)
	assert.NoError(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.NoError(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"get", "test"}
	err = controller.Get(command)
	assert.NoError(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n1\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"get", "test.cgroup"}
	err = controller.Get(command)
	assert.NoError(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n2\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	command = []string{"get", "test.cgroup"}
	err = controller.Get(command)
	assert.NoError(t, err)
	assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

	mockTCPConn.WriteBuffer.Reset()

	q, err := repo.GetQueue("test")
	assert.NoError(t, err)
	q.Enqueue([]byte("3"))

	command = []string{"get", "test.cgroup"}
	err = controller.Get(command)
	assert.NoError(t, err)
	assert.Equal(t, "VALUE test 0 1\r\n3\r\nEND\r\n", mockTCPConn.WriteBuffer.String())

}

// Initialize original queue with 3 items
// Open reliable read using consumer group
// Abort the reliable read using
// 1) <queue>/abort
// 2) <queue>:<cursor>/abort syntax
// Make sure same item gets served again
func Test_Controller_ConsumerGroup_GetAbort(t *testing.T) {
	repo, controller, mockTCPConn := setupControllerTest(t, 3)
	defer cleanupControllerTest(repo)

	abortCommands := []string{"test/abort", "test.cgroup/abort"}

	for _, abortCommand := range abortCommands {

		command := []string{"get", "test.cgroup/open"}
		err = controller.Get(command)
		assert.NoError(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n",
			mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		command = []string{"get", abortCommand}
		err = controller.Get(command)
		assert.NoError(t, err)
		assert.Equal(t, "END\r\n", mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

		command = []string{"get", "test.cgroup/peek"}
		err = controller.Get(command)
		assert.NoError(t, err)
		assert.Equal(t, "VALUE test 0 1\r\n0\r\nEND\r\n",
			mockTCPConn.WriteBuffer.String())

		mockTCPConn.WriteBuffer.Reset()

	}
}
