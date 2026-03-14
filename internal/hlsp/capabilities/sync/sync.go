package sync

import (
	"bytes"
	"errors"

	"github.com/hephbuild/heph/internal/hlsp/runtime"
	"github.com/tliron/glsp"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_synchronization

var (
	ErrInvalidTree = errors.New("invalid tree")
	ErrInvalidDoc  = errors.New("invalid doc")
)

func TextDocumentDidOpenWrapper(manager *runtime.Manager) protocol.TextDocumentDidOpenFunc {
	return func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		parser := manager.Parser
		text := params.TextDocument.Text
		bts := []byte(text)

		if doc, ok := manager.GetDocument(params.TextDocument.URI); ok {
			_, err := manager.SwapTree(doc, bts)
			return errors.Join(ErrInvalidTree, err)
		}

		newTree := parser.Parse(bts, nil)
		if newTree == nil {
			return ErrInvalidTree
		}

		version := params.TextDocument.Version
		_, err := manager.NewDocument(params.TextDocument.URI, newTree, bts, version)
		if err != nil {
			newTree.Close()
			return ErrInvalidTree
		}

		return nil
	}
}

func TextDocumentDidChangeFuncWrapper(manager *runtime.Manager) protocol.TextDocumentDidChangeFunc {
	return func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		doc, ok := manager.GetDocument(params.TextDocument.URI)
		if !ok {
			return ErrInvalidDoc
		}

		for _, change := range params.ContentChanges {
			if event, ok := change.(protocol.TextDocumentContentChangeEvent); ok {
				text := doc.TextString
				insertBytes := []byte(event.Text)

				// TODO: bsena;
				// TO fix this we need to go though each string char and check
				// if it's rune is >= 0x10000 (with DecodeRuneInString) meaning it is
				// a two point rune (utf-16) which is the default/max accepted by lsp protocol

				startByteOffset, endByteOffset := event.Range.IndexesIn(text)
				// TODO: bsena; I think we need to calculate this better, bc its not 1 byte == one char
				endByte := startByteOffset + len(insertBytes)

				editInput := tree_sitter.InputEdit{
					StartByte:  uint(startByteOffset),
					OldEndByte: uint(endByteOffset),
					NewEndByte: uint(endByte),

					StartPosition: tree_sitter.Point{
						Row:    uint(event.Range.Start.Line),
						Column: uint(event.Range.Start.Character),
					},
					OldEndPosition: tree_sitter.Point{
						Row:    uint(event.Range.End.Line),
						Column: uint(event.Range.End.Character),
					},
					NewEndPosition: calcNewEndPosition(insertBytes, tree_sitter.Point{
						Row:    uint(event.Range.Start.Line),
						Column: uint(event.Range.Start.Character),
					}),
				}

				doc.Tree.Edit(&editInput)
				newText := ParseNewBytes(doc.Text, insertBytes, startByteOffset, endByteOffset)

				_, err := manager.SwapTree(doc, newText)
				if err != nil {
					return errors.Join(ErrInvalidTree, err)
				}
			}

			if event, ok := change.(protocol.TextDocumentContentChangeEventWhole); ok {
				bts := []byte(event.Text)
				_, err := manager.SwapTree(doc, bts)
				if err != nil {
					return errors.Join(ErrInvalidTree, err)
				}
			}
		}

		return nil
	}
}

// calcNewEndPosition computes the tree-sitter end point after applying insertBytes
// starting at start. Row advances by the number of newlines in insertBytes;
// Column is the byte count after the last newline (or start.Column + len if no newlines).
func calcNewEndPosition(insertBytes []byte, start tree_sitter.Point) tree_sitter.Point {
	newlines := bytes.Count(insertBytes, []byte{'\n'})
	if newlines == 0 {
		return tree_sitter.Point{
			Row:    start.Row,
			Column: start.Column + uint(len(insertBytes)),
		}
	}
	lastNL := bytes.LastIndexByte(insertBytes, '\n')
	return tree_sitter.Point{
		Row:    start.Row + uint(newlines),
		Column: uint(len(insertBytes) - lastNL - 1),
	}
}

// ParseNewBytes creates a new byte array to insert changes from client
func ParseNewBytes(current, insert []byte, offsetStart, offsetEnd int) []byte {
	result := append(current[:offsetStart:offsetStart], insert...)
	result = append(result, current[offsetEnd:]...)
	return result
}
