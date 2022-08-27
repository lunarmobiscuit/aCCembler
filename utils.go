package pomme

import (
	"errors"
	"fmt"
)

const TAB = 9
const LF = 10
const CR = 13

/*
 *  Skip one or more characters
 *  (returning the updated index)
 */
func (p *parser) skip(k int) {
	// skip k characters
	for (p.i < p.end) && (k > 0) {
		p.i += 1
		k -= 1;
	}
}

/*
 *  Skip any whitespace
 *  (returning the updated index)
 */
func (p *parser) skipWhitespace() {
	// skip whitespace
	for p.i < p.end {
		if (p.b[p.i] == CR) || (p.b[p.i] == LF) { // stop at end of a line, and leave i pointing to the CR/LF
			break
		} else if p.b[p.i] <= ' ' { // skip ' ' and all other control characters
			p.i += 1
		} else {
			break
		}
	}
}

/*
 *  Skip any whitespace including CR/LF
 *  (returning the updated index)
 */
func (p *parser) skipWhitespaceAndEOL() {
	// skip whitespace and CR/LF
	for p.i < p.end {
		if (p.b[p.i] == CR) || (p.b[p.i] == LF) { // skip CR/LF too
			p.i += 1
			p.n += 1
		} else if p.b[p.i] <= ' ' { // skip ' ' and all other control characters
			p.i += 1
		} else {
			break
		}
	}
}

/*
 *  Skip to a specific character
 *  (returning the updated index)
 */
func (p *parser) skipTo(c uint8) {
	// skip to the specified character
	for p.i < p.end {
		if (p.b[p.i] != c) || (p.b[p.i] == CR) || (p.b[p.i] == LF) {
			break
		} else {
			p.i += 1
		}
	}
}

/*
 *  Skip any comments or blank lines
 */
func (p *parser) skipComment() bool {
	sym1 := p.peekChar()
	sym2 := p.peekAhead(1)
	if (sym1 == ';') {	// assembly-style comment
		p.nextLine()
		p.skipWhitespaceAndEOL()
		return true
	} else if (sym1 == '/') && (sym2 == '/') {	// C-style 1-line
		p.nextLine()
		p.skipWhitespaceAndEOL()
		return true
	} else if (sym1 == '/') && (sym2 == '*') {	// C-style multi-line comment
		// skip to '*/'
		p.i += 2
		for p.i < p.end {
			// skipping to a new line
			if p.b[p.i] == LF {
				p.n += 1
			}

			// searching for the '*/'
			if (p.b[p.i] == '*') && (p.peekAhead(1) == '/') {
				p.i += 2
				break
			} else {
				p.i += 1
			}
		}

		p.skipWhitespaceAndEOL()
		return true
	} else if (sym1 == CR) || (sym1 == LF) {	// blank line
		p.skipWhitespaceAndEOL()
		return true
	}

	return false
}

/*
 *  Skip to the end of the line
 *  (returning the updated index)
 */
func (p *parser) nextLine() {
	// skip to the CR/LF
	for p.i < p.end {
		if (p.b[p.i] == CR) || (p.b[p.i] == LF) {
			p.i += 1
			// check for CR+LF and skip that too
			if (p.i < p.end) && ((p.b[p.i] == CR) || (p.b[p.i] == LF)) {
				p.i += 1
				p.n += 1
			}
			break
		} else {
			p.i += 1
		}
	}

	p.n += 1
}

/*
 *  Does the next token start with A-Za-z
 *  (returning boolean)
 */
func (p *parser) isNextAZ() bool {
	// skip past whitespace
	j := p.i
	for p.b[j] <= ' ' {
		j += 1
	}

	// is the next character AZaz?
	if ((p.b[j] >= 'A') && (p.b[j] <= 'Z')) ||
			(p.b[j] >= 'a') && (p.b[j] <= 'z') || (p.b[j] == '_') {
		return true
	} else {
		return false
	}
}

/*
 *  Return the next A-Za-z0-9_ token
 *  (returning the string and index)
 */
func (p *parser) nextAZ_az_09() string {
	// skip past whitespace
	p.skipWhitespace()

	// must start with AZ character
	j := p.i
	if ((p.b[p.i] >= 'A') && (p.b[p.i] <= 'Z')) ||
			(p.b[p.i] >= 'a') && (p.b[p.i] <= 'z') || (p.b[p.i] == '_') {
		p.i += 1
	} else {
		return ""
	}

	// find the next non AZ09 character
	for p.i < p.end {
		if ((p.b[p.i] >= 'A') && (p.b[p.i] <= 'Z')) ||
				((p.b[p.i] >= 'a') && (p.b[p.i] <= 'z')) ||
					((p.b[p.i] >= '0') && (p.b[p.i] <= '9')) || (p.b[p.i] == '_') {
			p.i += 1
		} else {
			break
		}
	}

	return string(p.b[j:p.i])
}

/*
 *  Return the next number (w/ or w/out a $ or 0x prefix)
 *  (returning the string and index)
 */
func (p *parser) nextValue() (int, error) {
	// skip past whitespace
	p.skipWhitespace()

	// Hexidecimal or decimal?
	if p.b[p.i] == '$' {
		p.skip(1)
		return p.nextHexidecimal(), nil
	} else if (p.i+1 < p.end) && (p.b[p.i] == '0') && (p.b[p.i+1] == 'x') {
		p.skip(2)
		return p.nextHexidecimal(), nil
	} else {
		return p.nextDecimal(), nil
	}

	return 0, errors.New("the expected value is not a number")
}

/*
 *  Parse the next 0-9A-Fa-f characdters as a hexideimal value
 *  (returning the value and index)
 */
func (p *parser) nextHexidecimal() int {
	// find the next non AZ09 character
	value := 0
	for p.i < p.end {
		if (p.b[p.i] >= '0') && (p.b[p.i] <= '9') {
			value = (value * 16) + int(p.b[p.i] - '0')
			p.i += 1
		} else if (p.b[p.i] >= 'A') && (p.b[p.i] <= 'F') {
			value = (value * 16) + int(p.b[p.i] - 'A' + 10)
			p.i += 1
		} else if (p.b[p.i] >= 'a') && (p.b[p.i] <= 'f') {
			value = (value * 16) + int(p.b[p.i] - 'a' + 10)
			p.i += 1
		} else {
			break
		}
	}

	return value
}

/*
 *  Parse the next 0-9 characters as a hexideimal value
 *  (returning the value and index)
 */
func (p *parser) nextDecimal() int {
	// find the next non AZ09 character
	value := 0
	for p.i < p.end {
		if (p.b[p.i] >= '0') && (p.b[p.i] <= '9') {
			value = (value * 10) + int(p.b[p.i] - '0')
			p.i += 1
		} else {
			break
		}
	}

	return value
}

/*
 *  Return the next characters until a quote is found
 *  (returning the string and index)
 */
func (p *parser) untilQuote() string {
	// look for the next character that isn't a quote
	j := p.i
	for p.i < p.end {
		if p.b[p.i] == '"' {
			return string(p.b[j:p.i])
		}
		p.i += 1
	}

	return ""
}

/*
 *  Are the next bytes the start of a comment?
 *  (returning true/false)
 */
func (p *parser) isComment() bool {
	sym1 := p.peekChar()
	sym2 := p.peekAhead(1)
	if (sym1 == ';') {	// assembly-style comment
		return true
	} else if (sym1 == '/') && (sym2 == '/') {	// C-style 1-line
		return true
	} else if (sym1 == '/') && (sym2 == '*') {	// C-style multi-line comment
		return true
	}

	return false
}

/*
 *  Return the next character
 *  (returning the char but not updating the index)
 */
func (p *parser) peekChar() uint8 {
	if p.i > p.end {
		return 0
	}
	return p.b[p.i]
}

/*
 *  Return the nth next character
 *  (returning the char but not updating the index)
 */
func (p *parser) peekAhead(k int) uint8 {
	if p.i+k > p.end {
		return 0
	}
	return p.b[p.i+k]
}

/*
 *  Determine the correct prefix for the size of the address or value
 */
func addressToPrefix(v int) int {
	if (v <= 0x0FFFF) {
		return A16
	} else if (v <= 0x0FFFFFF) {
		return A24
	} else {
		return A32
	}
}
func valueToPrefix(v int) int {
	if (v <= 0x0FF) {
		return R08
	} else if (v <= 0x0FFFF) {
		return R16
	} else if (v <= 0x0FFFFFF) {
		return R24
	} else {
		return R32
	}
}

/*
 *  Turn the size into a string
 */
func sizeToSuffix(sz int) string {
	suffix := ""
	if (sz != A16) {
		if (sz == R16) || (sz == W16) {
			suffix += ".w"
		} else if (sz == R24) || (sz == W24) {
			suffix += ".t"
		}
		if (sz & A24) == A24 {
			suffix += ".a24"
		}
	}

	return suffix
}


// Stringer
func (p *parser) String() string {
	return fmt.Sprintf("b[%d] i%d n%d\n", p.end, p.i, p.n)
}
