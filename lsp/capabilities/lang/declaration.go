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
		return nil, nil
	}
}
