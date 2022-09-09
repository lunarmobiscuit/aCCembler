package aCCembler

import (
	//"errors"
	"fmt"
	"os"
)

/*
 *  Machine code generator
 */
func (p *parser) generateCode(out *os.File, listing *os.File) error {
	fmt.Printf("GENERATE CODE\n")

	// Resolve forward referenes
	err := p.resolveSymbols()
	if (err != nil) {
		return err
	}

	// Check for overlapping addresses
	err = p.checkAddressRanges()
	if (err != nil) {
		return err
	}

	// Generate the machine code and listing
	err = p.outputCode(out, listing)
	if (err != nil) {
		return err
	}

	fmt.Printf("+ %6d bytes ($%x) of DATA\n", p.dataSize, p.dataSize)
	fmt.Printf("+ %6d bytes ($%x) in TOTAL\n", p.codeSize + p.dataSize, p.codeSize + p.dataSize)
	fmt.Printf("+ %6d bytes ($%x) of FILLER\n", p.fillerSize, p.fillerSize)
	fmt.Printf("+ %6d bytes ($%x) in the OUTPUT file\n", p.codeSize + p.dataSize + p.fillerSize, p.codeSize + p.dataSize + p.fillerSize)
	fmt.Printf("\n")

	return nil
}

/*
 *  Resolve any symbols that were forward references
 */
func (p *parser) resolveSymbols() error {
	// Loop through all the code blocks
	for b := p.code; b != nil; b = b.next {
		err := p.resolveCodeSymbols(b)
		if (err != nil) {
			return err
		}
	}

	return nil
}

/*
 *  Resolve any symbols that were forward references
 */
func (p *parser) resolveCodeSymbols(b *codeBlock) error {
	// Loop through all the instructions
	for i := b.instr; i != nil; i = i.next {
		// Dive into sub-blocks
		if (i.subBlock != nil) {
			err := p.resolveCodeSymbols(i.subBlock.block)
			if (err != nil) {
				return err
			}

			continue
		}

		// Skip labels
		if (i.mnemonic == 0) {
			continue
		}

		// Skip comments and expressions
		if (i.comment != nil) || (i.expr != nil) {
			continue
		}

		// Skip implicit
		if (i.addressMode == modeImplicit) {
			continue
		}

		// Skip XY modes
		if (i.addressMode == modeX) || (i.addressMode == modeXY) {
			continue
		}

		// Compute branches
		if (i.addressMode == modeRelative) && (i.hasValue == false) {
			targetAddr, err := b.lookupInstructionLabel(i.symbol)
			if (err == nil) {
				i.hasValue = true
				diff := targetAddr - (i.address + i.len)
				if (i.prefix == A16) && ((diff < -128) || (diff > 127)) {
					return fmt.Errorf("%s target %s is %d bytes apart, too far for 8-bit branch", mnemonics[i.mnemonic].name, i.symbol, diff)
				} else if (i.prefix == A24) && ((diff < -32768) || (diff > 32767)) {
					return fmt.Errorf("%s target %s is %d bytes apart,  too far for 16-bit branch", mnemonics[i.mnemonic].name, i.symbol, diff)
				}
				i.value = diff
			}
		}

		// Find the matching label
		if i.hasValue == false {
			v, err := b.lookupInstructionLabel(i.symbol)
			if (err == nil) {
				i.hasValue = true
				i.value = v + i.value
			} else {
				s := p.lookupSubroutineName(i.symbol)
				if (s != nil) {
					i.hasValue = true
					i.value = s.startAddr
				} else {
					d := p.lookupDataName(i.symbol)
					if (d != nil) {
						i.hasValue = true
						i.value = d.startAddr + i.value
					} else {
						return fmt.Errorf("*** %s is an unknown symbol", i.symbol)
					}
				}
			}

			// Modify the address mode if the length is shorter or longer than expected
			if (i.addressMode == modeImmediate) {
				i.prefix |= valueToPrefix(i.value)
				if (i.prefix & R32 == R32) {
					return fmt.Errorf("the symbol '%s' resolved to a value too large for #%d", i.symbol, i.value)
				}
			}


			// The symbol was not resolved
			if i.hasValue == false {
				// the symbol was not found
				return fmt.Errorf("the symbol '%s' was not found", i.symbol)
			}
		}
	}

	return nil
}


/*
 *  Check to ensure the address ranges of the subroutines and data blocks do not overlap
 */
func (p *parser) checkAddressRanges() error {
	for b := p.code; b != nil; b = b.next {
		fmt.Printf("  SUB  @$%06x-$%06x  '%s'\n", b.startAddr, b.endAddr, b.name)
	}
	for d := p.data; d != nil; d = d.next {
		fmt.Printf("  DATA @$%06x-$%06x  '%s'\n", d.startAddr, d.endAddr, d.name)
	}

	// Loop through all the subroutine blocks looking for overlaps
	for b := p.code; b != nil; b = b.next {
		for c := b.next; c != nil; c = c.next {
			if (c.startAddr < b.endAddr) && (c.endAddr > b.startAddr) {
				return fmt.Errorf("SUB '%s' @$%06x-$%06x overlaps addresses with SUB '%s' @$%06x-$%06x",
					b.name, b.startAddr, b.endAddr, c.name, c.startAddr, c.endAddr)
			}
		}
		for d := p.data; d != nil; d = d.next {
			if (d.startAddr < b.endAddr) && (d.endAddr > b.startAddr) {
				return fmt.Errorf("SUB '%s' @$%06x-$%06x overlaps addresses with DATA '%s' @$%06x-$%06x",
					b.name, b.startAddr, b.endAddr, d.name, d.startAddr, d.endAddr)
			}
		}
	}

	// Loop through all the data blocks looking for overlaps
	for d := p.data; d != nil; d = d.next {
		for e := d.next; e != nil; e = e.next {
			if (e.startAddr < d.endAddr) && (e.endAddr > d.startAddr) {
				return fmt.Errorf("DATA '%s' @$%06x-$%06x overlaps addresses with DATA '%s' @$%06x-$%06x",
					d.name, d.startAddr, d.endAddr, e.name, e.startAddr, e.endAddr)
			}
		}
	}

	// Loop through all the code and data blocks, ensuring they are in ascending order of addresses
	b := p.code
	d := p.data
	lastStartAddr := 0
	lastEndAddr := 0
	for (b != nil) || (d != nil) {
		if (b != nil) && ((d == nil) || (b.startAddr < d.startAddr)) {
			if (b.startAddr < lastEndAddr) {
				return fmt.Errorf("SUB '%s' @$%06x-$%06x is specified after @$%06x-$%06x and there is no auto sort",
					d.name, d.startAddr, d.endAddr, lastStartAddr, lastEndAddr)
			}

			lastStartAddr = b.startAddr
			lastEndAddr = b.endAddr
			b = b.next
		} else if (d != nil) && ((b == nil) || (d.startAddr < b.startAddr)) {
			if (d.startAddr < lastEndAddr) {
				return fmt.Errorf("DATA '%s' @$%06x-$%06x is specified after @$%06x-$%06x and there is no auto sort",
					d.name, d.startAddr, d.endAddr, lastStartAddr, lastEndAddr)
			}

			lastStartAddr = d.startAddr
			lastEndAddr = d.endAddr
			d = d.next
		} else if (b == nil) {
			return fmt.Errorf("DATA '%s' @$%06x-$%06x is specified after @$%06x-$%06x and there is no auto sort",
				d.name, d.startAddr, d.endAddr, lastStartAddr, lastEndAddr)
		} else {
			return fmt.Errorf("SUB '%s' @$%06x-$%06x is specified after @$%06x-$%06x and there is no auto sort",
				d.name, d.startAddr, d.endAddr, lastStartAddr, lastEndAddr)
		}
	}

	// Make sure the lowest address is code, not data, unless there is no code
	if (p.code != nil) {
		if (p.data != nil) {
			if (p.data.startAddr < p.code.startAddr) {
				return fmt.Errorf("The lowest address must be code SUB '%s' @$%06x-$%06x rather than DATA '%s' @$%06x-$%06x and there is no auto sort",
					p.code.name, p.code.startAddr, p.code.endAddr,
					p.data.name, p.data.startAddr, p.data.endAddr)
			}
		}
	}

	fmt.Printf("\n")
	return nil
}


/*
 *  Output the code and data
 */
func (p *parser) outputCode(out *os.File, listing *os.File) error {
	// Dump the global variables at the top
	for v := p.global; v != nil; v = v.next {
		size := "       "
		switch (v.size & R32) {
		case R16: size = fmt.Sprintf("-%06x", v.address+1)
		case R24: size = fmt.Sprintf("-%06x",  v.address+2)
		}
		variable := fmt.Sprintf("%06x%s   ; GLOBAL @%s\n", v.address, size, v.name)
		listing.WriteString(variable)
		//fmt.Printf(variable)

	}

	// Walk through two linked lists in ascending order of addresses
	b := p.code
	d := p.data
	lastEndAddr := 0

	// The data can't come first unless there is no code (already been checked)
	if (b == nil) {
		lastEndAddr = d.endAddr
	} else {
		lastEndAddr = b.endAddr
	}

	// Loop through all the code and data blocks, in ascending order of addresses
	for (b != nil) || (d != nil) {
		// The next block is code
		if (b != nil) && ((d == nil) || (b.startAddr < d.startAddr)) {
			// Fill any gaps between blocks
			if (b.startAddr > lastEndAddr) {
				p.outputFiller(b.startAddr - lastEndAddr, out, listing)
				p.fillerSize += b.startAddr - lastEndAddr
			}
			lastEndAddr = b.endAddr

			sub := fmt.Sprintf("\n%06x ; SUB %s:\n", b.startAddr, b.name)
			listing.WriteString(sub)
			//fmt.Printf(sub)

			err := p.outputCodeBlock(b, out, listing)
			if (err != nil) {
				return err
			}

			p.codeSize += b.endAddr - b.startAddr
			b = b.next
		} else {
			// Fill any gaps between blocks
			if (d.startAddr > lastEndAddr) {
				p.outputFiller(d.startAddr - lastEndAddr, out, listing)
				p.fillerSize += d.startAddr - lastEndAddr
			}
			lastEndAddr = d.endAddr

			data := fmt.Sprintf("\n%06x ; DATA %s:\n", d.startAddr, d.name)
			listing.WriteString(data)
			//fmt.Printf(data)

			err := p.outputDataBlock(d, out, listing)
			if (err != nil) {
				return err
			}

			p.dataSize += d.endAddr - d.startAddr
			d = d.next
		}
	}

	return nil
}

/*
 *  Output the code block
 */
func (p *parser) outputCodeBlock(b *codeBlock, out *os.File, listing *os.File) error {
	// Dump the local variables
	for v := b.vrbl; v != nil; v = v.next {
		size := "       "
		switch (v.size & R32) {
		case R16: size = fmt.Sprintf("-%06x", v.address+1)
		case R24: size = fmt.Sprintf("-%06x",  v.address+2)
		}
		variable := fmt.Sprintf("%06x%s   ; VAR @%s\n", v.address, size, v.name)
		listing.WriteString(variable)
		//fmt.Printf(variable)
	}

	// Loop through all the instructions
	for i := b.instr; i != nil; i = i.next {
		line := fmt.Sprintf("%06x ", i.address)
		bytes := make([]byte, 16)
		byteIdx := 0

		// IF/FOR/DO/ETC
		if (i.subBlock != nil) {
			c := i.comment
			line += fmt.Sprintf("                                ;; %s", c.comment)
		// Comments
		} else if (i.comment != nil) {
			c := i.comment
			line += fmt.Sprintf("                                ;; %s", c.comment)
		// Expression
		} else if (i.expr != nil) {
			// Format the expression
			e := i.expr
			line += "                      ;; "
			switch (e.dest.location) {
			case MEMORY: line += fmt.Sprintf("M@$%0x", e.dest.addrval)
			case VARIABLE: line += fmt.Sprintf("V@$%0x", e.dest.addrval)
			case VALUE: return fmt.Errorf("the left side of an expression can not be a value (%x)", e.dest.addrval)
			case REG_A: line += fmt.Sprintf("A")
			case REG_X: line += fmt.Sprintf("X")
			case REG_Y: line += fmt.Sprintf("Y")
			}
			line += fmt.Sprintf("%s ", sizeToSuffix(e.dest.size))
			switch (e.equalOp) {
			case EQUALS: line += fmt.Sprintf("= ")
			case PLUS: line += fmt.Sprintf("+= ")
			case MINUS: line += fmt.Sprintf("-= ")
			case AND: line += fmt.Sprintf("&= ")
			case OR: line += fmt.Sprintf("|= ")
			case EOR: line += fmt.Sprintf("^= ")
			case SHIFT_LEFT: line += fmt.Sprintf("<<= ")
			case SHIFT_RIGHT: line += fmt.Sprintf(">>= ")
			}
			switch (e.src1.location) {
			case MEMORY: line += fmt.Sprintf("M@$%0x", e.src1.addrval)
			case VARIABLE: line += fmt.Sprintf("V@$%0x", e.src1.addrval)
			case VALUE: line += fmt.Sprintf("%d", e.src1.addrval)
			case REG_A: line += fmt.Sprintf("A")
			case REG_X: line += fmt.Sprintf("X")
			case REG_Y: line += fmt.Sprintf("Y")
			}
			line += fmt.Sprintf("%s ", sizeToSuffix(e.src1.size))
			if (e.op != NO_OP) {
				switch (e.op) {
				case PLUS: line += fmt.Sprintf("+ ")
				case MINUS: line += fmt.Sprintf("- ")
				case AND: line += fmt.Sprintf("& ")
				case OR: line += fmt.Sprintf("| ")
				case EOR: line += fmt.Sprintf("^ ")
				case SHIFT_LEFT: line += fmt.Sprintf("<< ")
				case SHIFT_RIGHT: line += fmt.Sprintf("<< ")
				}
				switch (e.src2.location) {
				case MEMORY: line += fmt.Sprintf("M@$%0x", e.src2.addrval)
				case VARIABLE: line += fmt.Sprintf("V@$%0x", e.src2.addrval)
				case VALUE: line += fmt.Sprintf("%d", e.src2.addrval)
				case REG_A: line += fmt.Sprintf("A")
				case REG_X: line += fmt.Sprintf("X")
				case REG_Y: line += fmt.Sprintf("Y")
				}
				line += fmt.Sprintf("%s ", sizeToSuffix(e.src2.size))
			}

			line += "\n"
			listing.WriteString(line)
			//fmt.Printf(line)
			continue
		} else if (i.mnemonic == 0) {
		// Label
			line += fmt.Sprintf("                %s:\n", i.symbol)

			listing.WriteString(line)
			//fmt.Printf(line)
			continue
		} else {
		// Mnemonic
			// Prefix code
			opcodes := ""
			length := i.len
			if (i.prefix > A16) {
				opcodes += fmt.Sprintf("%02x ", i.prefix)
				bytes[byteIdx] = byte(i.prefix & 0xff); byteIdx += 1;
				length -= 1
			}
			// Machine opcode
			opcodes += fmt.Sprintf("%02x ", i.opcode)
			bytes[byteIdx] = byte(i.opcode & 0xff); byteIdx += 1;
			length -= 1

			// Value
			if length > 3 {
				return fmt.Errorf("*** the length for %s is %d, too long", mnemonics[i.mnemonic].name, i.len)
			} else if length >= 3 {
				opcodes += fmt.Sprintf("%02x %02x %02x ",
					i.value & 0xff, (i.value >> 8) & 0xff, (i.value >> 16) & 0xff)
				bytes[byteIdx] = byte(i.value & 0xff); byteIdx += 1;
				bytes[byteIdx] = byte((i.value >> 8) & 0xff); byteIdx += 1;
				bytes[byteIdx] = byte((i.value >> 16) & 0xff); byteIdx += 1;
			} else if length == 2 {
				opcodes += fmt.Sprintf("%02x %02x ", i.value & 0xff, (i.value >> 8) & 0xff)
				bytes[byteIdx] = byte(i.value & 0xff); byteIdx += 1;
				bytes[byteIdx] = byte((i.value >> 8) & 0xff); byteIdx += 1;
			} else if length == 1 {
				opcodes += fmt.Sprintf("%02x ", i.value & 0xff)
				bytes[byteIdx] = byte(i.value & 0xff); byteIdx += 1;
			}

			// Assembly code mneumonic
			spaces := "                                        "
			line += fmt.Sprintf("%s%s %s", opcodes, spaces[:35-(i.len*3)], mnemonics[i.mnemonic].name)

			// Suffix
			line += sizeToSuffix(i.prefix)

			// Arguments
			args := ""
			switch (i.addressMode) {
			default: args += fmt.Sprintf(" ???%d/%x", i.addressMode, i.value)
			case modeImplicit:
			case modeImmediate:
				args += fmt.Sprintf(" #$%x", i.value)
			case modeZeroPage:
				args += fmt.Sprintf(" $%02x", i.value)
			case modeZeroPageX:
				args += fmt.Sprintf(" $%02x,X", i.value)
			case modeZeroPageY:
				args += fmt.Sprintf(" $%02x,Y", i.value)
			case modeRelative:
				if (i.value < 0) {
					args += fmt.Sprintf(" %d", i.value,)
				} else {
					args += fmt.Sprintf(" +%d", i.value)
				}
			case modeAbsolute:
				if (i.prefix == A16) {
					args += fmt.Sprintf(" $%04x", i.value)
				} else {
					args += fmt.Sprintf(" $%06x", i.value)
				}
			case modeAbsoluteX:
				if (i.prefix == A16) {
					args += fmt.Sprintf(" $%04x,X", i.value)
				} else {
					args += fmt.Sprintf(" $%06x,X", i.value)
				}
			case modeAbsoluteY:
				if (i.prefix == A16) {
					args += fmt.Sprintf(" $%04x,Y", i.value)
				} else {
					args += fmt.Sprintf(" $%06x,Y", i.value)
				}
			case modeIndirect:
				if (i.prefix == A16) {
					args += fmt.Sprintf(" ($%04x)", i.value)
				} else {
					args += fmt.Sprintf(" ($%06x)", i.value)
				}
			case modeIndexedIndirectX:
				args += fmt.Sprintf(" ($%02x,X)", i.value & 0xff)
			case modeIndirectIndexedY:
				args += fmt.Sprintf(" ($%02x),Y", i.value & 0xff)
			case modeIndirectZeroPage:
				args += fmt.Sprintf(" ($%02x)", i.value & 0xff)
			case modeAbsoluteIndexedIndirectX:
				if (i.prefix == A16) {
					args += fmt.Sprintf(" ($%04x,X)", i.value)
				} else {
					args += fmt.Sprintf(" ($%06x,X)", i.value)
				}
			case modeX:
				args += fmt.Sprintf(" X")
			case modeXY:
				args += fmt.Sprintf(" XY")
			}
			line += args

			// Append a comment if the original code referenced a symbol
			if (i.symbol != "") {
				spaces := "                                        "
				line += fmt.Sprintf("%s", spaces[:65-len(line)])
				line += fmt.Sprintf("; %s", i.symbol)
			}
			// Append the computed address for relative branches
			if (i.addressMode == modeRelative) {
				line += fmt.Sprintf(" [%x]", i.address + i.len + i.value)
			}
		}

		line += "\n"
		listing.WriteString(line)
		if (byteIdx > 0) {
			out.Write(bytes[:byteIdx])
		}
		//fmt.Printf(line)

		// IF/FOR/LOOP/DO/ETC
		if (i.subBlock != nil) {
			p.outputCodeBlock(i.subBlock.block, out, listing)
		}
	}

	return nil
}

/*
 *  Output the data block
 */
func (p *parser) outputDataBlock(d *dataBlock, out *os.File, listing *os.File) error {
	// Loop through all the data entries
	for e := d.data; e != nil; e = e.next {
		line := fmt.Sprintf("%06x ", e.address)
		bytes := make([]byte, e.len)

		switch (e.size) {
		case R08:
			line += fmt.Sprintf("%02x", e.value & 0x0FF)
			bytes[0] = byte(e.value & 0x0FF)
		case R16:
			line += fmt.Sprintf("%02x %02x", e.value & 0x0FF, (e.value >> 8) & 0x0FF)
			bytes[0] = byte(e.value & 0x0FF)
			bytes[1] = byte((e.value >> 8) & 0x0FF)
		case R24:
			line += fmt.Sprintf("%02x %02x %02x",
					e.value & 0x0FF, (e.value >> 8) & 0x0FF, (e.value >> 16) & 0x0FF)
			bytes[0] = byte(e.value & 0x0FF)
			bytes[1] = byte((e.value >> 8) & 0x0FF)
			bytes[2] = byte((e.value >> 16) & 0x0FF)
		case DSTRING:
			for s := 0; s < len(e.string); s++ {
				line += fmt.Sprintf("%02x ", e.string[s])
				bytes[s] = e.string[s]
			}
			line += fmt.Sprintf("00")
			bytes[e.len-1] = 0
		default:
			return fmt.Errorf("invalid data type %x", e.size)
		}
		line += "\n"

		listing.WriteString(line)
		out.Write(bytes)
		//fmt.Printf(line)
	}

	return nil
}


/*
 *  Output filler between the blocks
 */
func (p *parser) outputFiller(length int, out *os.File, listing *os.File) {
	bytes := make([]byte, length)
	for j := range bytes { bytes[j] = 0x88 }
	filler := fmt.Sprintf("\n; %d BYTES of FILLER\n", length)
	listing.WriteString(filler)
	//fmt.Printf(filler)
	out.Write(bytes)
}

