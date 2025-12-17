package capabilities

import (
	"bytes"
	"errors"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Following LSP Spec
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_synchronization

var (
	ErrInvalidTree = errors.New("invalid tree")
	ErrInvalidDoc  = errors.New("invalid doc")
)

// protocol.TextDocumentDidOpenFunc
// protocol.TextDocumentDidChangeFunc
// protocol.TextDocumentWillSaveFunc
// protocol.TextDocumentWillSaveWaitUntilFunc
// protocol.TextDocumentDidSaveFunc
// protocol.TextDocumentDidCloseFunc

func TextDocumentDidOpenWrapper(manager *runtime.Manager) protocol.TextDocumentDidOpenFunc {
	logger := commonlog.GetLogger("sync")
	return func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		logger.Noticef("Get File: %s", params.TextDocument.URI)

		parser := manager.Parser
		text := params.TextDocument.Text
		buffer := bytes.NewBufferString(text)

		if doc := manager.GetDocument(params.TextDocument.URI); doc != nil {
			// TODO: bsena; implement this edit when file has changed
			// doc.Tree.Edit(...)
			newTree := parser.Parse(buffer.Bytes(), nil)
			if newTree == nil {
				return ErrInvalidTree
			}

			doc.SwapTree(newTree)

			return nil
		}

		newTree := parser.Parse(buffer.Bytes(), nil)
		// TODO: bsena; we need better err here
		if newTree == nil {
			return ErrInvalidTree
		}

		newDoc := runtime.NewDocument(newTree)
		version := params.TextDocument.Version
		manager.SetDocument(params.TextDocument.URI, version, newDoc)

		return nil
	}
}

func TextDocumentDidChangeFuncWrapper(manager *runtime.Manager) protocol.TextDocumentDidChangeFunc {
	return func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		parser := manager.Parser

		doc := manager.GetDocument(params.TextDocument.URI)
		if doc == nil {
			return ErrInvalidDoc
		}

		for _, change := range params.ContentChanges {
			if _, ok := change.(protocol.TextDocumentContentChangeEvent); ok {
				// TODO: bsena; How to implement this?
				// start := tree_sitter.Point{
				// 	Row:    uint(event.Range.Start.Line),
				// 	Column: uint(event.Range.Start.Character),
				// }

				// doc.Tree.Edit(&tree_sitter.InputEdit{
				// 	StartPosition: start,
				// })
			}

			if event, ok := change.(protocol.TextDocumentContentChangeEventWhole); ok {
				buffer := bytes.NewBufferString(event.Text)
				newTree := parser.Parse(buffer.Bytes(), nil)

				doc.SwapTree(newTree)
			}
		}

		return nil
	}
}
