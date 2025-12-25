package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// protocol.TextDocumentCompletionFunc

func TextDocumentCompletionFuncWrapper(manager *runtime.Manager) protocol.TextDocumentCompletionFunc {
	// TODO: bsena; Add builins as docs in manager doc map like heph://builtin
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		// TODO: search in manager
		// params.TextDocument.URI
		var completionItems []protocol.CompletionItem

		for _, symbol := range manager.BuiltinSymbols {
			name := symbol.Name
			sig := symbol.Signature
			doc := symbol.DocString
			kind := MachineKindToCompletionKind(symbol.Kind)

			// TODO: bsena; How to add parameters etc etc?
			completionItems = append(completionItems, protocol.CompletionItem{
				Label:         name,
				InsertText:    &name,
				Kind:          &kind,
				Detail:        &sig,
				Documentation: doc,
			})
		}

		return completionItems, nil
	}
}
