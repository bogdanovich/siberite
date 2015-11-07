package controller

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Controller_Error(t *testing.T) {
	err := &Error{"TEST_ERROR", "Test error message"}
	assert.EqualError(t, err, "TEST_ERROR Test error message")

	err = NewError("TEST_ERROR", errors.New("Error message"))
	assert.EqualError(t, err, "TEST_ERROR Error message")
}
