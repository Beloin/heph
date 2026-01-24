package runtime

import (
	"os"
	"path"
	"slices"
	"strings"

	"github.com/hephbuild/heph/lsp/runtime/builtin"
	"github.com/hephbuild/heph/lsp/runtime/document"
	"github.com/hephbuild/heph/lsp/runtime/symbol"
	"github.com/hephbuild/heph/specs"
	protocol "github.com/tliron/glsp/protocol_3_16"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const HephLanguage = "heph"

var Version = "0.0.1"

type docTuple struct {
	Document *document.Document

	// Last reported version
	version protocol.Integer
}

// TODO: bsena; Remove protocol usage here, use raw values

type Manager struct {
	DocumentMap map[protocol.DocumentUri]*docTuple // TODO: bsena; use sync.Map
	TargetMap   map[string]*specs.Target

	BuiltinSymbols []*symbol.Symbol
	Parser         *tree_sitter.Parser

	WorkspaceFolder string

	// dag.DAG // TODO: bsena; use the dag to know which symbols we can import into that specific BUILD file
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

func (m *Manager) NewDocument(name string, tree *tree_sitter.Tree, rawText []byte, version int32) (*document.Document, error) {
	newDoc, err := document.NewDocument(name, tree, rawText)
	if err != nil {
		return nil, err
	}

	m.setDocument(name, version, newDoc)

	// Load all other documents from all newDoc.Loads
	// go m.loadDocumentsFromLoads(newDoc)
	m.loadDocumentsFromLoads(newDoc)

	return newDoc, nil
}

// loadDocumentsFromLoads loads documents from load paths
func (m *Manager) loadDocumentsFromLoads(doc *document.Document) {
	for _, loadPath := range doc.Loads {
		if loadPath == "" {
			continue
		}

		// Convert load path to folder path
		folderPath := path.Join(m.WorkspaceFolder, loadPath)

		// Read directory
		entries, err := os.ReadDir(folderPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			// TODO: bsena; Probably need to read for BUILD.* or *.BUILD
			if !strings.Contains(entry.Name(), "BUILD") {
				continue
			}

			filePath := path.Join(folderPath, entry.Name())

			// Skip if already loaded
			if _, found := m.GetDocument(protocol.DocumentUri(filePath)); found {
				continue
			}

			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			// Parse the file
			tree := m.Parser.Parse(content, nil)
			if tree == nil {
				continue
			}

			// Create new document
			newDoc, err := document.NewDocument(filePath, tree, content)
			if err != nil {
				tree.Close()
				continue
			}

			m.setDocument(protocol.DocumentUri(filePath), 0, newDoc)
			doc.AddLoadedDoc(newDoc)

			// Recursively load its loads
			m.loadDocumentsFromLoads(newDoc)
		}
	}
}

func (m *Manager) setDocument(uri protocol.DocumentUri, version protocol.Integer, doc *document.Document) {
	if tuple, found := m.DocumentMap[uri]; found {
		tuple.Document.Close()

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

func (m *Manager) AllLoadedSymbols(filters ...Filter) []*symbol.Symbol {
	allSymbols := m.BuiltinSymbols
	for _, doc := range m.DocumentMap {
		// allSymbols = slices.Concat(allSymbols, doc.Symbols)

		for _, smb := range doc.Document.Symbols {
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

// kindStruct is a struct helper to get symbol per kind
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
		for _, smb := range doc.Document.Symbols {
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
		Variables:  vars,
		Functions:  funs,
	}
}

// TODO: bsena; use a prefix tree and have ALL nodes, even child nodes
// in that tree using Symbol.Fullname as index
// Also how to work with imports?
// Probalby this tree will be in Manager's struct
// add to /heph/utils/trie
func (m *Manager) Query(symbolName string) (*symbol.Symbol, bool) {
	if s, found := symbol.FindSymbol(m.BuiltinSymbols, symbolName); found {
		return s, true
	}

	if _, s, found := m.QueryDoc(symbolName); found {
		return s, found
	}

	return nil, false
}

func (m *Manager) QueryDoc(symbolName string) (*document.Document, *symbol.Symbol, bool) {
	for _, doc := range m.DocumentMap {
		if s, found := doc.Document.Query(symbolName); found {
			return doc.Document, s, true
		}
	}

	return nil, nil, false
}
