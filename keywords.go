package pomme

import (
	"fmt"
	"strings"
)

// Linked list of symbols
type keyword struct {
	name	string		// the "keyword"
	f		kFunc		// the func that parses the keyword
}
type kFunc func (*parser, string) error

var keywords = []keyword {
	{"var", parseLocalVariable},
	{"print", parsePrint},
	{"os", parseOs},
	{"if", parseIf},
	{"for", parseFor},
	{"while", parseWhile},
	{"var", parseVar},
}


/*
 *  Lookup the keyword
 */
func (p *parser) isKeyword(token string) bool {
	// Find the table entry for the keyword
	for k := range keywords {
		if (token == keywords[k].name) {
			return true
		}
	}

	return false
}

/*
 *  Lookup and parse the keyword
 */
func (p *parser) parseKeyword(token string) error {
	// Find the parser for the keyword
	for k := range keywords {
		if (token == keywords[k].name) {
			return keywords[k].f(p, token)
		}
	}

	return fmt.Errorf("keyword '%s' is invalid", token)
}

/*
 *  Parse the 'var' keyword
 */
func parseLocalVariable(p *parser, token string) error {
	p.skipWhitespace()
	return p.parseVariable(p.lastCode, p.nextAZ_az_09())
}

/*
 *  Parse the 'print' keyword
 */
func parsePrint(p *parser, token string) error {
	fmt.Printf("PRINT is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}

/*
 *  Parse the 'os' keyword
 */
func parseOs(p *parser, token string) error {
	fmt.Printf("IF is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}

/*
 *  Parse the 'if' keyword
 */
func parseIf(p *parser, token string) error {
	fmt.Printf("IF is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}

/*
 *  Parse the 'for' keyword
 */
func parseFor(p *parser, token string) error {
	ifloop := new(ifloop)
	ifloop.startAddr = p.lastCode.endAddr
	ifloop.endAddr = ifloop.startAddr

	v := p.nextAZ_az_09()
	if (v == "") {
		return fmt.Errorf("missing loop variable in FOR")
	}

	p.skipWhitespace()
	if (p.nextChar() != '=') {
		return fmt.Errorf("missing = in FOR")
	}

	start, err := p.nextValue()
	if (err != nil) {
		return fmt.Errorf("invalid starting value in FOR")
	}

	to := strings.ToUpper(p.nextAZ_az_09())
	if (to == "DOWN") {
		ifloop.upDown = false
		to = strings.ToUpper(p.nextAZ_az_09())
	} else {
		ifloop.upDown = true
	}

	if (to != "TO") {
		return fmt.Errorf("syntax error in FOR, missing TO")
	}

	end, err2 := p.nextValue()
	if (err2 != nil) {
		return fmt.Errorf("invalid ending value in FOR")
	}

	p.skipWhitespace()
	if (p.nextChar() != '{') {
		return fmt.Errorf("missing { in FOR")
	}


fmt.Printf("FOR %s = %d TO %d\n", v, start, end)

	return nil
}

/*
 *  Parse the 'while' keyword
 */
func parseWhile(p *parser, token string) error {
	fmt.Printf("WHILE is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}

/*
 *  Parse the 'var' keyword
 */
func parseVar(p *parser, token string) error {
	fmt.Printf("VAR is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}
