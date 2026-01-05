package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentDeclarationFuncWrapper(manager *runtime.Manager) protocol.TextDocumentDeclarationFunc {
	// TODO: bsena; I think we need to have a parsed heph tree/graph to know where to go when
	// mathecd something like //mgmt/go:protos
	// It can be from a string, from a `load` and so on
	// Probably something from package.Package
	// Or we parse the targets ourself

	return func(context *glsp.Context, params *protocol.DeclarationParams) (any, error) {
		if location, found := extractLocation(manager, params.TextDocument.URI, &params.Position); found {
			return location, nil
		}

		return nil, nil
	}
}

func TextDocumentDefinitionFuncWrapper(manager *runtime.Manager) protocol.TextDocumentDefinitionFunc {
	// TODO: bsena; I think we need to have a parsed heph tree/graph to know where to go when
	// mathecd something like //mgmt/go:protos
	// It can be from a string, from a `load` and so on
	// Probably something from package.Package
	// Or we parse the targets ourself

	return func(context *glsp.Context, params *protocol.DefinitionParams) (any, error) {
		if location, found := extractLocation(manager, params.TextDocument.URI, &params.Position); found {
			return location, nil
		}

		return nil, nil
	}
}

func extractLocation(manager *runtime.Manager, uri string, pos *protocol.Position) (*protocol.Location, bool) {
	if doc, found := manager.GetDocument(uri); found {
		pos := uint(pos.IndexIn(doc.TextString))

		if symbolName := doc.ExtractCurrentSymbolName(pos); symbolName != "" {
			if doc, sym, found := manager.QueryDoc(symbolName); found {
				return &protocol.Location{
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
				}, true
			}
		}
	}

	return nil, false
}
