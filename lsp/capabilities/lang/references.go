package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentReferencesFuncWrapper(manager *runtime.Manager) protocol.TextDocumentReferencesFunc {
	return func(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
		if doc, found := manager.GetDocument(params.TextDocument.URI); found {
			pos := uint(params.Position.IndexIn(doc.TextString))

			if symbolName := doc.ExtractCurrentSymbolName(pos); symbolName != "" {
				return findReferences(manager, symbolName), nil
			}

			if symbolName := doc.ExtractCurrentFunctionName(pos); symbolName != "" {
				return findReferences(manager, symbolName), nil
			}
		}

		return nil, nil
	}
}

// TODO: bsena; Probably we would need to see only those who load current doc. So here we would need
// the DAG, or a tree (bidirecional)
func findReferences(manager *runtime.Manager, symbolName string) []protocol.Location {
	var locations []protocol.Location
	if calls := manager.QueryCallsDoc(symbolName); len(calls) > 0 {
		for _, call := range calls {
			doc := call.Doc
			for _, sym := range call.Symbols {
				loc := protocol.Location{
					URI: doc.FullPath,
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      protocol.UInteger(sym.Position.RowStart),
							Character: protocol.UInteger(sym.Position.ColumnStart),
						},
						End: protocol.Position{
							Line:      protocol.UInteger(sym.Position.RowEnd),
							Character: protocol.UInteger(sym.Position.ColumnEnd),
						},
					},
				}

				locations = append(locations, loc)
			}
		}
	}

	return locations
}
