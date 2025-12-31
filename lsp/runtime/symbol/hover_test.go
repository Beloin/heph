package symbol_test

import (
	"testing"

	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/stretchr/testify/require"
)

func TestFindHelloWorld(t *testing.T) {
	// UTF-8
	currText := "def hello_world():"

	// Find from 'h'
	actualSymbol := symbol.ExtractCurrentSymbol([]byte(currText), 5)

	expectedSymbol := "hello_world"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestFindClassName(t *testing.T) {
	// UTF-8
	currText := `
		class MyHephClass:
			pass
	`

	// Find from 'M'
	actualSymbol := symbol.ExtractCurrentSymbol([]byte(currText), 9)

	expectedSymbol := "MyHephClass"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestNotFindClassName(t *testing.T) {
	// UTF-8
	currText := "class MyHephClass:\n\tpass"

	// Find from def' 'My...
	actualSymbol := symbol.ExtractCurrentSymbol([]byte(currText), 5)

	expectedSymbol := ""
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestFindClassMethod(t *testing.T) {
	// UTF-8
	currText := `myclass.method(param)`

	actualSymbol := symbol.ExtractCurrentSymbol([]byte(currText), 2)

	expectedSymbol := "myclass.method"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestFindVariable(t *testing.T) {
	// UTF-8
	currText := `myvar = 12`

	actualSymbol := symbol.ExtractCurrentSymbol([]byte(currText), 2)

	expectedSymbol := "myvar"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestFindParam(t *testing.T) {
	// UTF-8
	currText := `myclass.method(param)`

	actualSymbol := symbol.ExtractCurrentSymbol([]byte(currText), 18)

	expectedSymbol := "param"
	require.Equal(t, expectedSymbol, actualSymbol)
}
