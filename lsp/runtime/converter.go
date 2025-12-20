package runtime

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func MachineKindToProtocolKind(kind string) protocol.SymbolKind {
	switch kind {
	case FunctionKind:
		return protocol.SymbolKindFunction

	case VariableKind:
		return protocol.SymbolKindVariable

	default:
		return -1
	}
}
