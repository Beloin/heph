package builtin

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/hephbuild/heph/internal/hlsp/runtime/query"
	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

//go:embed helpers.py
var helpers []byte

//go:embed pybt.py
var pybt []byte

const builtinSource = "hbuiltins"

var (
	helpersSymbols []*symbol.Symbol
	pybtSymbols    []*symbol.Symbol
	hephSymbols    []*symbol.Symbol

	// Primitive type symbols available in scope for type resolution
	primitiveSymbols []*symbol.Symbol
)

func init() {
	// Initialize primitive type symbols
	primitiveSymbols = []*symbol.Symbol{
		symbol.PrimitiveInt,
		symbol.PrimitiveFloat,
		symbol.PrimitiveBool,
		symbol.PrimitiveString,
		symbol.PrimitiveNull,
		symbol.ListType,
		symbol.DictType,
		symbol.ObjectType,
		symbol.UnknownType,
	}
}

// target is defined in-code as a standalone symbol.
var targetBt = symbol.Symbol{
	Name:      TargetName,
	Source:    builtinSource,
	Kind:      symbol.FunctionKind,
	Signature: "target(name:str, driver:str, labels:list, *args, **kwargs)",
	DocString: `Define a target for execution in the heph build system.

A target is defined by a name, and a driver. The driver specification dictates other args and commands.
This execution unit is isolated from the rest of the repo which allows for efficient caching and parallel execution.

Args:
    name (str): Target name (required)
    driver (str): The driver to execute this target (required)
    labels (list[str]): A list of labels to attach to this target
    *args (any): Arguments to be sent directly into the driver when executed.
    **kwargs (any): Keyword arguments to be passed directly to driver.`,
	Parameters: []*symbol.Parameter{
		{Name: "name", Type: symbol.PrimitiveString, DocString: "Target name (required)"},
		{Name: "driver", Type: symbol.PrimitiveString, DocString: "The driver to execute this target (required)"},
		{Name: "labels", Type: symbol.ListType, DocString: "A list of labels to attach to this target"},
		{Name: "*args", DocString: "Arguments to be sent directly into the driver when executed."},
		{Name: "**kwargs", DocString: "Keyword arguments to be passed directly to driver."},
	},
}

// InitBuiltins parses stub files and builds in-code builtins.
// Must be called once before using any getter.
func InitBuiltins(parser *tree_sitter.Parser) error {
	helpersTree := parser.Parse(helpers, nil)
	defer helpersTree.Close()

	pybtTree := parser.Parse(pybt, nil)
	defer pybtTree.Close()

	var err error

	helpersRoot, err := query.QueryAll(helpersTree, helpers, builtinSource)
	if err != nil {
		return err
	}
	helpersSymbols = query.FilterSymbolsByKind(helpersRoot.Symbols, symbol.FunctionKind)

	pybtRoot, err := query.QueryAll(pybtTree, pybt, builtinSource)
	if err != nil {
		return err
	}
	pybtSymbols = append(query.FilterSymbolsByKind(pybtRoot.Symbols, symbol.FunctionKind),
		query.FilterSymbolsByKind(pybtRoot.Symbols, symbol.VariableKind)...)

	hephSymbols = buildHephBuiltins()

	return nil
}

// GetTarget returns the standalone in-code target() symbol.
func GetTarget() *symbol.Symbol {
	return &targetBt
}

// GetHelpers returns the symbols parsed from helpers.py (load, group, text_file, …).
func GetHelpers() []*symbol.Symbol {
	return helpersSymbols
}

// GetSKBuiltins returns the symbols parsed from pybt.py (Starlark built-ins: abs, any, len, …).
func GetSKBuiltins() []*symbol.Symbol {
	return pybtSymbols
}

// GetHephBuiltins returns all available heph builtins.
func GetHephBuiltins() []*symbol.Symbol {
	return hephSymbols
}

// GetPrimitiveSymbols returns primitive type symbols for scope injection.
func GetPrimitiveSymbols() []*symbol.Symbol {
	return primitiveSymbols
}

// All returns all builtin symbols concatenated.
func All() []*symbol.Symbol {
	return slices.Concat([]*symbol.Symbol{GetTarget()}, helpersSymbols, pybtSymbols, hephSymbols, primitiveSymbols)
}

// buildHephBuiltins constructs all heph-namespace and utility builtin symbols.
// For dotted paths like "heph.pkg.dir", intermediate segments ("heph", "heph.pkg")
// are created with Kind=FieldKind and Type=ObjectType.
func buildHephBuiltins() []*symbol.Symbol {
	seen := map[string]bool{}
	var out []*symbol.Symbol

	add := func(syms []*symbol.Symbol) {
		for _, s := range syms {
			if !seen[s.Name] {
				seen[s.Name] = true
				out = append(out, s)
			}
		}
	}

	// TODO: bsena; Make a way so we can have description in each object

	// Top-level utilities
	add(makeFn("to_json", []*symbol.Parameter{
		{Name: "value", DocString: "Starlark object to serialize"},
	}, symbol.PrimitiveString, "Returns the string representation of a Starlark object.\n\nto_json(['hello']) # => [\"hello\"]"))

	// heph.*
	add(makePath("heph.canonicalize", []*symbol.Parameter{
		{Name: "target", Type: symbol.PrimitiveString, DocString: "Target address to canonicalize"},
	}, symbol.PrimitiveString, "Returns a canonicalized version of a target.\n\nheph.canonicalize(':test') # => //path/to:test"))

	add(makePath("heph.is_target", []*symbol.Parameter{
		{Name: "s", Type: symbol.PrimitiveString, DocString: "String to test"},
	}, symbol.PrimitiveBool, "Reports whether a string is a valid target address.\n\nheph.is_target(':test') # => True"))

	add(makePath("heph.split", []*symbol.Parameter{
		{Name: "target", Type: symbol.PrimitiveString, DocString: "Target address to split"},
	}, nil, "Splits a target address into (pkg, target, output).\n\npkg, target, output = heph.split('//some:addr')"))

	add(makePath("heph.param", []*symbol.Parameter{
		{Name: "name", Type: symbol.PrimitiveString, DocString: "Parameter name (set by -p)"},
	}, symbol.PrimitiveString, "Gets a build parameter set by -p.\n\nvalue = heph.param('test')"))

	// heph.pkg.*
	add(makePath("heph.pkg.dir", nil, symbol.PrimitiveString,
		"Gets the current package directory relative to the repo root.\n\nheph.pkg.dir() # => some/dir"))

	add(makePath("heph.pkg.name", nil, symbol.PrimitiveString,
		"Gets the current package name.\n\nheph.pkg.name() # => dir"))

	add(makePath("heph.pkg.addr", nil, symbol.PrimitiveString,
		"Gets the current package address.\n\nheph.pkg.addr() # => //some/dir"))

	return out
}

// makeFn creates a single function symbol with no dotted-path prefix.
func makeFn(name string, params []*symbol.Parameter, retType *symbol.Symbol, doc string) []*symbol.Symbol {
	sig := buildSig(name, params, retType)
	return []*symbol.Symbol{{
		Name:       name,
		Source:     builtinSource,
		Kind:       symbol.FunctionKind,
		Signature:  sig,
		Parameters: params,
		DocString:  doc,
	}}
}

// makePath creates symbols for a dotted path (e.g. "heph.pkg.dir").
// All segments before the last get Kind=FieldKind and Type=ObjectType.
// The final segment gets Kind=FunctionKind.
func makePath(dotted string, params []*symbol.Parameter, retType *symbol.Symbol, doc string) []*symbol.Symbol {
	parts := strings.Split(dotted, ".")
	var syms []*symbol.Symbol

	// Intermediate field segments
	for i := 1; i < len(parts); i++ {
		prefix := strings.Join(parts[:i], ".")
		syms = append(syms, &symbol.Symbol{
			Name:   prefix,
			Source: builtinSource,
			Kind:   symbol.FieldKind,
			Type:   symbol.ObjectType,
		})
	}

	// Final function symbol
	sig := buildSig(dotted, params, retType)
	syms = append(syms, &symbol.Symbol{
		Name:       dotted,
		Source:     builtinSource,
		Kind:       symbol.FunctionKind,
		Signature:  sig,
		Parameters: params,
		DocString:  doc,
	})

	return syms
}

// buildSig constructs a signature string: name(p1:type=val, …) -> retType
func buildSig(name string, params []*symbol.Parameter, retType *symbol.Symbol) string {
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString("(")
	for i, p := range params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(p.Name)
		if p.Type != nil {
			sb.WriteString(":" + p.Type.Name)
		}
		if p.Value != nil {
			sb.WriteString("=" + p.Value.Name)
		}
	}
	sb.WriteString(")")
	if retType != nil {
		sb.WriteString(fmt.Sprintf(" -> %s", retType.Name))
	}
	return sb.String()
}
