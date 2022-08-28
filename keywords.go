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

var keywords = []string {
	"var",
	"print",
	"os",
	"if",
	"loop",
	"for",
	"while",
	"break",
	"continue",
	"return",
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
	fmt.Printf("IF is not yet supported [%d-%d]\n", p.i, p.n)
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

	opEq := false
	opNe := false
	opGe := false
	opGt := false
	opLt := false
	opLe := false
	p.skipWhitespace()
	sym1 := p.peekChar()
	sym2 := p.peekAhead(1)
	if (sym1 == '(') {
		return fmt.Errorf("IF ( not yet implemented")
	} else if (sym1 == '=') && (sym2 == '=') {
		opEq = true
		p.skip(2)
	} else if (sym1 == '!') && (sym2 == '=') {
		opNe = true
		p.skip(2)
	} else if (sym1 == '>') && (sym2 == '=') {
		opGe = true
		p.skip(2)
	} else if (sym1 == '<') && (sym2 == '=') {
		opLe = true
		p.skip(2)
	} else if (sym1 == '>') {
		opGt = true
		p.skip(2)
	} else if (sym1 == '<') {
		opLt = true
		p.skip(2)
	} else {
		return fmt.Errorf("IF %c%c is an unknown syntax", sym1, sym2)
	}

	hasValue := false
	var value int
	var err error
	ifSz := R08
	p.skipWhitespace()
	if (p.peekChar() != '{') {
		// Constant
		if (p.isNextAZ()) {
			symbol := p.nextAZ_az_09()
			value, err = p.lookupConstant(symbol)
			if (err != nil) {
				return fmt.Errorf("invalid constant '%s' in IF", symbol)
			}
		// Value
		} else {
			value, err = p.nextValue()
			if (err != nil) {
				return fmt.Errorf("invalid value in IF")
			}
		}
		
		hasValue = true
		if (value > 0x0FFFFFF) {
			fmt.Errorf("IF %d does not fit into 24-bits", value)
		} else if (value > 0x0FFFF) {
			ifSz = R24
		} else if (value > 0x0FF) {
			ifSz = R16
		}
	}

	p.skipWhitespace()
	if (p.nextChar() != '{') {
		return fmt.Errorf("missing { in IF")
	}
	p.skipWhitespaceAndEOL()

	// Name this construct
	name := fmt.Sprintf("IF%x", p.currentCode.endAddr)

	// Add the instruction with the sub in the current block (before starting a new block)
	opStr := "=="
	if (opEq) { opStr = "=="
	} else if (opNe) { opStr = "!="
	} else if (opGe) { opStr = ">="
	} else if (opGt) { opStr = ">"
	} else if (opLt) { opStr = "<"
	} else if (opLe) { opStr = "<=" }
	valStr := ""
	if (hasValue) { valStr = fmt.Sprintf(" %d", value)}
	comment := fmt.Sprintf("IF %s%s {", opStr, valStr)
	p.addKeywordInstruction(sub, comment)

	// Add the code for the IF
	b := p.addCodeBlock(sub, "IF", name, false);

	// Have to assume the branch is far, but not beyond a signed 16-bits
	if (hasValue) {
		p.addExprInstruction("cmp", modeImmediate, ifSz, value)
	}
	var skipInstr *instruction
	if (opEq) {
		skipInstr = p.addExprInstructionWithSymbol("bne", modeRelative, A16, 0, name + "_end", false)
	} else if (opNe) {
		skipInstr = p.addExprInstructionWithSymbol("beq", modeRelative, A16, 0, name + "_end", false)
	} else if (opGe) {
		skipInstr = p.addExprInstructionWithSymbol("bcc", modeRelative, A16, 0, name + "_end", false)
	} else if (opGt) {
		skipInstr = p.addExprInstructionWithSymbol("beq", modeRelative, A16, 0, name + "_end", false)
		skipInstr = p.addExprInstructionWithSymbol("bcc", modeRelative, A16, 0, name + "_end", false)
	} else if (opLt) {
		skipInstr = p.addExprInstructionWithSymbol("bcs", modeRelative, A16, 0, name + "_end", false)
	} else if (opLe) {
		skipInstr = p.addExprInstructionWithSymbol("beq", modeRelative, A16, 0, name + "_if", false)
		skipInstr = p.addExprInstructionWithSymbol("bcs", modeRelative, A16, 0, name + "_end", false)
		p.addInstructionLabel(b.name + "_if")
	}

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
		p.addExprInstructionWithSymbol("bra", modeRelative, A16, 0, name + "_end", false)

		// Add a label to where the else goes, and change the NOT IF branch to here
		skipInstr.symbol = name + "_else"
		skipInstr.symbolLC = strings.ToLower(skipInstr.symbol)
		p.addInstructionLabel(b.name + "_else")

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

	p.addInstructionLabel(name + "_loop")

	// Parse the code
	err := p.parseCode(name)
	if (err != nil) {
		return err
	}

	// How far back is the top of the loop?
	target, err := p.currentCode.lookupInstructionLabel(name + "_loop")
	distance := p.currentCode.endAddr - target + 4
	if (distance < 0x7FFF) {
		loopSz := A16
		if (distance > 0x07F) {
			loopSz = A24
		}
		p.addExprInstructionWithSymbol("bra", modeRelative, loopSz, 0, name + "_loop", false)
	} else {
		p.addExprInstructionWithSymbol("jmp", modeAbsolute, A24, 0, name + "_loop", false)
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

	v := strings.ToLower(p.nextAZ_az_09())
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
		sub.upDown = false
		to = strings.ToUpper(p.nextAZ_az_09())
	} else {
		sub.upDown = true
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
	p.skipWhitespaceAndEOL()

	// Name this construct
	name := fmt.Sprintf("FOR%x", p.currentCode.endAddr)

	// Add the instruction with the sub in the current block (before starting a new block)
	down := "DOWN "
	if (sub.upDown == true) { down = "" }
	comment := fmt.Sprintf("FOR %s = %d %sTO %d {", v, start, down, end)
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

	if (v == "a") {
		fmt.Errorf("Bad form to FOR A, try X or Y or a variable")
	} else if (v == "x") {
		p.addExprInstruction("ldx", modeImmediate, loopSz, start)
	} else if (v == "y") {
		p.addExprInstruction("ldy", modeImmediate, loopSz, start)
	} else if (v[0:1] == "@") {
		// @@@ TODO - ADD LOOP VARIABLES
	}
	p.addInstructionLabel(name + "_loop")

	// Parse the code
	err = p.parseCode(name)
	if (err != nil) {
		return err
	}

	if (v == "x") {
		p.addExprInstruction("inx", modeImplicit, loopSz, 0)
	} else if (v == "y") {
		p.addExprInstruction("iny", modeImplicit, loopSz, 0)
	} else if (v[0:1] == "@") {
		// @@@ TODO - ADD LOOP VARIABLES
	}
	p.addExprInstruction("cmp", modeImmediate, loopSz, end)
	p.addExprInstructionWithSymbol("bne", modeRelative, loopSz, 0, name + "_loop", false)

	// Add a label to the end of the block
	p.addInstructionLabel(b.name + "_end")

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

	label := strings.ToLower(loop.name + "_start")
	target, _ := p.currentCode.lookupInstructionLabel(label)
	distance := p.currentCode.endAddr - target + 4
	if (distance < 0x7FFF) {
		loopSz := A16
		if (distance > 0x07F) {
			loopSz = A24
		}
		p.addExprInstructionWithSymbol("bra", modeRelative, loopSz, 0, label, false)
	} else {
		p.addExprInstructionWithSymbol("jmp", modeAbsolute, A24, 0, label, false)
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
