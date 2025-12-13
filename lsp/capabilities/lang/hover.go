package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentHoverFuncWrapper(manager *runtime.Manager) protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc, ok := manager.GetDocument(params.TextDocument.URI)
		if !ok {
			return &protocol.Hover{}, nil
		}

		bytePos := params.Position.IndexIn(doc.TextString)
		// TODO: bsena; Make hover context-aware
		// look for target function etc

		// TODO: bsena; also query for target //build/target:run
		// For now extract the targets from each document loaded into a target spec, so we can load it in runtime?
		// In the future is the best to have a Heph Server that runs and change at each file change, so we always have a fast
		// DAG available

		symbolName := doc.ExtractCurrentSymbolName(uint(bytePos))

		// Query first for current document symbols
		if symbol, found := doc.Query(symbolName); found {
			return createHover(symbol), nil
		}

		if symbol, found := manager.Query(symbolName); found {
			return createHover(symbol), nil
		}

		return nil, nil
	}
}

func createHover(symbol *symbol.Symbol) *protocol.Hover {
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: symbol.Source + "\n---\n" + langDecorate(symbol.Signature, runtime.HephLanguage) + "\n---\n" + symbol.DocString,
		},
		Range: &protocol.Range{
			Start: protocol.Position{
				Line:      protocol.UInteger(symbol.Position.RowStart),
				Character: protocol.UInteger(symbol.Position.ColumnStart),
			},
			End: protocol.Position{
				Line:      protocol.UInteger(symbol.Position.RowEnd),
				Character: protocol.UInteger(symbol.Position.ColumnEnd),
			},
		},
	}
}

func langDecorate(text, lang string) string {
	return "```" + lang + "\n" + text + "\n```\n"
}
