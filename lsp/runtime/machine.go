package runtime

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

// TODO: bsena; Push to state machine

const (
	StateIdle               = "idle"
	StateFunction           = "function definition"
	StateFunctionIdentifier = "function definition"

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
	IdentifierToken = "identifier"

	BlockToken = "block"

	TypeToken = "type"

	// Token types
	StringTypeToken    = "string"
	StringStartToken   = "string_start"
	StringContentToken = "string_content"
	StringStopToken    = "string_stop"

	IntTypeToken   = "int"
	FloatTypeToken = "float"

	// Literal Tokens
	EqualToken              = "="
	LiteralOpenParenthesis  = "("
	LiteralCloseParenthesis = ")"
	LiteralColon            = ":"
	LiteralArrowType        = "->"
	LiteralComma            = ","
)

// Machine State Machine-like implementation.
// To make it concurrent use a channel to communicate between functions.
type Machine struct {
	state string
	text  []byte

	// HasSymbol flag for whether we have a symbol in the current context. Can only be assigned when state is finalized
	HasSymbol bool

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
// For now it needs to have an end condition and can only change its own state.
type StateFn func(*tree_sitter.Node) StateFn

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
		// TODO: bsena; Add here parameters and return type definitions
		kind := node.Kind()
		switch kind {
		case DefToken:
			return m.functionStart()
		case IdentifierToken:
			m.SymbolName = m.extractCurrentByteRange(node)
			m.TSSymbolKind = kind
			m.IsFunction = true
			return m.functionIdentifier()
		}

		return m.start()
	}
}

func (m *Machine) functionIdentifier() StateFn {
	return func(node *tree_sitter.Node) StateFn {
		kind := node.Kind()
		switch kind {
		// TODO: bsena; parse parameters
		case ParametersToken, LiteralOpenParenthesis, TypedParameterToken,
			IdentifierToken, LiteralColon, TypeToken,
			LiteralComma, LiteralCloseParenthesis, LiteralArrowType:
			return m.functionIdentifier()
		case BlockToken:
			return m.functionBlockStart()
		}

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
			// return m.functionStatmentStart()
		}

		// TODO: bsena; When fuction stops give tree_sitter.Node a clue to go back up in the tree to prevent parsing undesirable nodes
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
		return m.start()
	}
}

func (m *Machine) extractCurrentByteRange(node *tree_sitter.Node) string {
	start, end := node.ByteRange()
	v := m.text[start:end]
	return string(v)
}
