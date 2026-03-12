package document

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/hephbuild/heph/internal/hlsp/runtime/builtin"
	"github.com/hephbuild/heph/internal/hlsp/runtime/driver"
	"github.com/hephbuild/heph/internal/hlsp/runtime/query"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Document struct {
	// FullPath
	FullPath string

	Symbols []*symbol.Symbol
	Calls   []*symbol.Symbol
	Targets []*symbol.Symbol

	// Loads are the BUILD path,
	// if any file has it name changed or delete,
	// Loads are not updated until next file sync, so use DocLoads
	Loads      []*RawLoad
	DocLoads   map[string]*Load
	IsLoadedBy map[string]*Load

	Tree       *tree_sitter.Tree
	Text       []byte // UTF-16 encoded byte array https://microsoft.github.io/language-server-protocol/specifications/specification-3-16/#textDocuments
	TextString string // UTF-16 encoded string https://microsoft.github.io/language-server-protocol/specifications/specification-3-16/#textDocuments

	treeMutex sync.Mutex
	loadsMu   sync.RWMutex

	drivers *driver.Registry
}

type RawLoad struct {
	Path  string
	Loads []string
}

type Load struct {
	Doc   *Document
	Loads []string
}

func (l *Load) Hash() string {
	return l.Doc.FullPath + "|" + strings.Join(l.Loads, "|")
}

// lockTwoLoads locks both docs mutexes in a deterministic way. Ugly, but keep the caller from undestading how call the locks
func lockTwoLoads(mu1, mu2 *sync.RWMutex) {
	p1 := uintptr(unsafe.Pointer(mu1))
	p2 := uintptr(unsafe.Pointer(mu2))
	if p1 < p2 {
		mu1.Lock()
		mu2.Lock()
	} else {
		mu2.Lock()
		mu1.Lock()
	}
}

func unlockTwoLoads(mu1, mu2 *sync.RWMutex) {
	mu1.Unlock()
	mu2.Unlock()
}

func (d *Document) Close() {
	d.Tree.Close()
}

func NewDocument(name string, tree *tree_sitter.Tree, rawText []byte, drivers *driver.Registry) (*Document, error) {
	doc := &Document{
		FullPath: name, Tree: tree, Text: rawText, TextString: string(rawText),
		DocLoads: make(map[string]*Load), IsLoadedBy: make(map[string]*Load),
		drivers: drivers,
	}

	syms, calls, targets, err := extractSymbols(doc.Tree, doc.Text, doc.FullPath, doc)
	doc.Symbols = syms
	doc.Calls = calls
	doc.Targets = targets
	doc.extractLoads()
	go doc.extractTargets()

	return doc, err
}

// SwapTree atomic swaps current tree and return the closed old tree
func (d *Document) SwapTree(newT *tree_sitter.Tree, newText []byte) (*tree_sitter.Tree, error) {
	d.treeMutex.Lock()
	defer d.treeMutex.Unlock()

	oldTree := d.Tree

	syms, calls, targets, err := extractSymbols(newT, newText, d.FullPath, d)
	if err != nil {
		return nil, err
	}

	d.Symbols = syms
	d.Calls = calls
	d.Targets = targets
	d.Tree = newT
	d.Text = newText
	d.TextString = string(newText)

	d.resetLoads()
	d.extractLoads()
	go d.extractTargets()

	oldTree.Close()

	return oldTree, err
}

func (d *Document) extractTargets() {
	resolved := make([]*symbol.Symbol, 0, len(d.Targets))

	// TODO: bsena; add error treatment
	btTarget := builtin.GetTarget()[0]

	for _, call := range d.Targets {
		var driverName string
		for _, p := range call.Parameters {
			if p.Name == builtin.DriverName {
				driverName = strings.Trim(p.Value, "\"'")
				break
			}
		}

		// Ignore any invalid driver
		if driverName == "" {
			defaultTarget := *btTarget
			resolved = append(resolved, &defaultTarget)

			continue
		}

		schema, found := d.getDriverSchema(driverName)
		if !found {
			// Back to what it was
			schema = call
			schema.Signature += " // driver `" + driverName + "` not found"
		}

		// Preserve the call's position so position-based lookups still work.
		schema.Position = call.Position
		schema.DocString = btTarget.DocString

		resolved = append(resolved, schema)
	}

	// Sorting in order of position in file
	slices.SortFunc(resolved, func(i, j *symbol.Symbol) int {
		if i.Position.ByteStart < j.Position.ByteStart {
			return -1
		}

		if i.Position.ByteStart > j.Position.ByteStart {
			return 1
		}

		return 0
	})

	d.treeMutex.Lock()
	defer d.treeMutex.Unlock()

	d.Targets = resolved
}

func (d *Document) getDriverSchema(driverName string) (*symbol.Symbol, bool) {
	if drv, ok := d.drivers.GetDriver(driverName); ok {
		ctx, cancel := context.WithTimeout(context.TODO(), 200*time.Millisecond)
		defer cancel()

		s, err := driver.SymbolFromTargetDriver(ctx, drv)
		if err == nil {
			return s, true
		}
	}

	return nil, false
}

func extractSymbols(tree *tree_sitter.Tree, text []byte, source string, resolver query.SymbolResolver) ([]*symbol.Symbol, []*symbol.Symbol, []*symbol.Symbol, error) {
	symbols, err := query.QuerySymbols(tree, text, source, resolver)
	if err != nil {
		return nil, nil, nil, err
	}

	allCalls, err := query.QueryCalls(tree, text, source)
	if err != nil {
		return nil, nil, nil, err
	}

	var calls, targets []*symbol.Symbol
	for _, call := range allCalls {
		if call.Name == builtin.TargetName {
			call.Kind = symbol.TargetCallKind
			targets = append(targets, call)
		} else {
			calls = append(calls, call)
		}
	}

	return symbols, calls, targets, err
}

func (d *Document) extractLoads() {
	loads := []*RawLoad{}
	for _, call := range d.Calls {
		if call.Name == builtin.LoadName {
			if len(call.Parameters) < 2 {
				continue
			}

			rawValue := call.Parameters[0].Value
			path := strings.Trim(rawValue, "\"")
			loadsSlice := []string{}
			for i := 1; i < len(call.Parameters); i++ {
				rawFunction := call.Parameters[i].Value
				fun := strings.Trim(rawFunction, "\"")
				loadsSlice = append(loadsSlice, fun)
			}

			loads = append(loads, &RawLoad{Path: path, Loads: loadsSlice})
		}
	}

	d.Loads = loads
}

func (d *Document) ExtractCurrentStringLiteral(byteOffSet uint) string {
	return query.ExtractCurrentStringLiteral(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) ExtractCurrentSymbol(byteOffSet uint) (*symbol.Symbol, bool) {
	sName := d.ExtractCurrentSymbolName(byteOffSet)
	return d.Query(sName)
}

func (d *Document) ExtractCurrentSymbolName(byteOffSet uint) string {
	return query.ExtractCurrentSymbol(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) ExtractCurrentFunctionName(byteOffSet uint) string {
	return query.ExtractFunctionNameFromOffset(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) Query(symbolName string) (*symbol.Symbol, bool) {
	return symbol.FindSymbol(d.Symbols, symbolName)
}

// QueryType resolves a type name: primitives first, then own symbols, then loaded doc symbols.
// Implements query.SymbolResolver.
func (d *Document) QueryType(name string) (*symbol.Type, bool) {
	if p := symbol.ResolveType(name); p != nil {
		return p, true
	}

	if s, found := symbol.FindSymbol(d.Symbols, name); found {
		return &s.Type, true
	}

	d.loadsMu.RLock()
	defer d.loadsMu.RUnlock()
	for _, load := range d.DocLoads {
		if s, found := symbol.FindSymbol(load.Doc.Symbols, name); found {
			return &s.Type, true
		}
	}

	return nil, false
}

func (d *Document) QueryMany(symbolName []string) (*symbol.Symbol, bool) {
	return symbol.FindManySymbol(d.Symbols, symbolName)
}

func (d *Document) QueryAll(symbolName []string) []*symbol.Symbol {
	return symbol.FindManySymbols(d.Symbols, symbolName)
}

func (d *Document) QueryCalls(symbolName string) []*symbol.Symbol {
	return symbol.FindCalls(d.Calls, symbolName)
}

func (d *Document) QueryTargets() []*symbol.Symbol {
	return d.Targets
}

// QueryClosestTarget returns the target whose ByteStart is closest to (and not after) pos.
// Targets must be sorted by ByteStart, which extractTargets guarantees.
func (d *Document) QueryClosestTarget(pos uint) *symbol.Symbol {
	targets := d.Targets
	if len(targets) == 0 {
		return nil
	}

	// Find the first index where ByteStart > pos, then step back one.
	// So we get the closest from the start of the file
	idx, _ := slices.BinarySearchFunc(targets, pos, func(s *symbol.Symbol, p uint) int {
		if s.Position.ByteStart <= p {
			return -1
		}

		return 1
	})

	if idx == 0 {
		return nil
	}

	return targets[idx-1]
}

func (d *Document) RangeDocLoads(fn func(*Load)) {
	d.loadsMu.RLock()
	defer d.loadsMu.RUnlock()
	for _, doc := range d.DocLoads {
		fn(doc)
	}
}

func (d *Document) RangeIsLoadedBy(fn func(*Load)) {
	d.loadsMu.RLock()
	defer d.loadsMu.RUnlock()
	for _, doc := range d.IsLoadedBy {
		fn(doc)
	}
}

func (d *Document) resetLoads() {
	// Copy to remove from LoadedBy
	d.loadsMu.RLock()
	docs := make([]*Load, 0, len(d.DocLoads))
	for _, load := range d.DocLoads {
		docs = append(docs, load)
	}
	d.loadsMu.RUnlock()

	for _, load := range docs {
		doc := load.Doc
		doc.RemoveLoadedByDoc(d)
	}

	d.loadsMu.Lock()
	d.DocLoads = make(map[string]*Load)
	d.loadsMu.Unlock()
}

func (d *Document) AddLoadedDoc(doc *Document, loads []string) {
	// Locks both documents
	lockTwoLoads(&d.loadsMu, &doc.loadsMu)

	dl := &Load{Doc: doc, Loads: loads}
	d.DocLoads[dl.Hash()] = dl

	dlby := &Load{Doc: d, Loads: loads}
	doc.IsLoadedBy[dlby.Hash()] = dlby

	unlockTwoLoads(&d.loadsMu, &doc.loadsMu)
}

func (d *Document) RemoveLoadedDoc(doc *Document) {
	lockTwoLoads(&d.loadsMu, &doc.loadsMu)

	// remove from d.DocLoads
	for key, load := range d.DocLoads {
		if load.Doc == doc {
			delete(d.DocLoads, key)
			break
		}
	}

	// remove from doc.IsLoadedBy
	for key, load := range doc.IsLoadedBy {
		if load.Doc == d {
			delete(doc.IsLoadedBy, key)
			break
		}
	}

	unlockTwoLoads(&d.loadsMu, &doc.loadsMu)
}

func (d *Document) RemoveLoadedByDoc(doc *Document) {
	doc.loadsMu.Lock()

	for key, loader := range doc.IsLoadedBy {
		if loader.Doc == d {
			delete(doc.IsLoadedBy, key)
			break
		}
	}

	doc.loadsMu.Unlock()
}
