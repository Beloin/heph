package symbol

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

func ExtractCurrentSymbol(text []byte, start int) string {
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
