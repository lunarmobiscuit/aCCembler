package pomme

import (
	"fmt"
    "strings"
)


/*
 *  Lookup constant value
 */
func (p *parser) lookupConstant(name string) (int, error) {
	// Case insensitive
	name = strings.ToLower(name)

	// Iterate through all the constants
	for c := p.cnst; c != p.lastCnst; c = c.next {
		if (c.name == name) {
			return c.value, nil
		}
	}

	// Not found
	return 0, fmt.Errorf("const '%s' not defined", name)
}

/*
 *  Lookup symbol value as a subroutine block name
 */
func (p *parser) lookupSubroutineName(symbol string) *codeBlock {
	// Case insensitive
	symbol = strings.ToLower(symbol)

	// Try all the code blocks
	for b := p.code; b != nil; b = b.next {
		if b.name == symbol {
			return b
		}
	}

	// Not found
	return nil
}

/*
 *  Lookup symbol as a label in the code block
 */
func (b *codeBlock) lookupInstructionLabel(symbol string) (int, error) {
	// Look at every code entry
	for i := b.instr; i != nil; i = i.next {
		if (i.mnemonic == 0) && (i.symbol == symbol) {
			return i.address, nil
		}
	}

	// Not found
	return 0, fmt.Errorf("symbol '%s' not defined", symbol)
}

/*
 *  Lookup symbol value as a data block name
 */
func (p *parser) lookupDataName(symbol string) *dataBlock {
	// Case insensitive
	symbol = strings.ToLower(symbol)

	// Try all the data blocks
	for d := p.data; d != nil; d = d.next {
		if d.name == symbol {
			return d
		}
	}

	// Not found
	return nil
}

