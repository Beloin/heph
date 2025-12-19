package runtime

// TODO: bsena; Push to state machine

type Machine struct {
	state     string
	HasSymbol bool

	IsFunction bool
	IsVariable bool
}

type StateFn func(input string) StateFn

const (
	StateIdle     = "idle"
	StateFunction = "function definition"
)

const (
	FunctionToken   = "function_definition"
	IdentifierToken = "identifier"
)

// TODO: bsena; rename only to Start() when withiun its own package
func NewMachine() *Machine {
	return &Machine{}
}

func (m *Machine) Start() StateFn {
	return m.start()
}

func (m *Machine) reset(state string) {
	m.state = state
	m.HasSymbol = false
}

func (m *Machine) start() StateFn {
	return func(input string) StateFn {
		m.reset(StateIdle)
		switch input {
		case FunctionToken:
			return m.functionStart()
		}

		return m.start()
	}
}

func (m *Machine) functionStart() StateFn {
	return func(input string) StateFn {
		m.reset(StateFunction)
		switch input {
		case IdentifierToken:
			m.HasSymbol = true
			return m.start()
		}

		return m.functionStart()
	}
}
