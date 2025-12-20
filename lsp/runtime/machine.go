package runtime

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

// TODO: bsena; Push to state machine

// Machine State Machine-like implementation.
// To make it concurrent use a channel to communicate between functions.
type Machine struct {
	state string
	text  []byte

	// HasSymbol flag for whether we have a symbol in the current context. Can only be assigned when state is finalized
	HasSymbol  bool
	SymbolKind int

	// TODO: something tell me that these control variables are the opposite of what a state machine should be
	IsFunction bool
	IsVariable bool
	HasValue   bool

	TSSymbolKind string
	SymbolName   string
	SymbolValue  string
	SymbolDoc    string
}

// StateFn is Rob Pike's state machine pattern. For every state there is a function that represents it.
type StateFn func(*tree_sitter.Node) StateFn

const (
	StateIdle       = "idle"
	StateFunction   = "function definition"
	StateAssigment  = "Assignment"
	StateStatement  = "Statement"
	StateType       = "Type"
	StateTypeString = "TypeString"
)

const (
	// Function tokens
	FunctionToken       = "function_definition"
	DefToken            = "def"
	ParametersToken     = "parameters"
	TypedParameterToken = "typed_parameter"

	StatementToken  = "expression_statement"
	AssignmentToken = "assignment"

	// Variable token
	IdentifierToken = "identifier"

	TypeToken = "type"

	EqualToken = "="

	BlockToken = "block"

	// Token types
	StringTypeToken    = "string"
	StringStartToken   = "string_start"
	StringContentToken = "string_content"
	StringStopToken    = "string_stop"

	IntTypeToken   = "int"
	FloatTypeToken = "float"

	LiteralOpenParenthesis  = "("
	LiteralCloseParenthesis = ")"
	LiteralColon            = ":"
	LiteralArrowType        = "->"
)

// TODO: bsena; rename only to Start() when withiun its own package
func NewMachine(text []byte) *Machine {
	return &Machine{text: text}
}

func (m *Machine) Start() StateFn {
	return m.start()
}

func (m *Machine) reset(state string) {
	m.state = state

	m.HasSymbol = false
	m.IsVariable = false
	m.IsFunction = false
	m.HasValue = false

	m.SymbolValue = ""
}

// TODO: bsena; maybe we actually need something like "last know type" so IsComplete can be removed

func (m *Machine) start() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		m.reset(StateIdle)
		input := node.Kind()

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
		m.reset(StateFunction)

		// TODO: bsena; Add here parameters and return type definitions
		kind := node.Kind()
		switch kind {
		case DefToken:
			return m.functionStart()
		case IdentifierToken:
			m.SymbolName = m.extractCurrentByteRange(node)
			m.TSSymbolKind = kind
			m.HasSymbol = true
			m.IsFunction = true
			return m.start()
		}

		return m.start()
	}
}

func (m *Machine) statementStart() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		m.reset(StateStatement)

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
		m.reset(StateAssigment)

		kind := node.Kind()
		switch kind {
		case IdentifierToken:
			m.SymbolName = m.extractCurrentByteRange(node)
			m.IsVariable = true
			m.TSSymbolKind = kind
			return m.typeState()
		}

		return m.start()
	}
}

func (m *Machine) typeState() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		m.reset(StateType)

		kind := node.Kind()
		switch kind {
		case EqualToken:
			return m.typeState()
		case StringTypeToken:
			return m.stringTypeState()
		case IntTypeToken:
			fallthrough
		case FloatTypeToken:
		}

		m.HasSymbol = true
		return m.start()
	}
}

func (m *Machine) stringTypeState() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		m.reset(StateTypeString)

		kind := node.Kind()
		switch kind {
		case StringTypeToken, StringStartToken:
			return m.stringTypeState()
		case StringContentToken:
			v := m.extractCurrentByteRange(node)
			m.SymbolValue = v
			m.HasValue = true
		}

		m.HasSymbol = true
		m.IsVariable = true
		return m.start()
	}
}

func (m *Machine) extractCurrentByteRange(node *tree_sitter.Node) string {
	start, end := node.ByteRange()
	v := m.text[start:end]
	return string(v)
}
