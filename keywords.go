package aCCembler

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

var keywords = []string {
	"var",
	"print",
	"os",
	"if",
	"loop",
	"for",
	"do",
	"while",
	"break",
	"continue",
	"return",
}

// Boolean expression in IF, WHILE, etc.
type boolExpr struct {
	opEq 		bool
	opNe 		bool
	opGe 		bool
	opGt 		bool
	opLt 		bool
	opLe 		bool
	opPl 		bool
	opMi 		bool

	value		int
	hasValue	bool
	size		int

	string		string
}


/*
 *  Lookup the keyword
 */
func (p *parser) isKeyword(token string) bool {
	// Find the table entry for the keyword
	for k := range keywords {
		if (token == keywords[k]) {
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
	switch (token) {
	case "var": return p.parseLocalVariable(token)
	case "print": return p.parsePrint(token)
	case "os": return p.parseOs(token)
	case "if": return p.parseIf(token)
	case "loop": return p.parseLoop(token)
	case "for": return p.parseFor(token)
	case "do": return p.parseDo(token)
	case "while": return p.parseWhile(token)
	case "continue": return p.parseContinue(token)
	case "break": return p.parseBreak(token)
	case "return": return p.parseReturn(token)
	}

	return fmt.Errorf("keyword '%s' is invalid", token)
}

/*
 *  Parse the 'var' keyword
 */
func (p *parser) parseLocalVariable(token string) error {
	p.skipWhitespace()
	return p.parseVariable(p.currentCode, p.nextAZ_az_09())
}

/*
 *  Parse the 'print' keyword
 */
func (p *parser) parsePrint(token string) error {
	fmt.Printf("PRINT is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}

/*
 *  Parse the 'os' keyword
 */
func (p *parser) parseOs(token string) error {
	fmt.Printf("OS is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}

/*
 *  Parse the 'if' keyword
 */
func (p *parser) parseIf(token string) error {
	sub := new(subBlock)
	sub.keyword = KW_IF
	sub.startAddr = p.currentCode.endAddr
	sub.endAddr = sub.startAddr

	// Optional (
	p.skipWhitespace()
	sym := p.peekChar()
	if (sym == '(') {
		p.skip(1)
	}

	be, err := p.parseBooleanExpression("IF")
	if (err != nil) {
		return err
	}

	p.skipWhitespace()
	if (p.nextChar() != '{') {
		return fmt.Errorf("missing { in IF")
	}
	p.skipWhitespaceAndEOL()

	// Name this construct
	name := fmt.Sprintf("IF%x", p.currentCode.endAddr)

	// Add the instruction with the sub in the current block (before starting a new block)
	comment := fmt.Sprintf("IF %s {", be.string)
	p.addKeywordInstruction(sub, comment)

	// Add the code for the IF
	b := p.addCodeBlock(sub, "IF", name, false);

	// Generate the code for the boolan expression
	endLabel := name + "_end"
	skipInstr := p.outputBooleanExpression(*be, endLabel)

	// Parse the code
	err = p.parseCode(name)
	if (err != nil) {
		return err
	}

	// Check to see if there is an ELSE
	nextToken := p.peekAZ_az_09()
	if (nextToken == "ELSE") {
		// Skip past else
		p.nextAZ_az_09()

		// Check for '{'
		p.skipWhitespace()
		if (p.nextChar() != '{') {
			return fmt.Errorf("missing { in ELSE")
		}
		p.skipWhitespaceAndEOL()

		// Jump from the end of the IF block to after the ELSE
		p.addExprInstructionWithSymbol("bra", modeRelative, A16, 0, endLabel, false)

		// Add a label to where the else goes, and change the NOT IF branch to the ELSE block
		if (skipInstr.prev != nil) && (skipInstr.prev.symbol == endLabel) {
			skipInstr.prev.symbol = name + "_else"
			skipInstr.prev.symbolLC = strings.ToLower(skipInstr.symbol)
		}
		elseLabel := name + "_else"
		skipInstr.symbol = elseLabel
		skipInstr.symbolLC = strings.ToLower(skipInstr.symbol)
		p.addInstructionLabel(elseLabel)

		// Go back to parsing code for the main block
		p.endCodeBlock(sub)

		// And add yet-another sub instruction to hold the ELSE code
		sub := new(subBlock)
		sub.keyword = KW_ELSE
		sub.startAddr = p.currentCode.endAddr
		sub.endAddr = sub.startAddr
		p.addKeywordInstruction(sub, "ELSE")

		// Add the code for the ELSE
		p.addCodeBlock(sub, "ELSE", name, false);

		// Parse the code
		err = p.parseCode(name)
		if (err != nil) {
			return err
		}

		// Go back to parsing code for the main block
		p.endCodeBlock(sub)

		// Add a label to the end of the block
		p.addInstructionLabel(b.name + "_end")
	} else {
		// Add a label to the end of the block
		p.addInstructionLabel(b.name + "_end")

		// Go back to parsing code for the main block
		p.endCodeBlock(sub)
	}

	return nil
}

/*
 *  Parse the 'loop' keyword
 */
func (p *parser) parseLoop(token string) error {
	sub := new(subBlock)
	sub.keyword = KW_LOOP
	sub.startAddr = p.currentCode.endAddr
	sub.endAddr = sub.startAddr

	p.skipWhitespace()
	if (p.nextChar() != '{') {
		return fmt.Errorf("missing { in LOOP")
	}
	p.skipWhitespaceAndEOL()

	// Name this construct
	name := fmt.Sprintf("LOOP%x", p.currentCode.endAddr)

	// Add the instruction with the sub in the current block (before starting a new block)
	comment := "LOOP {"
	p.addKeywordInstructionAndLabel(sub, comment, name)

	// Add the code for the LOOP
	b := p.addCodeBlock(sub, "LOOP", name, true);

	loopLabel := name + "_loop"
	p.addInstructionLabel(loopLabel)

	// Parse the code
	err := p.parseCode(name)
	if (err != nil) {
		return err
	}

	// How far back is the top of the loop?
	target, err := p.currentCode.lookupInstructionLabel(loopLabel)
	distance := p.currentCode.endAddr - target
	if (distance+2 < 0x07F) {
		p.addExprInstructionWithSymbol("bra", modeRelative, A16, -(distance+2), loopLabel, true)
	} else if (distance+4 < 0x7FFF) {
		p.addExprInstructionWithSymbol("bra", modeRelative, A24, -(distance+4), loopLabel, true)
	} else {
		p.addExprInstructionWithSymbol("jmp", modeAbsolute, A24, target, loopLabel, true)
	}

	// Add a label to the end of the block
	p.addInstructionLabel(b.name + "_end")

	// Go back to parsing code for the main block
	p.endCodeBlock(sub)

	return nil
}

/*
 *  Parse the 'for' keyword
 */
func (p *parser) parseFor(token string) error {
	sub := new(subBlock)
	sub.keyword = KW_FOR
	sub.startAddr = p.currentCode.endAddr
	sub.endAddr = sub.startAddr

	var err error
	var forAddress int
	var forAddressMode int
	var forRegister string
	forIsMemory := false
	forIsRegister := false
	var forMRegStr string
	endIsMemory := false
	forSz := R08
	p.skipWhitespace()
	sym1 := p.peekChar()
	sym2 := p.peekAhead(1)
	if ((sym1 == 'M') || (sym1 == 'm')) && (sym2 == '@') {
		p.skip(2)
		forAddress, err = p.parseMemoryAddress()
		if (err != nil) {
			return err
		}
		if (forAddress <= 0xff) { forAddressMode = modeZeroPage } else { forAddressMode = modeAbsolute }
		forIsMemory = true
		forMRegStr = fmt.Sprintf("M$%x", forAddress)
		forSz = p.parseOpWidth()
	} else if (sym1 == '@') {
		p.skip(1)
		symbol := p.nextAZ_az_09()
		forAddress, forSz, err = p.lookupVariable(p.lastCode, symbol)
		if (err != nil) {
			return fmt.Errorf("variable '@%s' not found", symbol)
		}
		if (forAddress <= 0xff) { forAddressMode = modeZeroPage } else { forAddressMode = modeAbsolute }
		forIsMemory = true
		forMRegStr = fmt.Sprintf("@%s", symbol)
	} else if (sym1 == '%') && ((sym2 == 'R') || (sym2 == 'r')) {
		p.skip(2)
		forAddress, err := p.nextValue()
		if (err != nil) {
			return fmt.Errorf("invalid register '%R%c'", p.peekChar())
		}
		forAddressMode = modeZeroPage
		forIsMemory = true
		forMRegStr = fmt.Sprintf("%R%d", forAddress)
	} else if (sym1 == 'A') {
		return fmt.Errorf("you can't iterate a FOR loop on register A")
	} else if (sym1 == 'X') {
		p.skip(1)
		forRegister = "X"
		forIsRegister = true
		forMRegStr = "X"
		forSz = p.parseOpWidth()
	} else if (sym1 == 'Y') {
		p.skip(1)
		forRegister = "Y"
		forIsRegister = true
		forMRegStr = "Y"
		forSz = p.parseOpWidth()
	}

	p.skipWhitespace()
	if (p.nextChar() != '=') {
		return fmt.Errorf("missing = in FOR")
	}

	var start int
	p.skipWhitespace()
	sym1 = p.peekChar()
	sym2 = p.peekAhead(1)
	if ((sym1 == 'M') || (sym1 == 'm')) && (sym2 == '@') {
		p.skip(2)
		start, err = p.parseMemoryAddress()
		if (err != nil) {
			return err
		}
	} else if p.isNextAZ() {
		symbol := p.nextAZ_az_09()
		start, err = p.lookupConstant(symbol)
		if (err != nil) {
			return fmt.Errorf("constant '%s' not found", symbol)
		}
	} else {
		start, err = p.nextValue()
		if (err != nil) {
			return fmt.Errorf("invalid starting value in FOR")
		}
	}

	to := strings.ToUpper(p.nextAZ_az_09())
	if (to == "DOWN") {
		sub.upDown = false
		to = strings.ToUpper(p.nextAZ_az_09())
	} else {
		sub.upDown = true
	}

	if (to != "TO") {
		return fmt.Errorf("syntax error in FOR, missing TO")
	}

	var end int
	p.skipWhitespace()
	sym1 = p.peekChar()
	sym2 = p.peekAhead(1)
	if ((sym1 == 'M') || (sym1 == 'm')) && (sym2 == '@') {
		p.skip(2)
		end, err = p.parseMemoryAddress()
		if (err != nil) {
			return err
		}
		endIsMemory = true
	} else if (sym1 == '@') {
		p.skip(1)
		symbol := p.nextAZ_az_09()
		end, forSz, err = p.lookupVariable(p.lastCode, symbol)
		if (err != nil) {
			return fmt.Errorf("variable '@%s' not found", symbol)
		}
		if (end <= 0xff) { forAddressMode = modeZeroPage } else { forAddressMode = modeAbsolute }
		endIsMemory = true
	} else if (sym1 == '%') && ((sym2 == 'R') || (sym2 == 'r')) {
		p.skip(2)
		end, err = p.nextValue()
		if (err != nil) {
			return fmt.Errorf("invalid register '%R%c'", p.peekChar())
		}
		forAddressMode = modeZeroPage
		endIsMemory = true
	} else if p.isNextAZ() {
		symbol := p.nextAZ_az_09()
		end, err = p.lookupConstant(symbol)
		if (err != nil) {
			return fmt.Errorf("constant '%s' not found", symbol)
		}
	} else {
		end, err = p.nextValue()
		if (err != nil) {
			return fmt.Errorf("invalid ending value in FOR")
		}
	}

	p.skipWhitespace()
	if (p.nextChar() != '{') {
		return fmt.Errorf("missing { in FOR")
	}
	p.skipWhitespaceAndEOL()

	// Name this construct
	name := fmt.Sprintf("FOR%x", p.currentCode.endAddr)

	// Add the instruction with the sub in the current block (before starting a new block)
	down := "DOWN "
	if (sub.upDown == true) { down = "" }
	comment := fmt.Sprintf("FOR %s = %d %sTO %d {", forMRegStr, start, down, end)
	p.addKeywordInstructionAndLabel(sub, comment, name)

	// Add the code for the IF
	b := p.addCodeBlock(sub, "FOR", name, true);

	loopSz := R08
	if (start > 0x0FFFFFF) || (end > 0x0FFFFFF) {
		fmt.Errorf("FOR loop doesn't fit in 24-bits %d TO %d", start, end)
	} else if (start > 0x0FFFF) || (end > 0x0FFFF) {
		loopSz = R24
	} else if (start > 0x0FF) || (end > 0x0FF) {
		loopSz = R16
	}
	if (forSz != loopSz) {
		fmt.Errorf("FOR loop range doesn't match the size of the loop register/varaible/memory")
	}

	// Load the start value of the loop
	if (forIsMemory) {
		p.addExprInstruction("ldx", modeImmediate, loopSz, start)
		p.addExprInstruction("stx", forAddressMode, forSz, forAddress)
	} else if (forIsRegister) {
		if (forRegister == "X") {
			p.addExprInstruction("ldx", modeImmediate, loopSz, start)
		} else if (forRegister == "Y") {
			p.addExprInstruction("ldy", modeImmediate, loopSz, start)
		} else {
			fmt.Errorf("unknown FOR register %s", forRegister)
		}
	}
	p.addInstructionLabel(name + "_loop")

	// Parse the code
	err = p.parseCode(name)
	if (err != nil) {
		return err
	}

	// Increment/Decrement the loop count
	var mmm string
	if (forIsMemory) {
		if sub.upDown { mmm = "inc" } else { mmm = "dec"}
		p.addExprInstruction(mmm, forAddressMode, forSz, forAddress)
		p.addExprInstruction("ldx", forAddressMode, forSz, forAddress)
		p.addExprInstruction("cpx", modeImmediate, loopSz, end+1)
	} else if (forIsRegister) {
		if (forRegister == "X") {
			if sub.upDown { mmm = "inx" } else { mmm = "dex"}
			p.addExprInstruction(mmm, modeImplicit, loopSz, 0)
			if (endIsMemory) {
				p.addExprInstruction("cpx", modeAbsolute, loopSz, end)
			} else {
				p.addExprInstruction("cpx", modeImmediate, loopSz, end+1)
			}
		} else if (forRegister == "Y") {
			if sub.upDown { mmm = "iny" } else { mmm = "dey"}
			p.addExprInstruction(mmm, modeImplicit, loopSz, 0)
			if (endIsMemory) {
				p.addExprInstruction("cpy", modeAbsolute, loopSz, end)
			} else {
				p.addExprInstruction("cpy", modeImmediate, loopSz, end+1)
			}
		}
	}
	p.addExprInstructionWithSymbol("bne", modeRelative, A16, 0, name + "_loop", false)

	// Add a label to the end of the block
	p.addInstructionLabel(b.name + "_end")

	// Go back to parsing code for the main block
	p.endCodeBlock(sub)

	return nil
}

/*
 *  Parse the 'do' keyword
 */
func (p *parser) parseDo(token string) error {
	sub := new(subBlock)
	sub.keyword = KW_DO
	sub.startAddr = p.currentCode.endAddr
	sub.endAddr = sub.startAddr

	p.skipWhitespace()
	if (p.nextChar() != '{') {
		return fmt.Errorf("missing { in DO")
	}
	p.skipWhitespaceAndEOL()

	// Name this construct
	name := fmt.Sprintf("DO%x", p.currentCode.endAddr)

	// Add the instruction with the sub in the current block (before starting a new block)
	comment := "DO {"
	p.addKeywordInstructionAndLabel(sub, comment, name)

	// Add the code for the DO
	p.addCodeBlock(sub, "DO", name, true);

	loopLabel := name + "_loop"
	p.addInstructionLabel(loopLabel)

	// Parse the code
	err := p.parseCode(name)
	if (err != nil) {
		return err
	}

	while := strings.ToUpper(p.nextAZ_az_09())
	if (while != "WHILE") {
		return fmt.Errorf("DO without WHILE")
	}

	// Optional (
	p.skipWhitespace()
	sym := p.peekChar()
	if (sym == '(') {
		p.skip(1)
	}

	// Parse the WHILE
	var be *boolExpr
	be, err = p.parseBooleanExpression("WHILE")
	if (err != nil) {
		return err
	}
	p.skipWhitespaceAndEOL()

	// Generate the code for the boolean expression
	p.addInstructionComment("WHILE " + be.string)
	endLabel := name + "_end"
	p.outputBooleanExpression(*be, endLabel)

	// How far back is the top of the loop?
	target, err := p.currentCode.lookupInstructionLabel(loopLabel)
	distance := p.currentCode.endAddr - target
	if (distance+2 < 0x07F) {
		p.addExprInstructionWithSymbol("bra", modeRelative, A16, -(distance+2), loopLabel, true)
	} else if (distance+4 < 0x7FFF) {
		p.addExprInstructionWithSymbol("bra", modeRelative, A24, -(distance+4), loopLabel, true)
	} else {
		p.addExprInstructionWithSymbol("jmp", modeAbsolute, A24, target, loopLabel, true)
	}

	// Add a label to the end of the block
	p.addInstructionLabel(endLabel)

	// Go back to parsing code for the main block
	p.endCodeBlock(sub)

	return nil
}

/*
 *  Parse the 'while' keyword
 */
func (p *parser) parseWhile(token string) error {
	fmt.Printf("WHILE is not yet supported [%d-%d]\n", p.i, p.n)
	return nil
}

/*
 *  Parse the 'continue' keyword
 */
func (p *parser) parseContinue(token string) error {
	p.addInstructionComment("CONTINUE")

	loop := p.currentLoopBlock()
	if (loop == nil) {
		return fmt.Errorf("CONTINUE called outside of a loop")
	}

	// How far back is the top of the loop?
	loopLabel := strings.ToLower(loop.name + "_start")
	target, _ := p.currentCode.lookupInstructionLabel(loopLabel)
	distance := p.currentCode.endAddr - target
	if (distance+2 < 0x07F) {
		p.addExprInstructionWithSymbol("bra", modeRelative, A16, -(distance+2), loopLabel, true)
	} else if (distance+4 < 0x7FFF) {
		p.addExprInstructionWithSymbol("bra", modeRelative, A24, -(distance+4), loopLabel, true)
	} else {
		p.addExprInstructionWithSymbol("jmp", modeAbsolute, A24, target, loopLabel, true)
	}

	return nil
}

/*
 *  Parse the 'break' keyword
 */
func (p *parser) parseBreak(token string) error {
	p.addInstructionComment("BREAK")

	loop := p.currentLoopBlock()
	if (loop == nil) {
		return fmt.Errorf("CONTINUE called outside of a loop")
	}

	label := strings.ToLower(loop.name + "_end")
	p.addExprInstructionWithSymbol("bra", modeRelative, A24, 0, label, false)

	return nil
}

/*
 *  Parse the 'return' keyword
 */
func (p *parser) parseReturn(token string) error {
	hasValue := false
	value := 0
	var err error

	// Constant
	if (p.isNextAZ()) {
		symbol := p.nextAZ_az_09()
		value, err = p.lookupConstant(symbol)
		if (err != nil) {
			return fmt.Errorf("invalid constant '%s' in RETURN", symbol)
		}
	// Value
	} else if (p.isNext09()) {
		value, err = p.nextValue()
		if (err != nil) {
			return fmt.Errorf("invalid value in RETURN")
		}
		hasValue = true
	}

	if hasValue {
		returnSz := R08
		if value > 0x0FFFFFF {
			fmt.Errorf("RETURN value %d doesn't fit in 24-bits", value)
		} else if value > 0x0FFFF {
			returnSz = R24
		} else if value > 0x0FF {
			returnSz = R16
		}
		p.addExprInstruction("lda", modeImmediate, returnSz, value)
	}
	p.addExprInstruction("rts", modeImplicit, A24, 0)

	return nil
}


/*
 *  Parse the boolean expression in an IF, WHILE, etc.
 */
func (p *parser) parseBooleanExpression(keyword string) (*boolExpr, error) {
	be := new(boolExpr)

	// Reset the possibilities
	be.opEq = false
	be.opNe = false
	be.opGe = false
	be.opGt = false
	be.opLt = false
	be.opLe = false
	be.opPl = false
	be.opMi = false

	// Parse the expression
	p.skipWhitespace()
	sym1 := p.peekChar()
	sym2 := p.peekAhead(1)
	if (sym1 == '=') && (sym2 == '=') {
		be.opEq = true
		p.skip(2)
	} else if (sym1 == '!') && (sym2 == '=') {
		be.opNe = true
		p.skip(2)
	} else if (sym1 == '>') && (sym2 == '=') {
		be.opGe = true
		p.skip(2)
	} else if (sym1 == '<') && (sym2 == '=') {
		be.opLe = true
		p.skip(2)
	} else if (sym1 == '>') {
		be.opGt = true
		p.skip(1)
	} else if (sym1 == '<') {
		be.opLt = true
		p.skip(1)
	} else if (sym1 == '+') {
		be.opPl = true
		p.skip(1)
	} else if (sym1 == '-') {
		be.opMi = true
		p.skip(1)
	} else {
		return nil, fmt.Errorf("%s %c%c is an unknown syntax", keyword, sym1, sym2)
	}

	var err error
	be.hasValue = false
	be.size = R08
	p.skipWhitespace()
	sym := p.peekChar()
	if (sym != '{') && (sym != ')') {
		// Constant
		if (p.isNextAZ()) {
			symbol := p.nextAZ_az_09()
			be.value, err = p.lookupConstant(symbol)
			if (err != nil) {
				return nil, fmt.Errorf("invalid constant '%s' in %s", symbol, keyword)
			}
			if (p.peekChar() == '+') {
				p.skip(1)
				plus, _ := p.nextValue()
				be.value += plus
			}
		// Value
		} else {
			be.value, err = p.nextValue()
			if (err != nil) {
				return nil, fmt.Errorf("invalid value in %s", keyword)
			}
		}
		
		be.hasValue = true
		if (be.value > 0x0FFFFFF) {
			fmt.Errorf("%s %d does not fit into 24-bits", keyword, be.value)
		} else if (be.value > 0x0FFFF) {
			be.size = R24
		} else if (be.value > 0x0FF) {
			be.size = R16
		}
	}
	if (sym == ')') {
		p.skip(1)
	}
	p.skipWhitespace()
	sym = p.peekChar()
	if (sym == ')') {
		p.skip(1)
	}

	if (be.opEq) { be.string = "=="
	} else if (be.opNe) { be.string = "!="
	} else if (be.opGe) { be.string = ">="
	} else if (be.opGt) { be.string = ">"
	} else if (be.opLt) { be.string = "<"
	} else if (be.opLe) { be.string = "<="
	} else if (be.opPl) { be.string = "+"
	} else if (be.opMi) { be.string = "-" }
	if be.hasValue {
		be.string += fmt.Sprintf(" %d", be.value)
	}

	return be, nil
}

/*
 *  Output the code for the boolean expression in an IF, WHILE, etc.
 */
func (p *parser) outputBooleanExpression(be boolExpr, label string) *instruction {
	// If there a value, then do the compare
	if (be.hasValue) {
		p.addExprInstruction("cmp", modeImmediate, be.size, be.value)
	}

	// Generate the opposite branch logic to skip the block
	if (be.opEq) {
		return p.addExprInstructionWithSymbol("bne", modeRelative, A16, 0, label, false)
	} else if (be.opNe) {
		return p.addExprInstructionWithSymbol("beq", modeRelative, A16, 0, label, false)
	} else if (be.opGe) {
		return p.addExprInstructionWithSymbol("bcc", modeRelative, A16, 0, label, false)
	} else if (be.opGt) {
		p.addExprInstructionWithSymbol("beq", modeRelative, A16, 0, label, false)
		return p.addExprInstructionWithSymbol("bcc", modeRelative, A16, 0, label, false)
	} else if (be.opLt) {
		return p.addExprInstructionWithSymbol("bcs", modeRelative, A16, 0, label, false)
	} else if (be.opLe) {
		p.addExprInstructionWithSymbol("beq", modeRelative, A16, 0, label + "_eq", false)
		skipIntr := p.addExprInstructionWithSymbol("bcs", modeRelative, A16, 0, label, false)
		p.addInstructionLabel(label + "_eq")
		return skipIntr
	} else if (be.opPl) {
		return p.addExprInstructionWithSymbol("bmi", modeRelative, A16, 0, label, false)
	} else if (be.opMi) {
		return p.addExprInstructionWithSymbol("bpl", modeRelative, A16, 0, label, false)
	}

	return nil
}


/*
 *  Add a code block for the keyword's instructions
 */
func (p *parser) addCodeBlock(sub *subBlock, keyword string, name string, isLoop bool) *codeBlock {
	// Store this code block
	block := new(codeBlock)

	sub.block = block
	sub.block.up = p.currentCode
	p.currentCode = block

	block.next = nil
	block.startAddr = sub.block.up.endAddr
	block.endAddr = block.startAddr
	block.name = name
	block.nameLC = strings.ToLower(block.name)
	block.isLoop = isLoop
	block.instr = nil

	return block
}


/*
 *  Done with the block for the keyword's instructions
 */
func (p *parser) endCodeBlock(sub *subBlock) {
	p.currentCode = sub.block.up
	p.currentCode.endAddr = sub.block.endAddr
}

/*
 *  Add an sub label
 */
func (p *parser) addKeywordInstructionAndLabel(sub *subBlock, comment string, name string) {
	p.addInstructionLabel(name + "_start")
	p.addKeywordInstruction(sub, comment)
}

/*
 *  Add an sub instruction
 */
func (p *parser) addKeywordInstruction(sub *subBlock, comment string) {
	p.addInstructionComment(comment)
	p.currentCode.lastInstr.subBlock = sub
}


/*
 *  Search up through the code blocks to find the most recent loop
 */
func (p *parser) currentLoopBlock() *codeBlock {
	b := p.currentCode
	for b.isLoop == false {
		if (b.up == nil) {
			return nil
		}

		b = b.up
	}

	return b
}
