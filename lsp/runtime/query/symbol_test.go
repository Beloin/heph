package query_test

import (
	_ "embed"
	"testing"

	"github.com/hephbuild/heph/lsp/runtime/query"
	"github.com/stretchr/testify/require"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// findByteOffset finds the byte offset of a substring in source text
func findByteOffset(source []byte, searchStr string) int {
	for i := 0; i < len(source)-len(searchStr); i++ {
		if string(source[i:i+len(searchStr)]) == searchStr {
			return i
		}
	}
	return -1
}

func newParserForText(t *testing.T, text string) (*tree_sitter.Parser, *tree_sitter.Tree, *tree_sitter.Node) {
	parser := tree_sitter.NewParser()
	// For simple text tests, we need to use a language parser
	// Since we're testing simple text extraction, we'll use Python parser
	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	require.NoError(t, err)

	tree := parser.Parse([]byte(text), nil)
	root := tree.RootNode()
	return parser, tree, root
}

func TestFindClassName(t *testing.T) {
	// UTF-8
	currText := `
		class MyHephClass:
			pass
	`

	_, tree, root := newParserForText(t, currText)
	defer tree.Close()

	// Find from 'M' - byte offset 9
	byteOffset := uint(9)
	actualSymbol := query.ExtractCurrentSymbol(root, []byte(currText), byteOffset)

	expectedSymbol := "MyHephClass"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestFindVariable(t *testing.T) {
	// UTF-8
	currText := `myvar = 12`

	_, tree, root := newParserForText(t, currText)
	defer tree.Close()

	// Find from 'm' - byte offset 2
	byteOffset := uint(2)
	actualSymbol := query.ExtractCurrentSymbol(root, []byte(currText), byteOffset)

	expectedSymbol := "myvar"
	require.Equal(t, expectedSymbol, actualSymbol)
}

//go:embed testdata/test_python_symbols.py
var pythonTestFile []byte

func newPythonParser(t *testing.T) *tree_sitter.Parser {
	parser := tree_sitter.NewParser()
	err := parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_python.Language()))
	require.NoError(t, err)
	return parser
}

func TestExtractCurrentSymbol_FunctionName(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	// Position of 'h' in 'hello_world' (line 4, character 5)
	// def hello_world():
	// 0123456789...
	// The 'h' is at byte offset: count bytes from start of file
	source := pythonTestFile
	// Find the position of 'h' in 'hello_world'
	offset := findByteOffset(source, "hello_world")
	require.NotEqual(t, -1, offset, "Should find 'hello_world' in source")

	// Get the 'h' character (first character of hello_world)
	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentSymbol(root, source, byteOffset)

	expectedSymbol := "hello_world"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestExtractCurrentSymbol_ClassName(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	source := pythonTestFile
	// Find the position of 'M' in 'MyHephClass'
	offset := findByteOffset(source, "MyHephClass")
	require.NotEqual(t, -1, offset, "Should find 'MyHephClass' in source")

	// Get the 'M' character (first character of MyHephClass)
	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentSymbol(root, source, byteOffset)

	expectedSymbol := "MyHephClass"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestExtractCurrentSymbol_MethodName(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	source := pythonTestFile
	// Find the position of 'm' in 'my_method'
	offset := findByteOffset(source, "my_method")
	require.NotEqual(t, -1, offset, "Should find 'my_method' in source")

	// Get the 'm' character (first character of my_method)
	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentSymbol(root, source, byteOffset)

	expectedSymbol := "my_method"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestExtractCurrentSymbol_VariableName(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	source := pythonTestFile
	// Find the position of 'm' in 'my_variable'
	offset := findByteOffset(source, "my_variable")
	require.NotEqual(t, -1, offset, "Should find 'my_variable' in source")

	// Get the 'm' character (first character of my_variable)
	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentSymbol(root, source, byteOffset)

	expectedSymbol := "my_variable"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestExtractCurrentSymbol_FunctionParameter(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	source := pythonTestFile
	// Find the position of 'p' in 'param1' (in my_method parameters)
	offset := findByteOffset(source, "param1")
	require.NotEqual(t, -1, offset, "Should find 'param1' in source")

	// Get the 'p' character (first character of param1)
	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentSymbol(root, source, byteOffset)

	expectedSymbol := "param1"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestExtractCurrentSymbol_AnotherFunction(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	source := pythonTestFile
	// Find the position of 'a' in 'another_function'
	offset := findByteOffset(source, "another_function")
	require.NotEqual(t, -1, offset, "Should find 'another_function' in source")

	// Get the 'a' character (first character of another_function)
	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentSymbol(root, source, byteOffset)

	expectedSymbol := "another_function"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestExtractCurrentSymbol_StaticMethod(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	source := pythonTestFile
	// Find the position of 's' in 'static_method'
	offset := findByteOffset(source, "static_method")
	require.NotEqual(t, -1, offset, "Should find 'static_method' in source")

	// Get the 's' character (first character of static_method)
	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentSymbol(root, source, byteOffset)

	expectedSymbol := "static_method"
	require.Equal(t, expectedSymbol, actualSymbol)
}

func TestExtractStringLiteral(t *testing.T) {
	parser := newPythonParser(t)
	tree := parser.Parse(pythonTestFile, nil)
	defer tree.Close()
	root := tree.RootNode()

	source := pythonTestFile
	searchStr := "//heph/s"
	offset := findByteOffset(source, searchStr)
	require.NotEqual(t, -1, offset, "Should find '//heph/string/literal' in source")

	byteOffset := uint(offset)
	actualSymbol := query.ExtractCurrentStringLiteral(root, source, byteOffset)

	// Should return empty string at whitespace
	require.Equal(t, "//heph/string/literal", actualSymbol)
}
