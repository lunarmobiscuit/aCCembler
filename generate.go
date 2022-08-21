package pomme

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

	return nil
}

/*
 *  Resolve any symbols that were forward references
 */
func (p *parser) resolveSymbols() error {
	// Loop through all the code blocks
	for b := p.code; b != nil; b = b.next {
		// Loop through all the instructions
		for i := b.instr; i != nil; i = i.next {
			// Skip labels
			if (i.mnemonic == 0) {
				continue
			}

			// Skip implicit
			if (i.addressMode == modeImplicit) {
				continue
			}

			// Compute branches
			if (i.addressMode == modeRelative) && (i.hasValue == false) {
				targetAddr, err := b.lookupInstructionLabel(i.symbol)
				if (err == nil) {
					i.hasValue = true
					diff := targetAddr - i.address
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
					i.value = v
				} else {
					s := p.lookupSubroutineName(i.symbol)
					if (s != nil) {
						i.hasValue = true
						i.value = s.startAddr
					} else {
						d := p.lookupDataName(i.symbol)
						if (d != nil) {
							i.hasValue = true
							i.value = d.startAddr
						}
					}
				}

				// Modify the address mode if the length is shorter or longer than expected
				if i.hasValue == false {
					// the symbol was not found
					return fmt.Errorf("the symbol '%s' was not found", i.symbol)
				}
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
					b.name, b.startAddr, b.endAddr, d.name, d.startAddr, d.endAddr)
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
			}
			lastEndAddr = b.endAddr

			sub := fmt.Sprintf("%06x ; SUB %s:\n", b.startAddr, b.name)
			listing.WriteString(sub)
			fmt.Printf(sub)

			// Loop through all the instructions
			for i := b.instr; i != nil; i = i.next {
				line := fmt.Sprintf("%06x ", i.address)
				bytes := make([]byte, 16)
				byteIdx := 0

				// Label
				if (i.mnemonic == 0) {
					line += fmt.Sprintf("                 %s:\n", i.symbol)
					listing.WriteString(line)
					fmt.Printf(line)
					continue
				}

				// Prefix code
				opcodes := ""
				len := i.len
				if (i.prefix > A16) {
					opcodes += fmt.Sprintf("%02x ", i.prefix)
					bytes[byteIdx] = byte(i.prefix & 0xff); byteIdx += 1;
					len -= 1
				}
				// Machine opcode
				opcodes += fmt.Sprintf("%02x ", i.opcode)
				bytes[byteIdx] = byte(i.opcode & 0xff); byteIdx += 1;
				len -= 1

				// Value
				if len > 3 {
					return fmt.Errorf("*** the length for %s is %d, too long", mnemonics[i.mnemonic].name, i.len)
				} else if len >= 3 {
					opcodes += fmt.Sprintf("%02x %02x %02x ",
						i.value & 0xff, (i.value >> 8) & 0xff, (i.value >> 16) & 0xff)
					bytes[byteIdx] = byte(i.value & 0xff); byteIdx += 1;
					bytes[byteIdx] = byte((i.value >> 8) & 0xff); byteIdx += 1;
					bytes[byteIdx] = byte((i.value >> 16) & 0xff); byteIdx += 1;
				} else if len == 2 {
					opcodes += fmt.Sprintf("%02x %02x ", i.value & 0xff, (i.value >> 8) & 0xff)
					bytes[byteIdx] = byte(i.value & 0xff); byteIdx += 1;
					bytes[byteIdx] = byte((i.value >> 8) & 0xff); byteIdx += 1;
				} else if len == 1 {
					opcodes += fmt.Sprintf("%02x ", i.value & 0xff)
					bytes[byteIdx] = byte(i.value & 0xff); byteIdx += 1;
				}

				// Assembly code mneumonic
				spaces := "                    "
				line += fmt.Sprintf("%s%s %s", opcodes, spaces[:18-(i.len*3)], mnemonics[i.mnemonic].name)

				// Suffix
				if (i.prefix != A16) {
					if (i.prefix == R16) || (i.prefix == W16) {
						line += ".w"
					} else if (i.prefix == R24) || (i.prefix == W24) {
						line += ".t"
					}
					if (i.prefix & A24) == A24 {
						line += ".a24"
					}
				}

				// Arguments
				args := ""
				switch (i.addressMode) {
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
						args += fmt.Sprintf(" %d", i.value)
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
				}
				line += args + "\n"

				listing.WriteString(line)
				out.Write(bytes[:byteIdx])
				fmt.Printf(line)
			}

			listing.WriteString("\n")
			fmt.Printf("\n")

			b = b.next
		} else {
			// Fill any gaps between blocks
			if (d.startAddr > lastEndAddr) {
				p.outputFiller(d.startAddr - lastEndAddr, out, listing)
			}
			lastEndAddr = d.endAddr

			data := fmt.Sprintf("%06x ; DATA %s:\n", d.startAddr, d.name)
			listing.WriteString(data)
			fmt.Printf(data)

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
					bytes[1] = byte((e.value >> 16) & 0x0FF)
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
				fmt.Printf(line)
			}

			listing.WriteString("\n")
			fmt.Printf("\n")

			d = d.next
		}
	}

	return nil
}



/*
 *  Output filler between the blocks
 */
func (p *parser) outputFiller(length int, out *os.File, listing *os.File) {
	bytes := make([]byte, length)
	for j := range bytes { bytes[j] = 0x88 }
	filler := fmt.Sprintf("; %d BYTES of FILLER\n\n", length)
	listing.WriteString(filler)
	fmt.Printf(filler)
	out.Write(bytes)
}

