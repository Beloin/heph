package sync

import (
	"errors"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/lsp/runtime/document"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Following LSP Spec
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_synchronization

var (
	ErrInvalidTree = errors.New("invalid tree")
	ErrInvalidDoc  = errors.New("invalid doc")
)

var SyncLogger = commonlog.GetLogger("sync")

// protocol.TextDocumentDidOpenFunc            | Mandatory
// protocol.TextDocumentDidChangeFunc          | Mandatory
// protocol.TextDocumentWillSaveFunc
// protocol.TextDocumentWillSaveWaitUntilFunc
// protocol.TextDocumentDidSaveFunc
// protocol.TextDocumentDidCloseFunc           | Mandatory

// TODO: bsena; split this better so we can test it

func TextDocumentDidOpenWrapper(manager *runtime.Manager) protocol.TextDocumentDidOpenFunc {
	// TODO: bsena; We are panicking when we read invalid file, why?
	
	return func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		SyncLogger.Noticef("Get File: %s", params.TextDocument.URI)

		parser := manager.Parser
		text := params.TextDocument.Text
		bts := []byte(text)

		if doc, ok := manager.GetDocument(params.TextDocument.URI); ok {
			newTree := parser.Parse(bts, nil)
			if newTree == nil {
				return ErrInvalidTree
			}

			_, err := doc.SwapTree(newTree, bts)

			return err
		}

		newTree := parser.Parse(bts, nil)
		// TODO: bsena; we need better err here
		if newTree == nil {
			return ErrInvalidTree
		}

		newDoc, err := document.NewDocument(newTree, bts)
		if err != nil {
			return err
		}

		version := params.TextDocument.Version
		manager.SetDocument(params.TextDocument.URI, version, newDoc)

		return nil
	}
}

// TODO: bsena; Not working for now, maybe we need "did open", "did close", save etc
func TextDocumentDidChangeFuncWrapper(manager *runtime.Manager) protocol.TextDocumentDidChangeFunc {
	return func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		SyncLogger.Noticef("received TextDocumentDidChange")

		parser := manager.Parser

		doc, ok := manager.GetDocument(params.TextDocument.URI)
		if !ok {
			return ErrInvalidDoc
		}

		for _, change := range params.ContentChanges {
			if event, ok := change.(protocol.TextDocumentContentChangeEvent); ok {
				text := doc.TextString
				insertBytes := []byte(event.Text)

				SyncLogger.Noticef("TextDocumentContentChangeEvent: txt=%s", text)

				startByteOffset, endByteOffset := event.Range.IndexesIn(text)
				endByte := uint(endByteOffset) + uint(len(insertBytes))

				editInput := tree_sitter.InputEdit{
					StartByte:  uint(startByteOffset),
					OldEndByte: uint(endByteOffset),
					NewEndByte: endByte,

					StartPosition: tree_sitter.Point{
						Row:    uint(event.Range.Start.Line),
						Column: uint(event.Range.Start.Character),
					},
					OldEndPosition: tree_sitter.Point{
						Row:    uint(event.Range.End.Line),
						Column: uint(event.Range.End.Character),
					},
					NewEndPosition: tree_sitter.Point{
						Row:    uint(event.Range.End.Line),
						Column: endByte * 2, // TODO: bsena; this is wrong, fix it
					},
				}

				doc.Tree.Edit(&editInput)
				newText := InsertByteArray(doc.Text, insertBytes, startByteOffset)
				SyncLogger.Noticef("TextDocumentContentChangeEvent: newtxt=%s", string(newText))
				newTree := parser.Parse(newText, doc.Tree)

				_, err := doc.SwapTree(newTree, newText)
				if err != nil {
					return err
				}

			}

			if event, ok := change.(protocol.TextDocumentContentChangeEventWhole); ok {
				SyncLogger.Info("TextDocumentContentChangeEvent: txt=%s", event.Text)

				bts := []byte(event.Text)
				newTree := parser.Parse(bts, nil)

				_, err := doc.SwapTree(newTree, bts)
				if err != nil {
					return err
				}

			}
		}

		return nil
	}
}

func InsertByteArray(curr []byte, insertBytes []byte, insertStart int) []byte {
	// Copy the new edit text
	fullsize := len(curr) + len(insertBytes)
	newSlice := make([]byte, 0, fullsize)
	insertIndex := 0

	for i := 0; i < fullsize; i++ {
		if i >= insertStart && insertIndex < len(insertBytes) {
			newSlice = append(newSlice, insertBytes[insertIndex])
			insertIndex++
			continue
		}

		if i < len(curr) {
			newSlice = append(newSlice, curr[i])
		}
	}

	return newSlice
}
