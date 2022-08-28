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
	nameLC := strings.ToLower(name)

	// Iterate through all the constants
	for c := p.cnst; c != nil; c = c.next {
		if (c.nameLC == nameLC) {
			return c.value, nil
		}
	}

	// Not found
	return 0, fmt.Errorf("const '%s' not defined", name)
}

/*
 *  Lookup variable address
 */
func (p *parser) lookupVariable(b *codeBlock, name string) (int, int, error) {
	// Case insensitive
	nameLC := strings.ToLower(name)

	// Iterate through all the global variables
	for v := p.global; v != nil; v = v.next {
		if (v.nameLC == nameLC) {
			return v.address, v.size, nil
		}
	}

	// Iterate through the variables in the specified block
	if (b != nil) {
		for v := b.vrbl; v != nil; v = v.next {
			if (v.nameLC == nameLC) {
				return v.address, v.size, nil
			}
		}
	}

	// Not found
	return 0, 0, fmt.Errorf("variable '%s' not defined", name)
}

/*
 *  Lookup symbol value as a subroutine block name
 */
func (p *parser) lookupSubroutineName(symbol string) *codeBlock {
	// Case insensitive
	symbolLC := strings.ToLower(symbol)

	// Try all the code blocks
	for b := p.code; b != nil; b = b.next {
		if b.nameLC == symbolLC {
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
	// Case insensitive
	symbolLC := strings.ToLower(symbol)

	// Look at every code entry
	for i := b.instr; i != nil; i = i.next {
		if (i.mnemonic == 0) && (i.symbolLC == symbolLC) {
			return i.address, nil
		}
	}

	// If not found, try the parent block
	if (b.up != nil) {
		return b.up.lookupInstructionLabel(symbol)
	}

	// Not found
	return 0, fmt.Errorf("symbol '%s' not defined", symbol)
}

/*
 *  Lookup symbol value as a data block name
 */
func (p *parser) lookupDataName(name string) *dataBlock {
	// Case insensitive
	nameLC := strings.ToLower(name)

	// Try all the data blocks
	for d := p.data; d != nil; d = d.next {
		if d.nameLC == nameLC {
			return d
		}
	}

	// Not found
	return nil
}

