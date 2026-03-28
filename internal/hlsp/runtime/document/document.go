package document

import (
	"context"
	"errors"
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

	// Root symbol defining all symbol scope tree
	Root    *symbol.Symbol
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

	// TODO: bsena; this is a stop-the-world mutex, rename it
	treeMutex sync.Mutex
	loadsMu   sync.RWMutex

	drivers *driver.Registry
}

var ErrParseFailed = errors.New("parse failed")

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

	syms, calls, targets, err := extractSymbols(doc.Tree, doc.Text, doc.FullPath)
	doc.Root = &symbol.Symbol{Name: name, Symbols: syms, Kind: symbol.RootKind}
	doc.Calls = calls
	doc.Targets = targets
	doc.extractLoads()
	// go doc.extractTargets()
	go doc.extractTargets2()

	return doc, err
}

func (d *Document) SwapTree(parser *tree_sitter.Parser, newText []byte) (*tree_sitter.Tree, error) {
	d.treeMutex.Lock()
	defer d.treeMutex.Unlock()

	newT := parser.Parse(newText, nil)
	if newT == nil {
		return nil, ErrParseFailed
	}

	oldTree := d.Tree

	syms, calls, targets, err := extractSymbols(newT, newText, d.FullPath)
	if err != nil {
		return nil, err
	}

	d.Root = &symbol.Symbol{Name: d.FullPath, Symbols: syms, Kind: symbol.RootKind}
	d.Calls = calls
	d.Targets = targets
	d.Tree = newT
	d.Text = newText
	d.TextString = string(newText)

	d.resetLoads()
	d.extractLoads()
	// go d.extractTargets()
	go d.extractTargets2()

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
			defaultTarget.Position = call.Position
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

// extractTargets2 will add information to targets recursively
func (d *Document) extractTargets2() {
	d.treeMutex.Lock()
	defer d.treeMutex.Unlock()

	btTarget := builtin.GetTarget()[0]

	d.enrichTargets(d.Root.Symbols, btTarget)
}

func (d *Document) enrichTargets(syms []*symbol.Symbol, btTarget *symbol.Symbol) {
	for _, s := range syms {
		if s.Kind == symbol.FunctionKind {
			d.enrichTargets(s.Symbols, btTarget)
			continue
		}

		if s.Name != builtin.TargetName {
			continue
		}

		s.Kind = symbol.TargetCallKind
		s.DocString = btTarget.DocString

		var driverName string
		for _, p := range s.Parameters {
			if p.Name == builtin.DriverName {
				driverName = strings.Trim(p.Value, "\"'")
				break
			}
		}

		if driverName == "" {
			s.DocString = btTarget.DocString

			continue
		}

		if schema, found := d.getDriverSchema(driverName); found {
			s.Source = schema.Source
			s.Signature = schema.Signature
			s.Parameters = schema.Parameters
			s.DocString = schema.DocString
		} else {
			s.Signature += " // driver `" + driverName + "` not found"
		}
	}
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

func extractSymbols(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, []*symbol.Symbol, []*symbol.Symbol, error) {
	// TODO: bsena; Someone is updating tree/text withou waiting for this query
	result, err := query.QueryAll(tree, text, source)
	if err != nil {
		return nil, nil, nil, err
	}

	// TODO: bsena; add all inside all symbols
	symbols := append(result.Functions, result.Variables...)
	symbols = append(symbols, result.Calls...)

	// Target calls are special calls
	// TODO: bsena; Maybe do this to groups also?
	// Nah, we will remove this
	var calls, targets []*symbol.Symbol
	// for _, call := range result.Calls {
	// 	if call.Name == builtin.TargetName {
	// 		// call.Kind = symbol.TargetCallKind
	// 		calls = append(targets, call)
	// 	} else {
	// 		calls = append(calls, call)
	// 	}
	// }

	return symbols, calls, targets, nil
}

func (d *Document) extractLoads() {
	loads := []*RawLoad{}
	for _, call := range d.Root.Symbols {
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

// ExtractCurrentSymbolName extracts current idenfier name closest to current cursor position
func (d *Document) ExtractCurrentSymbolName(byteOffSet uint) string {
	return query.ExtractCurrentSymbol(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) WhereAmI(byteOffSet uint) query.CodeLocation {
	return query.WhereAmI(d.Tree.RootNode(), d.Text, byteOffSet)
}

func (d *Document) SymbolHierarchy(byteOffSet uint) []*symbol.Symbol {
	var all []*symbol.Symbol
	all = append(all, d.Root.Symbols...)
	d.RangeDocLoads(func(l *Load) {
		all = append(all, l.Doc.Root.Symbols...)
	})

	res := query.SymbolHierarchy(d.Tree.RootNode(), d.Text, byteOffSet, all)
	res = append(res, builtin.GetHelpers()...)
	res = append(res, builtin.GetHephBuiltins()...)
	res = append(res, builtin.GetSKBuiltins()...)

	return res
}

func (d *Document) SymbolHierarchyWithLocation(byteOffSet uint) (query.CodeLocation, string, []*symbol.Symbol) {
	var all []*symbol.Symbol
	all = append(all, d.Root)
	all = append(all, d.Root.Symbols...)
	d.RangeDocLoads(func(l *Load) {
		all = append(all, l.Doc.Root.Symbols...)
	})

	loc, symbolName, hierarchy := query.SymbolHierarchyWithLocation(d.Tree.RootNode(), d.Text, byteOffSet, all)
	// TODO: fix this size
	newRes := make([]*symbol.Symbol, 0, len(hierarchy)+len(hierarchy))
	newRes = append(newRes, builtin.GetHelpers()...)
	newRes = append(newRes, builtin.GetHephBuiltins()...)
	newRes = append(newRes, builtin.GetSKBuiltins()...)
	newRes = append(newRes, hierarchy...)

	return loc, symbolName, newRes
}

func (d *Document) ExtractCurrentFunctionName(byteOffSet uint) string {
	return query.ExtractFunctionNameFromOffset(d.Tree.RootNode(), d.Text, byteOffSet)
}

// QueryType resolves a type name to its defining symbol.
// Only PrimitiveKind and ClassKind symbols are returned to avoid confusing
// variables or functions that happen to share a type name.
func (d *Document) QueryType(name string) (*symbol.Symbol, bool) {
	if s := query.ResolveType(name); s != nil {
		return s, true
	}

	if s, found := symbol.FindSymbol(d.Root.Symbols, name); found {
		if s.Is(symbol.PrimitiveKind) || s.Is(symbol.ClassKind) {
			return s, true
		}
	}

	d.loadsMu.RLock()
	defer d.loadsMu.RUnlock()
	for _, load := range d.DocLoads {
		if s, found := symbol.FindSymbol(load.Doc.Root.Symbols, name); found {
			if s.Is(symbol.PrimitiveKind) || s.Is(symbol.ClassKind) {
				return s, true
			}
		}
	}

	return nil, false
}

func (d *Document) Query(symbolName string) (*symbol.Symbol, bool) {
	for _, symbol := range d.Root.Symbols {
		if symbol.FullyQualifiedName == symbolName {
			return symbol, true
		}
	}

	return nil, false
}

func (d *Document) QueryMany(symbolName []string) (*symbol.Symbol, bool) {
	return symbol.FindManySymbol(d.Root.Symbols, symbolName)
}

func (d *Document) QueryAll(symbolName []string) []*symbol.Symbol {
	return symbol.FindManySymbols(d.Root.Symbols, symbolName)
}

// TODO: bsena; All these Queries must be scope-aware
// Looking only for those if are not replaced in the scope
func (d *Document) QueryCalls(symbolName string) []*symbol.Symbol {
	return symbol.FindCalls(d.Root.Symbols, symbolName)
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
