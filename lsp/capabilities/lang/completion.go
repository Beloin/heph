package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// protocol.TextDocumentCompletionFunc

func TextDocumentCompletionFuncWrapper(manager *runtime.Manager) protocol.TextDocumentCompletionFunc {
	return func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		// TODO: bsena; put the first from this uri
		var completionItems []protocol.CompletionItem

		// TODO: bsena; Read based in position so we can get classes' methods

		for _, symbol := range manager.AllLoadedSymbols() {
			name := symbol.Name
			sig := symbol.Signature
			doc := symbol.DocString
			kind := MachineKindToCompletionKind(symbol.Kind)

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
