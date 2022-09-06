package aCCembler

import (
	"fmt"
	"strconv"
)

// All the valid assembly language mnemonics
type psedomnemonic struct {
	name		string		// the 3-letter pseudo-mnemonic
	mnemonic	string		// the matching assembly mnemonic
}
var psedomnemonics = []psedomnemonic {
	{"pshr", "pha"},
	{"pshw", "pha"},
	{"psht", "pha"},
	{"pulr", "pla"},
	{"pulw", "pla"},
	{"pult", "pla"},
	{"or_r", "ora"},
	{"or_w", "ora"},
	{"or_t", "ora"},
	{"andr", "and"},
	{"andw", "and"},
	{"andt", "and"},
	{"eorr", "eor"},
	{"eorw", "eor"},
	{"eort", "eor"},
	{"adcr", "adc"},
	{"adcw", "adc"},
	{"adct", "adc"},
	{"sbcr", "sbc"},
	{"sbcw", "sbc"},
	{"sbct", "sbc"},
	{"bitr", "bit"},
	{"bitw", "bit"},
	{"bitt", "bit"},
	{"cmpr", "cmp"},
	{"cmpw", "cmp"},
	{"cmpt", "cmp"},
	{"rolr", "rol"},
	{"rolw", "rol"},
	{"rolt", "rol"},
	{"rorr", "ror"},
	{"rorw", "ror"},
	{"rort", "ror"},
	{"aslr", "asl"},
	{"aslw", "asl"},
	{"aslt", "asl"},
	{"lsrr", "lsr"},
	{"lsrw", "lsr"},
	{"lsrt", "lsr"},
	{"incr", "inc"},
	{"incw", "inc"},
	{"inct", "inc"},
	{"decr", "dec"},
	{"decw", "dec"},
	{"dect", "dec"},
	{"ld_r", "lda"},
	{"ld_w", "lda"},
	{"ld_t", "lda"},
	{"st_r", "sta"},
	{"st_w", "sta"},
	{"st_t", "sta"},
	{"mv_r", "mv_"},
	{"mv_w", "mv_"},
	{"mv_t", "mv_"},
	{"bccr", "bcc"},
	{"bccw", "bcc"},
	{"bcct", "bcc"},
	{"cccr", "ccc"},
	{"cccw", "ccc"},
	{"ccct", "ccc"},
}

/*
 *  Lookup the pseudo-register mnemonic
 */
func (p *parser) isRegisterMnemonic(mnemonic string) bool {
	if (len(mnemonic) < 4) {
		return false
	}

	// Find the table entry for the pseudo-mnemonic
	for p := range psedomnemonics {
		if (mnemonic[:4] == psedomnemonics[p].name) {
			return true
		}
	}

	return false
}

/*
 *  Parse a pseudo-register mnemonic
 *  e.g. LDR0 #123, LDW4 #123456
 */
func (p *parser) parseRegister(mnemonic string) error {
	// Find the table entry for the pseudo-mnemonic
	for n := range psedomnemonics {
		if (mnemonic[:4] == psedomnemonics[n].name) {
			var err error
			reg64 := int64(0)
			if (psedomnemonics[n].mnemonic != "mv_") {
				reg64, err = strconv.ParseInt(mnemonic[4:], 10, 16)
				if (err != nil) {
					return fmt.Errorf("invalid register number (%s)", err)
				} else if (reg64 < 0) || (reg64 > 255) {
					return fmt.Errorf("registers must be 0-255, not %d", reg64)
				}
			}
			reg := int(reg64 & 0xffff)

			// Parse the/any args
			args, err := p.parseArgs()
			if (err != nil) {
				return err
			}

			// Arguments
			comment := fmt.Sprintf("%s", mnemonic)
			switch (args.mode) {
			default: comment += fmt.Sprintf(" ???%d/%x", args.mode, args.value)
			case modeImplicit:
			case modeImmediate:
				comment += fmt.Sprintf(" #$%x", args.value)
			case modeZeroPage:
				comment += fmt.Sprintf(" $%02x", args.value)
			case modeZeroPageX:
				comment += fmt.Sprintf(" $%02x,X", args.value)
			case modeZeroPageY:
				comment += fmt.Sprintf(" $%02x,Y", args.value)
			case modeRelative:
				return fmt.Errorf("%s can not specify a relative address", mnemonic)
			case modeAbsolute:
				if (args.size == A16) {
					comment += fmt.Sprintf(" $%04x", args.value)
				} else {
					comment += fmt.Sprintf(" $%06x", args.value)
				}
			case modeAbsoluteX:
				if (args.size == A16) {
					comment += fmt.Sprintf(" $%04x,X", args.value)
				} else {
					comment += fmt.Sprintf(" $%06x,X", args.value)
				}
			case modeAbsoluteY:
				if (args.size == A16) {
					comment += fmt.Sprintf(" $%04x,Y", args.value)
				} else {
					comment += fmt.Sprintf(" $%06x,Y", args.value)
				}
			case modeIndirect:
				if (args.size == A16) {
					comment += fmt.Sprintf(" ($%04x)", args.value)
				} else {
					comment += fmt.Sprintf(" ($%06x)", args.value)
				}
			case modeIndexedIndirectX:
				comment += fmt.Sprintf(" ($%02x,X)", args.value & 0xff)
			case modeIndirectIndexedY:
				comment += fmt.Sprintf(" ($%02x),Y", args.value & 0xff)
			case modeIndirectZeroPage:
				comment += fmt.Sprintf(" ($%02x)", args.value & 0xff)
			case modeAbsoluteIndexedIndirectX:
				if (args.size == A16) {
					comment += fmt.Sprintf(" ($%04x,X)", args.value)
				} else {
					comment += fmt.Sprintf(" ($%06x,X)", args.value)
				}
			case modeXY:
				comment += fmt.Sprintf(" XY")
			case modeIndirectXY:
				comment += fmt.Sprintf(" (XY)")
			case modeIndexedIndirectXY:
				comment += fmt.Sprintf(" (X),Y")
			}
			p.addInstructionComment(comment)

			sz := R08
			switch (mnemonic[3:4]) {
			case "w": sz = R16
			case "t": sz = R24
			}

			switch (psedomnemonics[n].mnemonic) {
			case "pha":
				return fmt.Errorf("instruction not yet implemented")
			case "pla":
				return fmt.Errorf("instruction not yet implemented")
			case "ora":
				return fmt.Errorf("instruction not yet implemented")
			case "and":
				return fmt.Errorf("instruction not yet implemented")
			case "eor":
				return fmt.Errorf("instruction not yet implemented")
			case "adc":
				return fmt.Errorf("instruction not yet implemented")
			case "sbc":
				return fmt.Errorf("instruction not yet implemented")
			case "bit":
				return fmt.Errorf("instruction not yet implemented")
			case "cmp":
				switch (sz) {
				case R08:
					p.addRegisterInstruction("cmp", modeZeroPage, R08, reg)
				default:
					return fmt.Errorf("cmpw and cmpt are not allowed")
				}
			case "zzz":
			case "ccc":
				reg64, err = strconv.ParseInt(args.symbol[1:], 10, 16)
				if (err != nil) {
					return fmt.Errorf("invalid register number (%s)", err)
				} else if (reg64 < 0) || (reg64 > 255) {
					return fmt.Errorf("registers must be 0-255, not %d", reg64)
				}
				reg2 := int(reg64)
				switch (sz) {
				case R08:
					p.addRegisterInstruction("cmp", modeZeroPage, R08, reg2)
				case R16:
					p.addRegisterInstruction("sec", modeImplicit, R08, 0)
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg)
					p.addRegisterInstruction("sbc", modeZeroPage, R08, reg2)
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg+1)
					p.addRegisterInstruction("sbc", modeZeroPage, R08, reg2+1)
				case R24:
					// Only sets the C bit correctly
					p.addRegisterInstruction("sec", modeImplicit, R08, 0)
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg)
					p.addRegisterInstruction("sbc", modeZeroPage, R08, reg2)
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg+1)
					p.addRegisterInstruction("sbc", modeZeroPage, R08, reg2+1)
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg+2)
					p.addRegisterInstruction("sbc", modeZeroPage, R08, reg2+2)
				}
			case "rol":
				return fmt.Errorf("instruction not yet implemented")
			case "ror":
				return fmt.Errorf("instruction not yet implemented")
			case "asl":
				return fmt.Errorf("instruction not yet implemented")
			case "lsr":
				return fmt.Errorf("instruction not yet implemented")
			case "inc":
				return fmt.Errorf("instruction not yet implemented")
			case "dec":
				return fmt.Errorf("instruction not yet implemented")
			case "lda":
				switch (args.mode) {
				case modeImplicit:
					switch (sz) {
					case R08:
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg)
					default:
						return fmt.Errorf("ld_w and ld_t with makes no sense")
					}
					case modeImmediate:
					p.addRegisterInstruction("lda", modeImmediate, R08, args.value & 0x0FF)
					p.addRegisterInstruction("sta", modeZeroPage, R08, reg)
					if (sz != R08) {
						p.addRegisterInstruction("lda", modeImmediate, R08, (args.value >> 8) & 0x0FF)
						p.addRegisterInstruction("sta", modeZeroPage, R08, reg+1)
					}
					if (sz == R24) {
						p.addRegisterInstruction("lda", modeImmediate, R08, (args.value >> 16) & 0x0FF)
						p.addRegisterInstruction("sta", modeZeroPage, R08, reg+2)
					}
				default:
					p.addRegisterInstruction("lda", args.mode, R08, args.value)
					p.addRegisterInstruction("sta", modeZeroPage, R08, reg)
					if (sz != R08) {
						p.addRegisterInstruction("lda", args.mode, R08, args.value+1)
						p.addRegisterInstruction("sta", modeZeroPage, R08, reg+1)
					}
					if (sz == R24) {
						p.addRegisterInstruction("lda", args.mode, R08, args.value+2)
						p.addRegisterInstruction("sta", modeZeroPage, R08, reg+2)
					}
				}
			case "sta":
				if (args.mode == modeImmediate) {
					return fmt.Errorf("You mean to ld_r, not st_r with that #immediate value")
				}
				p.addRegisterInstruction("lda", modeZeroPage, R08, reg)
				p.addRegisterInstruction("sta", args.mode, R08, args.value)
				if (sz != R08) {
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg+1)
					p.addRegisterInstruction("sta", args.mode, R08, args.value+1)
				}
				if (sz == R24) {
					p.addRegisterInstruction("lda", modeZeroPage, R08, reg+2)
					p.addRegisterInstruction("sta", args.mode, R08, args.value+2)
				}
			case "mv_":
				if (args.mode == modeImmediate) {
					return fmt.Errorf("You mean to ld_r, not mv_r with that #immediate value")
				}
				// Parse the second set of args
				args2, err2 := p.parseArgs()
				if (err2 != nil) {
					return err2
				}
				p.addRegisterInstruction("lda", args.mode, R08, args.value)
				p.addRegisterInstruction("sta", args2.mode, R08, args2.value)
				if (sz != R08) {
					p.addRegisterInstruction("lda", args.mode, R08, args.value+1)
					p.addRegisterInstruction("sta", args2.mode, R08, args2.value+1)
				}
				if (sz == R24) {
					p.addRegisterInstruction("lda", args.mode, R08, args.value+2)
					p.addRegisterInstruction("sta", args2.mode, R08, args2.value+2)
				}
			}
			return nil
		}
	}

	return fmt.Errorf("invalid psedomnemonic %s", mnemonic)
}


/*
 *  Add a blank instruction as a comment in the listing
 */
func (p *parser) addInstructionComment(c string) {
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
	instr.comment = new(comment)
	instr.comment.comment = c
	instr.address = p.currentCode.endAddr
}

/*
 *  Add an instruction to implement the expression
 */
func (p *parser) addRegisterInstruction(mmm string, addressMode int, size int, value int) {
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
	instr.hasValue = true
	instr.value = value

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
}


	
	
