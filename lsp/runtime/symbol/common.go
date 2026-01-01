package symbol

import (
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

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

func ExtractCurrentSymbol(root *tree_sitter.Node, source []byte, byteOffSet uint) string {
	for root != nil && root.Kind() != "identifier" {
		root = root.FirstChildForByte(byteOffSet)
	}

	if root == nil {
		return ""
	}

	return root.Utf8Text(source)
}

var stopTokens = map[byte]bool{
	byte(' '):  true,
	byte('\t'): true,
	byte('\n'): true,
	byte(';'):  true,
	byte(':'):  true,
	byte('"'):  true,
	byte('('):  true,
	byte(')'):  true,
}

func ExtractCurrentLiteral(text []byte, start int) string {
	if start >= len(text) || start <= 0 || stopTokens[text[start]] {
		return ""
	}

	var startStop, endStop bool
	iStart, iEnd := start, start

	for {
		// Make iStart exclusive
		nextStart := iStart - 1
		if nextStart >= 0 && !startStop {
			currentStartToken := text[nextStart]
			if !stopTokens[currentStartToken] {
				iStart--
			} else {
				startStop = true
			}
		} else {
			startStop = true
		}

		if iEnd < len(text) && !endStop {
			currentEndToken := text[iEnd]
			if !stopTokens[currentEndToken] {
				iEnd++
			} else {
				endStop = true
			}
		} else {
			endStop = true
		}

		if startStop && endStop {
			break
		}

	}

	return string(text[iStart:iEnd])
}
