package runtime

import (
	"slices"

	"github.com/hephbuild/heph/lsp/runtime/builtin"
	"github.com/hephbuild/heph/lsp/runtime/document"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	protocol "github.com/tliron/glsp/protocol_3_16"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const HephLanguage = "heph"

var Version = "0.0.1"

type docTuple struct {
	*document.Document

	// Last reported version
	version protocol.Integer
}

type Manager struct {
	BuiltinSymbols []*symbol.Symbol
	DocumentMap    map[protocol.DocumentUri]*docTuple // TODO: bsena; use sync.Map
	Parser         *tree_sitter.Parser
}

func NewManager(parser *tree_sitter.Parser) (*Manager, error) {
	builtins, err := builtin.ParseBuiltins(parser)
	if err != nil {
		return nil, err
	}

	return &Manager{DocumentMap: map[protocol.DocumentUri]*docTuple{}, Parser: parser, BuiltinSymbols: builtins}, nil
}

// GetDocument queries and look for an existing document in Manager.
// returns nil if not present
func (m *Manager) GetDocument(uri protocol.DocumentUri) (*document.Document, bool) {
	if tuple, found := m.DocumentMap[uri]; found {
		return tuple.Document, true
	}

	return nil, false
}

func (m *Manager) SetDocument(uri protocol.DocumentUri, version protocol.Integer, doc *document.Document) {
	if tuple, found := m.DocumentMap[uri]; found {
		tuple.Document = doc
		tuple.version = version
	} else {
		m.DocumentMap[uri] = &docTuple{
			Document: doc,
			version:  version,
		}
	}
}

type Filter func(s *symbol.Symbol) bool

// TODO: bsena; Return items based on Kind
func (m *Manager) AllLoadedSymbols(filters ...Filter) []*symbol.Symbol {
	allSymbols := m.BuiltinSymbols
	for _, doc := range m.DocumentMap {
		// allSymbols = slices.Concat(allSymbols, doc.Symbols)

		for _, smb := range doc.Symbols {
			shoulAdd := true
			for _, f := range filters {
				if !f(smb) {
					shoulAdd = false
				}
			}

			if shoulAdd {
				allSymbols = append(allSymbols, smb)
			}

		}
	}

	return allSymbols
}

type kindStruct struct {
	AllSymbols []*symbol.Symbol
	Variables  []*symbol.Symbol
	Functions  []*symbol.Symbol
}

func (m *Manager) AllLoadedSymbolsPerKind() kindStruct {
	allSymbols := []*symbol.Symbol{}
	vars := []*symbol.Symbol{}
	funs := []*symbol.Symbol{}

	for _, smb := range m.BuiltinSymbols {
		switch smb.Kind {
		case symbol.FunctionKind:
			funs = append(funs, smb)
		case symbol.VariableKind:
			vars = append(vars, smb)
		default:
			allSymbols = append(allSymbols, smb)
		}
	}

	for _, doc := range m.DocumentMap {
		for _, smb := range doc.Symbols {
			switch smb.Kind {
			case symbol.FunctionKind:
				funs = append(funs, smb)
			case symbol.VariableKind:
				vars = append(vars, smb)
			default:
				allSymbols = append(allSymbols, smb)
			}
		}
	}

	allSymbols = slices.Concat(allSymbols, vars, funs)

	return kindStruct{
		AllSymbols: allSymbols,
		Variables: vars,
		Functions: funs,
	}
}

// TODO: bsena; use a prefix tree and have ALL nodes, even child nodes
// in that tree using Symbol.Fullname as index
// Also how to work with imports?
// Probalby this tree will be in Manager's struct
func (m *Manager) Query(symbolName string) (*symbol.Symbol, bool) {
	if s, found := symbol.FindSymbol(m.BuiltinSymbols, symbolName); found {
		return s, true
	}

	for _, doc := range m.DocumentMap {
		if s, found := doc.Query(symbolName); found {
			return s, true
		}
	}

	return nil, false
}
