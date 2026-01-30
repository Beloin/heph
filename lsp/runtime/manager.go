package runtime

import (
	"os"
	"path"
	"slices"
	"strings"
	"sync"

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
	DocumentMap sync.Map
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

	return &Manager{DocumentMap: sync.Map{}, Parser: parser, BuiltinSymbols: builtins}, nil
}

// GetDocument queries and look for an existing document in Manager.
// returns nil if not present
func (m *Manager) GetDocument(uri protocol.DocumentUri) (*document.Document, bool) {
	uri = normalizeDocName(uri)
	if val, ok := m.DocumentMap.Load(uri); ok {
		tuple := val.(*docTuple)
		return tuple.Document, true
	}

	return nil, false
}

func (m *Manager) NewDocument(name string, tree *tree_sitter.Tree, rawText []byte, version int32) (*document.Document, error) {
	name = normalizeDocName(name)
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

func (m *Manager) SwapTree(doc *document.Document, newT *tree_sitter.Tree, newText []byte) (*tree_sitter.Tree, error) {
	t, err := doc.SwapTree(newT, newText)

	if err != nil {
		return nil, err
	}


	// TODO: this will be a problem with update in doc loads...
	// Load all other documents from all doc.Loads
	// go m.loadDocumentsFromLoads(doc)
	m.loadDocumentsFromLoads(doc)

	return t, nil
}

// TODO: bsena; load statement must import at least 1 symbol
// Probably fix Loads to Load and the function loaded, so we need to get only exported loads
// loadDocumentsFromLoads loads documents from load paths
func (m *Manager) loadDocumentsFromLoads(doc *document.Document) {
	for _, rawLoad := range doc.Loads {
		loadPath := rawLoad.Path 
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

			normalizedName := normalizeDocName(filePath)
			// Skip if already loaded
			if fDoc, found := m.GetDocument(protocol.DocumentUri(normalizedName)); found {
				doc.AddLoadedDoc(fDoc, rawLoad.Loads)
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

			// Create new document and ignore if there's an error
			newDoc, err := document.NewDocument(normalizedName, tree, content)
			if err != nil {
				tree.Close()
				continue
			}

			// Cross ref
			doc.AddLoadedDoc(newDoc, rawLoad.Loads)

			m.setDocument(protocol.DocumentUri(normalizedName), 0, newDoc)

			// Recursively load its loads
			m.loadDocumentsFromLoads(newDoc)
		}
	}
}

func (m *Manager) setDocument(uri protocol.DocumentUri, version protocol.Integer, doc *document.Document) {
	if val, ok := m.DocumentMap.Load(uri); ok {
		tuple := val.(*docTuple)
		tuple.Document.Close()

		tuple.Document = doc
		tuple.version = version
	} else {
		m.DocumentMap.Store(uri, &docTuple{
			Document: doc,
			version:  version,
		})
	}
}

type Filter func(s *symbol.Symbol) bool

func (m *Manager) AllLoadedSymbols(filters ...Filter) []*symbol.Symbol {
	allSymbols := m.BuiltinSymbols
	m.DocumentMap.Range(func(key, value interface{}) bool {
		doc := value.(*docTuple)
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
		return true
	})

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

	m.DocumentMap.Range(func(key, value any) bool {
		doc := value.(*docTuple)
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
		return true
	})

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
	var foundDoc *document.Document
	var foundSymbol *symbol.Symbol
	m.DocumentMap.Range(func(key, value any) bool {
		doc := value.(*docTuple)
		if s, found := doc.Document.Query(symbolName); found {
			foundDoc = doc.Document
			foundSymbol = s

			return false
		}

		return true
	})
	if foundDoc != nil {
		return foundDoc, foundSymbol, true
	}

	return nil, nil, false
}

func (m *Manager) QueryBuiltin(symbolName string) (*symbol.Symbol, bool) {
	return symbol.FindSymbol(m.BuiltinSymbols, symbolName)
}

type callQueryResult struct {
	Doc     *document.Document
	Symbols []*symbol.Symbol
}

func (m *Manager) QueryCallsDoc(symbolName string) []callQueryResult {
	res := []callQueryResult{}
	m.DocumentMap.Range(func(key, value any) bool {
		doc := value.(*docTuple)
		if calls := doc.Document.QueryCalls(symbolName); len(calls) > 0 {
			res = append(res, callQueryResult{doc.Document, calls})
		}

		return true
	})

	return res
}

func normalizeDocName(uri string) string {
	uri = strings.TrimPrefix(uri, "file://")

	return uri
}
