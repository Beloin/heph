package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var HoverLogger = commonlog.GetLogger("hover")

func TextDocumentHoverFuncWrapper(manager *runtime.Manager) protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc, ok := manager.GetDocument(params.TextDocument.URI)
		if !ok {
			return &protocol.Hover{}, nil
		}

		bytePos := params.Position.IndexIn(doc.TextString)
		symbolName := symbol.ExtractCurrentSymbol(doc.Text, bytePos)
		HoverLogger.Noticef("looking for %q", symbolName)

		if symbol, found := doc.Query(symbolName); found {
			HoverLogger.Noticef("Found %q", symbol.Signature)
			hover := protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindPlainText,
					Value: symbol.Signature + "\n" + symbol.DocString,
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

			return &hover, nil
		}

		return nil, nil
	}
}
