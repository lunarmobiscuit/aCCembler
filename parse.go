package pomme

import (
	"errors"
	"fmt"
    "strings"
)

/*
 *  Parser
 */
func (p *parser) parseFile(filename string, buffer []uint8) error {
	fmt.Printf("PARSE %s\n", filename)

	// (Re)Initialize the parser buffer and counts
	p.b = buffer
	p.end = len(buffer)-1
	p.i = 0
	p.n = 1

	// Loop until there are no more top-level blocks in the file
	for p.i < p.end {

		// Skip until the next parsable character
		p.skipWhitespaceAndEOL()

		// Next alphanumeric token
		token := strings.ToLower(p.nextAZ_az_09())

		// No token, check for comments
		if (token == "") {
			sym := p.peekChar()
			if (sym == '#') {
				p.skip(1)
				err := p.parseHashcode()
				if (err != nil) {
					fmt.Printf("ERROR in %s [line %d] -- %s\n", filename, p.n, err)
					return err
				}
				continue
			} else if (p.skipComment()) {
				continue
			} else {
				err := errors.New("expected @addr or #command or sub")
				fmt.Printf("ERROR in %s [line %d] -- %s\n", filename, p.n, err)
				return err
			}
		}
		
		// Check for valid top-level keywords
		switch (token) {
		case "const":
			var label string
			label = p.nextAZ_az_09()
			err := p.parseConstant(label)
			if (err != nil) {
				fmt.Printf("ERROR in %s [line %d] -- %s\n", filename, p.n, err)
				return err
			}
		case "global":
			var label string
			label = p.nextAZ_az_09()
			err := p.parseVariable(nil, label)
			if (err != nil) {
				fmt.Printf("ERROR in %s [line %d] -- %s\n", filename, p.n, err)
				return err
			}
		case "sub":
			var label string
			label = p.nextAZ_az_09()
			err := p.parseSubroutineBlock(label)
			if (err != nil) {
				fmt.Printf("ERROR in %s [line %d] -- %s\n", filename, p.n, err)
				return err
			}
		case "data":
			var label string
			label = p.nextAZ_az_09()
			err := p.parseDataBlock(label)
			if (err != nil) {
				fmt.Printf("ERROR in %s [line %d] -- %s\n", filename, p.n, err)
				return err
			}
		case "default":
			err := errors.New("expected asm or sub")
			fmt.Printf("ERROR in %s [line %d] -- %s\n", filename, p.n, err)
			return err
		}
	}

	return nil
}


/*
 *  Parse the hashcode directive
 */
func (p *parser) parseHashcode() error {
	// Which hashcode is it?
	hashcode := strings.ToLower(p.nextAZ_az_09())
	p.skipWhitespace()

	switch (hashcode) {
	case "include":
		// Parse the filename, e.g. #include "foo.pom"
		if (p.peekChar() != '"') {
			return fmt.Errorf("#include is missing the opening \"\n")
		}
		p.skip(1)
		filename := p.untilQuote()
		if (p.peekChar() != '"') {
			return fmt.Errorf("#include is missing the closing \"\n")
		}
		p.skip(1)

		// Load the file
		buffer, err := readFile(filename)
		if err != nil {
			fmt.Printf("ERROR trying to include %s -- %s\n", filename, err)
			return err
		}

		// Remember where we left off
		saveB := p.b
		saveEnd := p.end
		saveI := p.i
		saveN := p.n

		// Parse the file
		err = p.parseFile(filename, buffer)
		if (err != nil) {
			return err
		}

		// Restore where we left off
		p.b = saveB
		p.end = saveEnd
		p.i = saveI
		p.n = saveN

		return nil
	}


	return fmt.Errorf("#%s is not a valid compiler directive", hashcode)
}

/*
 *  Parse a defined constant
 */
func (p *parser) parseConstant(label string) error {
	if (label == "") {
		return errors.New("const is missing a name")
	}

	// all constants are stored as lowercase
	labelLC := strings.ToLower(label)

	// Skip past whitespace
	p.skipWhitespace()

	// Check for duplicate
	_, err := p.lookupConstant(labelLC)
	if (err == nil) {
		return fmt.Errorf("const '%s' is already defined", label)
	}

	// = value
	if (p.peekChar() != '=') {
		return errors.New("const is missing a '='")
	}
	p.skip(1)

	var value int
	if p.isNextAZ() {
		ref := p.nextAZ_az_09()
		value, err = p.lookupConstant(ref)
		if (err != nil) {
			return fmt.Errorf("const '%s' references '%s' which is not defined", label, ref)
		}
	} else {
		var err error
		value, err = p.nextValue()
		if (err != nil) {
			return fmt.Errorf("const '%s' does not specify a value", label)
		}
	}

	// Skip until the next parsable character
	p.skipWhitespaceAndEOL()

	// Store this constant
	cnst := new(cnst)
	if p.cnst == nil {
		p.cnst = cnst
	} else if p.lastCnst != nil {
		p.lastCnst.next = cnst
	}
	p.lastCnst = cnst
	cnst.next = nil
	cnst.name = label
	cnst.nameLC = labelLC
	cnst.value = value

	return nil
}

/*
 *  Parse a defined variable
 */
func (p *parser) parseVariable(b *codeBlock, name string) error {
	var keyword string
	if (b == nil) {
		keyword = "global"
	} else {
		keyword = "var"
	}

	if (name == "") {
		return fmt.Errorf("%s is missing a name", keyword)
	}

	// all variable names are stored as lowercase
	nameLC := strings.ToLower(name)

	// Skip past whitespace
	p.skipWhitespace()

	// Check for duplicate
	_, err := p.lookupConstant(nameLC)
	if (err == nil) {
		return fmt.Errorf("%s '%s' is already defined as the name of a constant", keyword, name)
	}
	_, _, err = p.lookupVariable(b, nameLC)
	if (err == nil) {
		return fmt.Errorf("%s '%s' is already defined as another variable", keyword, name)
	}

	// = value
	if (p.peekChar() != '=') {
		return fmt.Errorf("%s '%s' is missing a '='", keyword, name)
	}
	p.skip(1)

	p.skipWhitespace()
	if (p.peekChar() != '@') {
		return fmt.Errorf("%s '%s' is missing a '@'", keyword, name)
	}
	p.skip(1)

	// M@$aaaa
	var address int
	if (p.peekChar() == '$') { // M@$aaaa
		address, err = p.nextValue()
		if (err != nil) {
			return fmt.Errorf("%s '%s' specifies an invalid address", keyword, name)
		}
	} else {
		// M@label  ; label = contstant or variable
		token := p.nextAZ_az_09()
		address, err = p.lookupConstant(token)
		if (err != nil) {
			address, _, err = p.lookupVariable(b, token)
			if (err != nil) {
				return fmt.Errorf("%s '%s' specifies an unknown variable or constant '%s'", keyword, name, token)
			}
		}
	}
	
	// Optional size
	size := p.parseOpWidth()

	// Skip until the next parsable character
	p.skipWhitespaceAndEOL()

	// Store this variable (either as a global or local to a block)
	vrbl := new(vrbl)
	if (b == nil) {
		if p.global == nil {
			p.global = vrbl
		} else if p.lastGlobal != nil {
			p.lastGlobal.next = vrbl
		}
		p.lastGlobal = vrbl
	} else {
		if b.vrbl == nil {
			b.vrbl = vrbl
		} else if b.lastVrbl != nil {
			b.lastVrbl.next = vrbl
		}
		b.lastVrbl = vrbl
	}
	vrbl.next = nil
	vrbl.name = name
	vrbl.nameLC = nameLC
	vrbl.address = address
	vrbl.size = size

	return nil
}

/*
 *  Parse a (optionally named) block of assembly
 */
func (p *parser) parseSubroutineBlock(label string) error {
	if (label == "") {
		return errors.New("sub is missing a name")
	}

	// Skip past whitespace
	p.skipWhitespace()

	// Optional @ADDR
	address := 0
	if (p.peekChar() == '@') {
		p.skip(1)
		var err error
		address, err = p.nextValue()
		if (err != nil) {
			return errors.New("'@'' does not specify a value")
		}
		p.skipWhitespace()
	} else {
		address = p.endestAddr()
	}

	// Set the default width based on the address of this block (@@@ not 100% correct but a good first guess)
	if (address <= 0xFFFF) {
		p.abWidth = A16
	} else {
		p.abWidth = A24
	}

	// Must have a { before the end of the line
	if (p.peekChar() != '{') {
		return errors.New("expected {")
	}

	// Skip until the next parsable character
	p.skipWhitespaceAndEOL()

	// Store this code block
	block := new(codeBlock)
	if p.code == nil {
		p.code = block
		block.prev = nil
	} else if p.lastCode != nil {
		block.prev = p.lastCode
		p.lastCode.next = block
	}
	p.lastCode = block
	p.currentCode = block
	block.next = nil
	block.startAddr = address
	block.endAddr = address
	block.name = label
	block.nameLC = strings.ToLower(label)
	block.instr = nil

	// Parse the code
	p.skip(1)
	return p.parseCode(label)
}

/*
 *  Parse the block of code
 */
func (p *parser) parseCode(label string) error {
	// Should be a sequence of mnemonics, keywords, and label, followed by a '}'
	var token string
	for p.i < p.end {
		token = strings.ToLower(p.nextAZ_az_09())

		// Not a AZ09 symbol, so is it a blank line or comment or variable or syntax error?
		if (token == "") {
			if (p.skipComment()) {
				continue
			} else if p.peekChar() == '}' {	// end of the block
				p.nextLine()
				return nil
			} else if p.peekChar() == '@' {	// must be start of a variable in an expression
				err := p.parseExpression(token)
				if (err != nil) {
					return err
				}
				continue
			} else {
				return fmt.Errorf("found '%c' instead of valid mnemonic or keyword in {...}", p.peekChar())
				break
			}
		}

		// Parse the keyword
		if (token == "a") || (token == "x") || (token == "y") || (token == "m") {
			err := p.parseExpression(token)
			if (err != nil) {
				return err
			}
		} else if p.isRegisterMnemonic(token) {
			err := p.parseRegister(token)
			if (err != nil) {
				return err
			}
		} else if p.isKeyword(token) {
			err := p.parseKeyword(token)
			if (err != nil) {
				return err
			}
		} else if p.isMnemonic(token) {
			err := p.parseMnemonic(token)
			if (err != nil) {
				return err
			}
		} else if p.peekChar() == ':' {
			err := p.parseLabel(token)
			if (err != nil) {
				return err
			}
		} else {
			return fmt.Errorf("'%s' is an unknown keyword/mnemonic/label in '%s'", token, label)
		}
	}

	return errors.New("unexpected end of file within {...}")
}

/*
 *  Parse a block of data
 */
func (p *parser) parseDataBlock(label string) error {
	if (label == "") {
		return errors.New("data is missing a name")
	}

	// Skip past whitespace
	p.skipWhitespace()

	// Optional @ADDR
	address := 0
	if (p.peekChar() == '@') {
		p.skip(1)
		var err error
		address, err = p.nextValue()
		if (err != nil) {
			return errors.New("'@'' does not specify a value")
		}

		// Skip past whitespace
		p.skipWhitespace()
	} else {
		address = p.endestAddr()
	}

	// Description of the data size comes next
	token := strings.ToLower(p.nextAZ_az_09())
	size := R08
	if (token == "") {
		return errors.New("missing data size (e.g. byte, u8, word, u16, trip, u24, str)")
	} else {
		switch (token) {
		case "byte", "u8":
			size = R08
		case "word", "u16":
			size = R16
		case "trip", "u24":
			size = R24
		case "str", "string":
			size = DSTRING
		default:
			return fmt.Errorf("invalid data size '%s', expecting byte, u8, word, u16, trip, u24, str", token)
		}
	}

	// Skip past whitespace
	p.skipWhitespace()

	// Must have a { before the end of the line
	if (p.peekChar() != '{') {
		return errors.New("expected {")
	}
	p.skip(1)

	// Skip until the next parsable character
	p.skipWhitespaceAndEOL()

	// Store this data block
	block := new(dataBlock)
	if p.data == nil {
		p.data = block
		block.prev = nil
	} else if p.lastData != nil {
		block.prev = p.lastData
		p.lastData.next = block
	}
	p.lastData = block
	block.next = nil
	block.startAddr = address
	block.endAddr = address
	block.name = label
	block.nameLC = strings.ToLower(label)
	block.data = nil

	// Parse the data
	err := p.parseData(size, label, block)
	if (err != nil) {
		return err
	}

	return nil
}

/*
 *  Parse the block of data
 */
func (p *parser) parseData(size int, label string, block *dataBlock) error {
	// Should be a sequence of values followed by a '}'
	var token string
	for p.i < p.end {
		// Skip past whitespace
		p.skipWhitespace()

		token = strings.ToLower(p.nextAZ_az_09())

		// Not a AZ09 symbol, so is it a blank line or comment or syntax error?
		if (token == "") {
			if (p.skipComment()) {
				continue
			} else if p.peekChar() == ',' {	// next item
				p.skip(1)
				continue
			} else if p.peekChar() == '}' {	// end of the block
				p.nextLine()
				return nil
			} else if size == DSTRING {
				if p.peekChar() == '"' {
					p.skip(1)
					str := p.untilQuote()
					block.addData(DSTRING, 0, str, len(str)+1)
					p.skip(1)
				} else {
					return fmt.Errorf("was expecting quoted string in '%s'", label)
				}
			} else {
				val, err := p.nextValue()
				if (err != nil) {
					return fmt.Errorf("was expecting a numeric value in '%s'", label)
				}
				switch size {
				default:
					if val > 0x0FF {
						return fmt.Errorf("%d is bigger than 8-bits (in '%s')", val, label)
					}
					block.addData(R08, val, "", 1)
				case R16:
					if val > 0x0FFFF {
						return fmt.Errorf("%d is bigger than 16-bits (in '%s')", val, label)
					}
					block.addData(R16, val, "", 2)
				case R24:
					if val > 0x0FFFFFF {
						return fmt.Errorf("%d is bigger than 24-bits (in '%s')", val, label)
					}
					block.addData(R24, val, "", 3)
				}
			}
		} else {
			// Alphanumeric value but no quotes, and constants can't be strings
			if size == DSTRING {
				return fmt.Errorf("'%s' is an missing quotes in '%s'", token, label)
			}

			// Lookup value as contant, subroutine, or data block name
			val, err := p.lookupConstant(token)
			if (err != nil) {
				sub := p.lookupSubroutineName(token)
				if (sub != nil) {
					val = sub.startAddr
				} else {
					data := p.lookupDataName(token)
					if (data != nil) {
						val = data.startAddr
					} else {
						return fmt.Errorf("'%s' is an unknown data value in '%s'", token, label)
					}
				}
			}

			switch size {
			default:
				if val > 0x0FF {
					return fmt.Errorf("%d is bigger than 8-bits (in '%s')", val, label)
				}
				block.addData(R08, val, "", 1)
			case R16:
				if val > 0x0FFFF {
					return fmt.Errorf("%d is bigger than 16-bits (in '%s')", val, label)
				}
				block.addData(R16, val, "", 2)
			case R24:
				if val > 0x0FFFFFF {
					return fmt.Errorf("%d is bigger than 24-bits (in '%s')", val, label)
				}
				block.addData(R24, val, "", 3)
			}

		}
	}

	return errors.New("unexpected end of file within {...}")
}

/*
 *  Add the opcode to the (latest) block of code
 */
func (b *dataBlock) addData(size int, value int, str string, len int) {
	data := new(data)
	if (b.data == nil) {
		b.data = data
		data.prev = nil
	} else if b.lastData != nil {
		data.prev = b.lastData
		b.lastData.next = data
	}
	b.lastData = data
	data.next = nil
	data.size = size
	data.value = value
	data.string = str
	data.len = len
	b.endAddr += len

	if (data.prev == nil) {
		data.address = b.startAddr
	} else {
		data.address = data.prev.address + data.prev.len
		b.endAddr = data.address + len
	}
}


/*
 *  Return the highest end address
 */
func (p *parser) endestAddr() int {
	if (p.lastData == nil) {
		return p.lastCode.endAddr
	} else if (p.lastCode == nil) {
		return p.lastData.endAddr
	} else if (p.lastCode.endAddr > p.lastData.endAddr) {
		return p.lastCode.endAddr
	} else {
		return p.lastData.endAddr
	}
}
