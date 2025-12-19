package runtime

// TODO: bsena; Push to state machine

type Machine struct {
	CurrState string

	HasSymbol bool

	IsFunction   bool
	IsIdentifier bool
}

type StateFn func(input string) StateFn

const (
	FunctionStart = "function_definition"
	FunctionName  = "name"
)

// TODO: bsena; rename only to Start() when withiun its own package
func NewMachine() *Machine {
	return &Machine{}
}

func (m *Machine) Start() StateFn {
	return m.start()
}

func (m *Machine) start() StateFn {
	return func(input string) StateFn {
		switch input {
		case FunctionStart:
			return m.functionStart()
		}

		return m.start()
	}
}

func (m *Machine) functionStart() StateFn {
	return func(input string) StateFn {
		return m.functionStart()
	}
}
