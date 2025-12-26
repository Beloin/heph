package sync_test

import (
	"testing"

	"github.com/hephbuild/heph/lsp/capabilities/sync"
	"github.com/stretchr/testify/require"
)

func TestByteInsert(t *testing.T) {
	// UTF-8
	currText := "def hello_world():"
	newText := "world_hello"
	currBytes := []byte(currText)
	newBytes := []byte(newText)
	insertedArray := sync.InsertByteArray(currBytes, newBytes, 4)

	expectedText := "def world_hello():"
	actualText := string(insertedArray)
	require.Equal(t, expectedText, actualText)
}
