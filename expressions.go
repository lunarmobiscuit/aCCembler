package aCCembler

import (
	"fmt"
	"strings"
)

const (
	MEMORY = iota
	VARIABLE
	VALUE
	REG_A
	REG_X
	REG_Y
	N_A
)

const (
	NO_OP = iota
	EQUALS
	PLUS
	MINUS
	AND
	OR
	EOR
	SHIFT_LEFT
	SHIFT_RIGHT
)

/*
 *  Parse an expression that uses a register, can be U op= 'V', or U op= 'V op W'
 *  e.g. M$1234 = 123 or A = 1 + 2 or A = X << 1 or A = M@$1234 & M@foo
 */
func (p *parser) parseExpression(token string) error {
	var err error
	expr := new(expression)

	// Parse the left side of the expression
	expr.dest.location, expr.dest.addrval, expr.dest.size, err = p.parseExpressionArg(token)
	if (err != nil) {
		return err
	} else if (expr.dest.location == VALUE) {
		return fmt.Errorf("the left side of the expression must be a register or memory address")
	}

	// Parse the equals operation
	p.skipWhitespace()
	sym1 := p.peekChar()
	sym2 := p.peekAhead(1)
	sym3 := p.peekAhead(2)
	if (sym1 == '=') {
		p.skip(1)
		expr.equalOp = EQUALS
	} else if (sym1 == '+') && (sym2 == '=') {
		p.skip(2)
		expr.equalOp = PLUS
	} else if (sym1 == '-') && (sym2 == '=') {
		p.skip(2)
		expr.equalOp = MINUS
	} else if (sym1 == '&') && (sym2 == '=') {
		p.skip(2)
		expr.equalOp = AND
	} else if (sym1 == '|') && (sym2 == '=') {
		p.skip(2)
		expr.equalOp = OR
	} else if (sym1 == '^') && (sym2 == '=') {
		p.skip(2)
		expr.equalOp = EOR
	} else if (sym1 == '<') && (sym2 == '<') && (sym3 == '=') {
		p.skip(3)
		expr.equalOp = SHIFT_LEFT
	} else if (sym1 == '>') && (sym2 == '>') && (sym3 == '=') {
		p.skip(3)
		expr.equalOp = SHIFT_RIGHT
	} else {
		fmt.Errorf("Missing '=', found '%c'", sym1)
	}

	// Parse the first argument on the right side of the expression
	p.skipWhitespace()
	arg1 := strings.ToLower(p.nextAZ_az_09())
	expr.src1.location, expr.src1.addrval, expr.src1.size, err = p.parseExpressionArg(arg1)
	if (err != nil) {
		return err
	}

	// Operation
	p.skipWhitespace()
	sym1 = p.peekChar()
	sym2 = p.peekAhead(1)
	if (sym1 == '+') {
		p.skip(1)
		expr.op = PLUS
	} else if (sym1 == '-') {
		p.skip(1)
		expr.op = MINUS
	} else if (sym1 == '&') {
		p.skip(1)
		expr.op = AND
	} else if (sym1 == '|') {
		p.skip(1)
		expr.op = OR
	} else if (sym1 == '^') {
		p.skip(1)
		expr.op = EOR
	} else if (sym1 == '<') && (sym2 == '<') {
		p.skip(2)
		expr.op = SHIFT_LEFT
	} else if (sym1 == '>') && (sym2 == '>') {
		p.skip(2)
		expr.op = SHIFT_RIGHT
	} else {
		expr.op = NO_OP
	}

	// For any ?= except =, a second arg is not allowed
	if ((expr.equalOp != EQUALS) && (expr.op != NO_OP)) {
		switch (expr.equalOp) {
		case PLUS: return fmt.Errorf("Expressions with '+=', can't also have a '%c'", sym1)
		case MINUS: return fmt.Errorf("Expressions with '-=', can't also have a '%c'", sym1)
		case AND: return fmt.Errorf("Expressions with '&=', can't also have a '%c'", sym1)
		case OR: return fmt.Errorf("Expressions with '|=', can't also have a '%c'", sym1)
		case EOR: return fmt.Errorf("Expressions with '^=', can't also have a '%c'", sym1)
		case SHIFT_LEFT: return fmt.Errorf("Expressions with '<<=', can't also have a '%c'", sym1)
		case SHIFT_RIGHT: return fmt.Errorf("Expressions with '>>=', can't also have a '%c'", sym1)
		}
	}

	// The second argument isn't needed if there is no operation in the expression, e.g. A = 123
	if (expr.op == NO_OP) {
		expr.src2.location = VALUE
		expr.src2.addrval = 0
		expr.src2.size = R08
	} else {
		p.skipWhitespace()
		arg2 := strings.ToLower(p.nextAZ_az_09())
		expr.src2.location, expr.src2.addrval, expr.src2.size, err = p.parseExpressionArg(arg2)
		if (err != nil) {
			return err
		}
	}

	// // Generate the code
	p.addExpression(expr)

	// The simplest expression: R = V
	if (expr.equalOp == EQUALS) && (expr.op == NO_OP) {
		return p.generateExpressionEquals(expr)
	// Next simplest is : R ?= V
	} else if (expr.op == NO_OP) {
		return p.generateExpressionOpEquals(expr)
	}

	// NOT YET IMPLEMENTED
	p.addExprInstruction("nop", modeImplicit, R08, 0) // nop

	return nil
}
	

/*
 *  Generate the code for R|M = R|M|V
 */
func (p *parser) generateExpressionEquals(expr *expression) error {
	size := unionAddressMode(expr.dest.size, expr.src1.size)
	saveRestoreSize := p.lastAsz
	
	switch (expr.dest.location) {
	case MEMORY, VARIABLE:
		switch (expr.src1.location) {
		case MEMORY, VARIABLE:
			p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // sta R0
			if (expr.src1.addrval <= 0x0FF) {
				p.addExprInstruction("lda", modeZeroPage, size, expr.src1.addrval)
			} else {
				p.addExprInstruction("lda", modeAbsolute, size, expr.src1.addrval)
			}
			if (expr.dest.addrval <= 0x0FF) {
				p.addExprInstruction("sta", modeZeroPage, size, expr.dest.addrval)
			} else {
				p.addExprInstruction("sta", modeAbsolute, size, expr.dest.addrval)
			}
			p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
		case VALUE:
			p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // sta R0
			p.addExprInstruction("lda", modeImmediate, expr.src1.size, expr.src1.addrval)
			if (expr.dest.addrval <= 0x0FF) {
				p.addExprInstruction("sta", modeZeroPage, size, expr.dest.addrval)
			} else {
				p.addExprInstruction("sta", modeAbsolute, size, expr.dest.addrval)
			}
			p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
		case REG_A:
			if (expr.dest.addrval <= 0x0FF) {
				p.addExprInstruction("sta", modeZeroPage, size, expr.dest.addrval)
			} else {
				p.addExprInstruction("sta", modeAbsolute, size, expr.dest.addrval)
			}
		case REG_X:
			if (expr.dest.addrval <= 0x0FF) {
				p.addExprInstruction("stx", modeZeroPage, size, expr.dest.addrval)
			} else {
				p.addExprInstruction("stx", modeAbsolute, size, expr.dest.addrval)
			}
		case REG_Y:
			if (expr.dest.addrval <= 0x0FF) {
				p.addExprInstruction("sty", modeZeroPage, size, expr.dest.addrval)
			} else {
				p.addExprInstruction("sty", modeAbsolute, size, expr.dest.addrval)
			}
		}
	case REG_A:
		switch (expr.src1.location) {
		case MEMORY, VARIABLE:
			if (expr.src1.addrval <= 0x0FF) {
				p.addExprInstruction("lda", modeZeroPage, size, expr.src1.addrval)
			} else {
				p.addExprInstruction("lda", modeAbsolute, size, expr.src1.addrval)
			}
		case VALUE:
			p.addExprInstruction("lda", modeImmediate, size, expr.src1.addrval)
		case REG_A:
			return fmt.Errorf("A = A is not allowed")
		case REG_X:
			p.addExprInstruction("txa", modeImplicit, size, 0)
		case REG_Y:
			p.addExprInstruction("tya", modeImplicit, size, 0)
		}
	case REG_X:
		switch (expr.src1.location) {
		case MEMORY, VARIABLE:
			if (expr.src1.addrval <= 0x0FF) {
				p.addExprInstruction("ldx", modeZeroPage, size, expr.src1.addrval)
			} else {
				p.addExprInstruction("ldx", modeAbsolute, size, expr.src1.addrval)
			}
		case VALUE:
			p.addExprInstruction("ldx", modeImmediate, expr.src1.size, expr.src1.addrval)
		case REG_A:
			p.addExprInstruction("tax", modeImplicit, size, 0)
		case REG_X:
			return fmt.Errorf("X = X is not allowed")
		case REG_Y:
			p.addExprInstruction("sty", modeZeroPage, size, 0) // sty R0
			p.addExprInstruction("ldx", modeZeroPage, size, 0) // ldx R0
		}
	case REG_Y:
		switch (expr.src1.location) {
		case MEMORY, VARIABLE:
			if (expr.src1.addrval <= 0x0FF) {
				p.addExprInstruction("ldy", modeZeroPage, size, expr.src1.addrval)
			} else {
				p.addExprInstruction("ldy", modeAbsolute, size, expr.src1.addrval)
			}
		case VALUE:
			p.addExprInstruction("ldy", modeImmediate, expr.src1.size, expr.src1.addrval)
		case REG_A:
			p.addExprInstruction("tay", modeImplicit, size, 0)
		case REG_X:
			p.addExprInstruction("stx", modeZeroPage, size, 0) // sty R0
			p.addExprInstruction("ldy", modeZeroPage, size, 0) // ldx R0
		case REG_Y:
			return fmt.Errorf("Y = Y is not allowed")
		}
	}

	return nil
}

/*
 *  Generate the code for R|M ?= R|M|V
 */
func (p *parser) generateExpressionOpEquals(expr *expression) error {
	size := unionAddressMode(expr.dest.size, expr.src1.size)
	saveRestoreSize := p.lastAsz

	switch (expr.equalOp) {
	case PLUS:
		switch (expr.dest.location) {
		case MEMORY, VARIABLE:
			p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // sta R0
			p.addExprInstruction("clc", modeImplicit, 0, 0)
			var mmm string
			if (expr.src1.location == REG_A) {
				mmm = "adc"
			} else {
				mmm = "lda"
			}
			if (expr.dest.addrval <= 0x0FF) {
				p.addExprInstruction(mmm, modeZeroPage, expr.dest.size, expr.dest.addrval)
			} else {
				p.addExprInstruction(mmm, modeAbsolute, expr.dest.size, expr.dest.addrval)
			}
			switch (expr.src1.location) {
			case MEMORY, VARIABLE:
				if (expr.src1.addrval <= 0x0FF) {
					p.addExprInstruction("adc", modeZeroPage, expr.src1.size, expr.src1.addrval) // @@@ FLAW: if src1.size < dst.size then adc truncates partial sum
				} else {
					p.addExprInstruction("adc", modeAbsolute, expr.src1.size, expr.src1.addrval)
				}
			case VALUE:
				p.addExprInstruction("adc", modeImmediate, expr.src1.size, expr.src1.addrval) // @@@ FLAW: if src1.size < dst.size then adc truncates partial sum
			case REG_X:
				p.addExprInstruction("adx", modeImmediate, expr.src1.size, expr.src1.addrval)
			case REG_Y:
				p.addExprInstruction("ady", modeImmediate, expr.src1.size, expr.src1.addrval)
			}
			if (expr.dest.addrval <= 0x0FF) {
				p.addExprInstruction("sta", modeZeroPage, size, expr.dest.addrval)
			} else {
				p.addExprInstruction("sta", modeAbsolute, size, expr.dest.addrval)
			}
			p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
		case REG_A:
			switch (expr.src1.location) {
			case MEMORY, VARIABLE:
				p.addExprInstruction("clc", modeImplicit, 0, 0)
				if (expr.src1.addrval <= 0x0FF) {
					p.addExprInstruction("adc", modeZeroPage, expr.src1.size, expr.src1.addrval)
				} else {
					p.addExprInstruction("adc", modeAbsolute, expr.src1.size, expr.src1.addrval)
				}
			case VALUE:
				if (expr.src1.addrval == 1) {
					p.addExprInstruction("inc", modeImplicit, 0, 0)
				} else {
					p.addExprInstruction("clc", modeImplicit, 0, 0)
					p.addExprInstruction("adc", modeImmediate, expr.src1.size, expr.src1.addrval)
				}
			case REG_A:
				p.addExprInstruction("asl", modeImplicit, 0, 0) // A << 1 = A + A
			case REG_X:
				p.addExprInstruction("adx", modeImplicit, size, 0)
			case REG_Y:
				p.addExprInstruction("ady", modeImplicit, size, 0)
			}
		case REG_X:
			switch (expr.src1.location) {
			case MEMORY, VARIABLE:
				p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // sta R0
				p.addExprInstruction("txa", modeImplicit, expr.dest.size, 0)
				p.addExprInstruction("clc", modeImplicit, 0, 0)
				if (expr.src1.addrval <= 0x0FF) {
					p.addExprInstruction("adc", modeZeroPage, expr.src1.size, expr.src1.addrval)
				} else {
					p.addExprInstruction("adc", modeAbsolute, expr.src1.size, expr.src1.addrval)
				}
				p.addExprInstruction("tax", modeImplicit, size, 0)
				p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
			case VALUE:
				p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // sta R0
				p.addExprInstruction("tya", modeImplicit, expr.dest.size, 0)
				p.addExprInstruction("clc", modeImplicit, 0, 0)
				if (expr.src1.addrval <= 0x0FF) {
					p.addExprInstruction("adc", modeZeroPage, expr.src1.size, expr.src1.addrval)
				} else {
					p.addExprInstruction("adc", modeAbsolute, expr.src1.size, expr.src1.addrval)
				}
				p.addExprInstruction("tay", modeImplicit, size, 0)
				p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
			case REG_A:
				p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // sta R0
				p.addExprInstruction("clc", modeImplicit, 0, 0)
				p.addExprInstruction("adx", modeImplicit, size, 0)
				p.addExprInstruction("tax", modeImplicit, size, 0)
				p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
			case REG_X:
				p.addExprInstruction("xsl", modeImplicit, 0, 0) // X << 1 = X + X
			case REG_Y:
				p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // sta R0
				p.addExprInstruction("txa", modeImplicit, expr.dest.size, 0)
				p.addExprInstruction("clc", modeImplicit, 0, 0)
				p.addExprInstruction("ady", modeImplicit, size, 0)
				p.addExprInstruction("tax", modeImplicit, size, 0)
				p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
			}
		case REG_Y:
			// TBD
		}
		return nil
	case MINUS:
	case AND:
		switch (expr.dest.location) {
		case MEMORY, VARIABLE:
			// TBD
			p.addExprInstruction("nop", modeImplicit, R08, 0) // nop
		case REG_A:
			switch (expr.src1.location) {
			case MEMORY, VARIABLE:
				if (expr.src1.addrval <= 0x0FF) {
					p.addExprInstruction("and", modeZeroPage, expr.src1.size, expr.src1.addrval)
				} else {
					p.addExprInstruction("and", modeAbsolute, expr.src1.size, expr.src1.addrval)
				}
			case VALUE:
				p.addExprInstruction("and", modeImmediate, expr.src1.size, expr.src1.addrval)
			case REG_A:
				return fmt.Errorf("A &= A isn't allowed")
			case REG_X:
				return fmt.Errorf("A &= X isn't allowed")
			case REG_Y:
				return fmt.Errorf("A &= Y isn't allowed")
			}
		case REG_X:
			// TBD
			p.addExprInstruction("nop", modeImplicit, R08, 0) // nop
		case REG_Y:
			// TBD
			p.addExprInstruction("nop", modeImplicit, R08, 0) // nop
		}
		return nil
	case OR:
	case EOR:
	case SHIFT_LEFT:
		switch (expr.src1.location) {
		case MEMORY, VARIABLE:
			switch (expr.dest.location) {
			case MEMORY, VARIABLE:
				// TBD
			case REG_A:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("asl", modeImplicit, expr.dest.size, 0)
				}
			case REG_X:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("xsl", modeImplicit, expr.dest.size, 0)
				}
			case REG_Y:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("ysl", modeImplicit, expr.dest.size, 0)
				}
			}
		case VALUE:
			switch (expr.dest.location) {
			case MEMORY, VARIABLE:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("asl", modeAbsolute, expr.src1.size, expr.src1.addrval) // asl M[aaa]
				}
			case REG_A:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("asl", modeImplicit, expr.dest.size, 0)
				}
			case REG_X:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("xsl", modeImplicit, expr.dest.size, 0)
				}
			case REG_Y:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("ysl", modeImplicit, expr.dest.size, 0)
				}
			}
		case REG_A:
		case REG_X:
		case REG_Y:
		}
		return nil
	case SHIFT_RIGHT:
		if (expr.src1.location == VALUE) {
			switch (expr.dest.location) {
			case MEMORY, VARIABLE:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("lsr", modeAbsolute, expr.src1.size, expr.src1.addrval) // asl M[aaa]
				}
			case REG_A:
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("lsr", modeImplicit, expr.dest.size, 0)
				}
			case REG_X:
				p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // lda R0
				p.addExprInstruction("txa", modeImplicit, expr.dest.size, 0)
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("lsr", modeImplicit, expr.dest.size, 0)
				}
				p.addExprInstruction("tax", modeImplicit, expr.dest.size, 0)
				p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
			case REG_Y:
				p.addExprInstruction("sta", modeZeroPage, saveRestoreSize, 0) // lda R0
				p.addExprInstruction("tya", modeImplicit, expr.dest.size, 0)
				for i := 0; i < expr.src1.addrval; i++ {
					p.addExprInstruction("lsr", modeImplicit, expr.dest.size, 0)
				}
				p.addExprInstruction("tay", modeImplicit, expr.dest.size, 0)
				p.addExprInstruction("lda", modeZeroPage, saveRestoreSize, 0) // lda R0
			}
		}
		return nil
	}

	// NOT YET IMPLEMENTED
	p.addExprInstruction("nop", modeImplicit, R08, 0) // nop

	return nil
}

	
/*
 *  Parse a source or destination of an expresson
 *  e.g. A, X, Y, M$1234 or M.reference
 */
func (p *parser) parseExpressionArg(token string) (int, int, int, error) {
	// Parse the token name
	switch (token) {
	case "a": return REG_A, 0, p.parseOpWidth(), nil
	case "x": return REG_X, 0, p.parseOpWidth(), nil
	case "y": return REG_Y, 0, p.parseOpWidth(), nil
	case "r":
		// Rnnn  ; nnn must be decimal 0-256
		p.skip(1)
		sym := p.peekChar()
		if (sym < '0') || (sym > '9') {
			return 0, 0, 0, fmt.Errorf("register numbers must be specified as a number between 0 and 255")
		}
		address := p.nextDecimal()
		if (address > 255) {
			return 0, 0, 0, fmt.Errorf("register numbers must be between 0 and 255, not #%d", address)
		}

		return MEMORY, address, p.parseOpWidth(), nil
	case "m":
		sym := p.peekChar()
		if (sym != '@') {
			fmt.Errorf("M followed unexpectedly by %c instead of '$'", sym)
		}

		// M@$aaaa or M@label
		p.skip(1)
		address, err := p.parseMemoryAddress()
		if (err != nil) {
			return 0, 0, 0, err
		}

		return MEMORY, address, p.parseOpWidth(), nil
	case "":
		// @variable
		if p.peekChar() == '@' {
			p.skip(1)
			token = p.nextAZ_az_09()
			address, size, err := p.lookupVariable(p.currentCode, token)
			if (err != nil) {
				return 0, 0, 0, fmt.Errorf("unknown variable '@%s'", token)
			}
			return MEMORY, address, size, nil
		// number
		} else {
			value, err := p.nextValue()
			if (err != nil) {
				return 0, 0, 0, fmt.Errorf("expected a value on the right side of the expression")
			}
			return VALUE, value, valueToPrefix(value), err
		}
	default:
		value, err := p.lookupConstant(token)
		if (err == nil) {
			return VALUE, value, valueToPrefix(value), nil
		}
		return 0, 0, 0, fmt.Errorf("constant '%s' not found", token)
	}
}


/*
 *  Parse a reference to a memory address
 *  e.g. M@$1234 or M@reference
 */
func (p *parser) parseMemoryAddress() (int, error) {
	sym := p.peekChar()
	if (sym == '$') { // M@$aaaa
		return p.nextValue()
	}
	
	// M@label
	token := p.nextAZ_az_09()
	value, err := p.lookupConstant(token)
	if (err == nil) {
		return value, err
	}
	data := p.lookupDataName(token)
	if (data != nil) {
		return data.startAddr, err
	}
	return 0, fmt.Errorf("M@%s is an unknown variable address", token)
}


/*
 *  Add a blank instruction to comment the expression
 */
func (p *parser) addExpression(expr *expression) {
	instr := new(instruction)
	if (p.currentCode.instr == nil) {
		p.currentCode.instr = instr
		instr.prev = nil
	} else if p.currentCode.lastInstr != nil {
		instr.prev = p.currentCode.lastInstr
		p.currentCode.lastInstr.next = instr
	}
	p.currentCode.lastInstr = instr
	instr.next = nil
	instr.len = 0
	instr.hasValue = true
	instr.expr = expr
	instr.address = p.currentCode.endAddr
}

/*
 *  Add an instruction to implement the expression
 */
func (p *parser) addExprInstruction(mmm string, addressMode int, size int, value int) *instruction {
	return p.addExprInstructionWithSymbol(mmm, addressMode, size, value, "", true)
}
func (p *parser) addExprInstructionWithSymbol(mmm string, addressMode int, size int, value int, symbol string, hasValue bool) *instruction {
	instr := new(instruction)
	if (p.currentCode.instr == nil) {
		p.currentCode.instr = instr
		instr.prev = nil
	} else if p.currentCode.lastInstr != nil {
		instr.prev = p.currentCode.lastInstr
		p.currentCode.lastInstr.next = instr
	}
	p.currentCode.lastInstr = instr
	instr.next = nil

	var o *opcode
	instr.mnemonic, o, _ = lookupMnemonic(mmm, addressMode, size)
	instr.prefix = o.size
	instr.opcode = o.opcode
	instr.addressMode = o.mode
	instr.len = o.len
	instr.value = value
	instr.symbol = symbol
	instr.symbolLC = strings.ToLower(symbol)
	instr.hasValue = hasValue

	// Remember the size of the opcode for each of A, X, and Y registers
	switch (mnemonics[instr.mnemonic].reg) {
	case REG_A:
		p.lastAsz = size
	case REG_X:
		p.lastXsz = size
	case REG_Y:
		p.lastYsz = size
	}

	instr.address = p.currentCode.endAddr
	p.currentCode.endAddr += instr.len

	return instr
}


	
	
