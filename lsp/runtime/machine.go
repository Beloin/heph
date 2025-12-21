package runtime

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

// TODO: bsena; Push to state machine

const (
	// Function tokens
	FunctionToken              = "function_definition"
	FunctionCallToken          = "function"
	DefToken                   = "def"
	ParametersToken            = "parameters"
	ListSplatToken             = "list_splat_pattern"
	DefaultParameterToken      = "default_parameter"
	TypeParamaterToken         = "type_parameter"
	TypedParameterToken        = "typed_parameter"
	TypedDefaultParameterToken = "typed_default_parameter"

	GenericTypeToken = "generic_type"

	StatementToken  = "expression_statement"
	AssignmentToken = "assignment"
	IdentifierToken = "identifier"

	BlockToken = "block"

	TypeToken = "type"

	// Token types
	StringTypeToken    = "string"
	StringStartToken   = "string_start"
	StringContentToken = "string_content"
	StringEndToken     = "string_end"

	IntTypeToken   = "integer"
	FloatTypeToken = "float"
	ListTypeToken  = "list"
	DictTypeToken  = "dictionary"
	NoneTypeToken  = "none"

	// Literal Tokens
	LiteralEqualToken            = "="
	LiteralOpenParenthesisToken  = "("
	LiteralCloseParenthesisToken = ")"
	LiteralOpenSBracketToken     = "["
	LiteralCloseSBracketToken    = "]"
	LiteralOpenCBracketToken     = "{"
	LiteralCloseCBracketToken    = "}"
	LiteralColonToken            = ":"
	LiteralArrowTypeToke         = "->"
	LiteralCommaToken            = ","
	LiteralSplatToken            = "*"

	LiteralFalseToken = "false"
	LiteralTrueToken  = "true"
)

// Machine State Machine-like implementation.
// To make it concurrent use a channel to communicate between functions.
type Machine struct {
	text []byte

	// HasSymbol flag for whether we have a symbol in the current context. Can only be assigned when state is finalized.
	HasSymbol bool

	Symbol Symbol
}

// StateFn is Rob Pike's state machine pattern. For every state there is a function that represents it.
// For now it needs to have an end condition and can only change its own state.
type StateFn func(*tree_sitter.Node) StateFn

// TODO: bsena; rename only to Start() when withiun its own package
func NewMachine(text []byte) *Machine {
	return &Machine{text: text}
}

func (m *Machine) Start() StateFn {
	return m.start()
}

func (m *Machine) reset() {
	m.HasSymbol = false
	m.Symbol = Symbol{}
}

// TODO: bsena; What if instead we go here, we just query the tree?
// like get symbols from tree.Query(...)
// The idea is to have a tree+rawText+pre-fetched symbols 
// and keep changing tree+rawText+symbols when documented has been updated

func (m *Machine) start() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		m.reset()

		input := node.Kind()

		// TODO: bsena; parse classes and methods to have inner
		switch input {
		case FunctionToken:
			return m.functionStart()
		case StatementToken:
			return m.statementStart()
		}

		return m.start()
	}
}

func (m *Machine) functionStart() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		// TODO: bsena; Add here parameters and return type definitions

		kind := node.Kind()
		switch kind {
		case DefToken:
			return m.functionStart()
		case IdentifierToken:
			// Setup start
			m.Symbol.Position.RowStart = node.Range().StartPoint.Row
			m.Symbol.Position.ColumnStart = node.Range().StartPoint.Column

			m.Symbol.Name = m.extractCurrentByteRange(node)
			m.Symbol.Kind = FunctionKind

			m.Symbol.signaturePosition.ByteStart = node.Range().StartByte
			return m.functionIdentifier()
		}

		return m.start()
	}
}

func (m *Machine) functionIdentifier() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		// TODO: bsena; parse parameters, probably put as children, or use submachines
		kind := node.Kind()
		switch kind {
		case LiteralColonToken:
			// TODO: make a better way to do this
			m.Symbol.signaturePosition.ByteEnd = node.Range().StartByte
			m.Symbol.Signature = m.extractByteRange(m.Symbol.signaturePosition.ByteStart, m.Symbol.signaturePosition.ByteEnd)
			fallthrough
		case ParametersToken, DefaultParameterToken, TypedParameterToken, TypedDefaultParameterToken, TypeParamaterToken, ListSplatToken,

			LiteralOpenParenthesisToken, LiteralCloseParenthesisToken,
			LiteralOpenSBracketToken, LiteralCloseSBracketToken,
			LiteralOpenCBracketToken, LiteralCloseCBracketToken,
			LiteralSplatToken,

			GenericTypeToken,
			IdentifierToken,

			LiteralFalseToken, LiteralTrueToken,

			StringTypeToken, StringStartToken, StringContentToken, StringEndToken,
			IntTypeToken, FloatTypeToken, ListTypeToken, NoneTypeToken, DictTypeToken,

			TypeToken, LiteralEqualToken,
			LiteralCommaToken, LiteralArrowTypeToke:
			return m.functionIdentifier()
		case BlockToken:
			return m.functionBlockStart()
		}

		m.Symbol.Position.RowEnd = node.Range().EndPoint.Row
		m.Symbol.Position.ColumnEnd = node.Range().EndPoint.Column
		m.HasSymbol = true
		return m.start()
	}
}

// functionBlockStart defines the start block of a function
// Only parse if its function documentation, if it's any other other statement stop parsing
func (m *Machine) functionBlockStart() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		kind := node.Kind()
		switch kind {
		case StatementToken:
			return m.typeState()
		}

		// TODO: bsena; When fuction stops give tree_sitter.Node a clue to go back up in the tree to prevent parsing undesirable nodes
		m.Symbol.Position.RowEnd = node.Range().EndPoint.Row
		m.Symbol.Position.ColumnEnd = node.Range().EndPoint.Column
		m.HasSymbol = true
		return m.start()
	}
}

// functionStatmentStart defines the first statement of a function
// Only parse if its function documentation, if it's any other other statement stop parsing
func (m *Machine) functionStatmentStart() StateFn {
	// TODO: bsena; Create a new inner state machine to identify nested blocks and stop when they have been processed?
	return func(node *tree_sitter.Node) StateFn {
		kind := node.Kind()
		switch kind {
		case StringTypeToken:
			return m.stringTypeState()
		}

		m.HasSymbol = true
		return m.start()
	}
}

func (m *Machine) statementStart() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		input := node.Kind()
		switch input {
		case AssignmentToken:
			return m.assignmentState()
		}

		return m.start()
	}
}

func (m *Machine) assignmentState() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		kind := node.Kind()
		switch kind {
		case IdentifierToken:
			// Setup start
			m.Symbol.Position.RowStart = node.Range().StartPoint.Row
			m.Symbol.Position.ColumnStart = node.Range().StartPoint.Column

			m.Symbol.Name = m.extractCurrentByteRange(node)
			m.Symbol.Kind = VariableKind

			return m.typeState()
		}

		return m.start()
	}
}

func (m *Machine) typeState() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		kind := node.Kind()
		switch kind {
		case LiteralEqualToken:
			return m.typeState()
		case StringTypeToken:
			return m.stringTypeState()
		case IntTypeToken:
			fallthrough
		case FloatTypeToken:
		}

		m.Symbol.Position.RowEnd = node.Range().EndPoint.Row
		m.Symbol.Position.ColumnEnd = node.Range().EndPoint.Column
		m.HasSymbol = true
		return m.start()
	}
}

func (m *Machine) stringTypeState() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		kind := node.Kind()
		switch kind {
		case StringTypeToken, StringStartToken:
			return m.stringTypeState()
		case StringContentToken:
			v := m.extractCurrentByteRange(node)
			m.Symbol.Value = v
		}

		m.Symbol.Position.RowEnd = node.Range().EndPoint.Row
		m.Symbol.Position.ColumnEnd = node.Range().EndPoint.Column
		m.HasSymbol = true
		return m.start()
	}
}

func (m *Machine) extractCurrentByteRange(node *tree_sitter.Node) string {
	start, end := node.ByteRange()
	return m.extractByteRange(start, end)
}

func (m *Machine) extractByteRange(start, end uint) string {
	v := m.text[start:end]
	return string(v)
}
