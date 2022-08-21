package pomme

import (
	"fmt"
)

/*
 *  Parse an expression that uses a register
 *  e.g. A = 1 or A = M$1234 or A = X + Y
 */
func (p *parser) parseRegister(token string) error {
	fmt.Printf("A/X/Y expressions are not yet supported [line %d]\n", p.n)
	return nil
}
