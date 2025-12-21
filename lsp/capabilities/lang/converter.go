package lang

import (
	"github.com/hephbuild/heph/lsp/runtime"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func MachineKindToCompletionKind(kind runtime.SymbolKind) protocol.CompletionItemKind {
	switch kind {
	case runtime.FunctionKind:
		return protocol.CompletionItemKindFunction

	case runtime.VariableKind:
		return protocol.CompletionItemKindVariable

	default:
		return protocol.CompletionItemKindValue
	}
}
