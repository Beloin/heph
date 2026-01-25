package lang

import (
	"path"
	"strings"

	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var logger = commonlog.GetLogger("lifecycle")

// TODO: bsena; DECLARATIONS AND DEFINITIONS SHOULD GO ONLY TO FILES THAT ARE LOADED WITH `load(...)`
// TODO: bsena; REFERENCES SHOULD SEARCH ONLY TO FILES THAT ARE LOADED WITH `load(...)`

func TextDocumentDeclarationFuncWrapper(manager *runtime.Manager) protocol.TextDocumentDeclarationFunc {
	// TODO: bsena; I think we need to have a parsed heph tree/graph to know where to go when
	// mathecd something like //mgmt/go:protos
	// It can be from a string, from a `load` and so on
	// Probably something from package.Package
	// Or we parse the targets ourself

	return func(context *glsp.Context, params *protocol.DeclarationParams) (any, error) {
		if location, found := extractLocation(manager, params.TextDocument.URI, &params.Position); found {
			logger.Noticef("Declaration location: %v", location)
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
			logger.Noticef("Declaration location: %v", location)
			return location, nil
		}

		return nil, nil
	}
}

func extractLocation(manager *runtime.Manager, uri string, pos *protocol.Position) (*protocol.Location, bool) {
	if doc, found := manager.GetDocument(uri); found {
		pos := uint(pos.IndexIn(doc.TextString))

		// If its target address, open from current workspace
		if literal := doc.ExtractCurrentStringLiteral(pos); literal != "" {
			// Check if its a target location
			if strings.HasPrefix(literal, "//") {

				// Fallback to default BUILD file of that directory
				literal = strings.Split(literal, ":")[0]
				literal = path.Join(literal, "BUILD")
				fullPath := path.Join(manager.WorkspaceFolder, literal)

				return &protocol.Location{
					URI: addProtocol(fullPath),
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      0,
							Character: 0,
						},
						End: protocol.Position{
							Line:      0,
							Character: 0,
						},
					},
				}, true
			}
		}

		if symbolName := doc.ExtractCurrentSymbolName(pos); symbolName != "" {
			// First check loaded documents
			for _, loadedDoc := range doc.DocLoads {
				if sym, found := loadedDoc.Query(symbolName); found {
					return buildLocationFromSymbol(loadedDoc.FullPath, sym), true
				}
			}

			// Fallback to all documents
			if doc, sym, found := manager.QueryDoc(symbolName); found {
				return buildLocationFromSymbol(doc.FullPath, sym), true
			}
		}

		if symbolName := doc.ExtractCurrentFunctionName(pos); symbolName != "" {
			// First check loaded documents
			for _, loadedDoc := range doc.DocLoads {
				logger.Noticef("Found function name: %s", symbolName)
				if sym, found := loadedDoc.Query(symbolName); found {
					return buildLocationFromSymbol(loadedDoc.FullPath, sym), true
				}
			}

			// Fallback to all documents
			if doc, sym, found := manager.QueryDoc(symbolName); found {
				return buildLocationFromSymbol(doc.FullPath, sym), true
			}
		}
	}

	return nil, false
}

func buildLocationFromSymbol(uri string, sym *symbol.Symbol) *protocol.Location {
	return &protocol.Location{
		URI: addProtocol(uri),
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
}

func addProtocol(uri string) string {
	if !strings.HasPrefix(uri, "file://") {
		uri = "file://" + uri
	}
	return uri
}
