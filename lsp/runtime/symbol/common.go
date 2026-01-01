package symbol

// TODO: bsena; change to iterative instead of recursive
func FindSymbol(symbols []*Symbol, sName string) (*Symbol, bool) {
	for _, symbol := range symbols {
		if symbol.FullyQualifiedName == sName {
			return symbol, true
		}

		if childS, found := FindSymbol(symbol.Symbols, sName); found {
			return childS, true
		}
	}

	return nil, false
}
