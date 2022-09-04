package pomme

import (
	"errors"
	"fmt"
	"strings"
)

// Prefix codes
const A16 = 0x00	// 16-bit address / 8-bit registers (no prefix)
const A24 = 0x4F	// 24-bit address / 8-bit registers
const A32 = 0x8F	// 32-bit address / 8-bit registers (not implemented in 65C24T8)
const A48 = 0xCF	// 48-bit address / 8-bit registers (not implemented in 65C24T8)
const R08 = 0x00	// 16-bit address / 8-bit registers (no prefix)
const R16 = 0x1F	// 16-bit address / 16-bit registers
const R24 = 0x2F	// 16-bit address / 24-bit registers
const R32 = 0x3F	// 16-bit address / 32-bit registers (not implemented in 65C24T8)
const W16 = 0x5F	// 24-bit address / 16-bit registers
const W24 = 0x6F	// 24-bit address / 24-bit registers
const W32 = 0x7F	// 24-bit address / 32-bit registers (not implemented in 65C24T8)
const V16 = 0x9F	// 32-bit address / 16-bit registers (not implemented in 65C24T8)
const V24 = 0xAF	// 32-bit address / 24-bit registers (not implemented in 65C24T8)
const V32 = 0xBF	// 32-bit address / 32-bit registers (not implemented in 65C24T8)
const U16 = 0xDF	// 48-bit address / 16-bit registers (not implemented in 65C24T8)
const U24 = 0xEF	// 48-bit address / 24-bit registers (not implemented in 65C24T8)
const U32 = 0xFF	// 48-bit address / 32-bit registers (not implemented in 65C24T8)


// Addressing modes
const (
	unknownMode = iota
	modeImplicit // or modeAccumulator
	modeImmediate
	modeZeroPage
	modeZeroPageX
	modeZeroPageY
	modeRelative
	modeAbsolute
	modeAbsoluteX
	modeAbsoluteY
	modeIndirect
	modeIndexedIndirectX
	modeIndirectIndexedY
	// Added on the 65c02
	modeIndirectZeroPage
	modeAbsoluteIndexedIndirectX
)

// All the valid assembly language mnemonics
type mnemonic struct {
	name	string		// the 3-letter mnemonic
	opcode	[]opcode	// the list of valid opcodes
	reg		int			// which register is touched
}
type opcode struct {
	mode	int			// the valid addressing mode
	size	int 		// for the specified size (prefix code)
	opcode	int			// the 1-byte machine opcode
	len		int			// the length in bytes of the machine code
}

var opNOP = []opcode {
	{modeImplicit, A16, 0xEA, 1},
}
var opBRK = []opcode {
	{modeImplicit, A16, 0x00, 1},
}
var opJMP = []opcode {
	{modeAbsolute, A16, 0x4C, 3}, {modeAbsolute, A24, 0x4C, 5},
	{modeIndirect, A16, 0x6C, 3}, {modeIndirect, A24, 0x6C, 5},
	{modeAbsoluteIndexedIndirectX, A16, 0x7C, 3}, {modeAbsoluteIndexedIndirectX, A24, 0x7C, 5},
}
var opJSR = []opcode {
	{modeAbsolute, A16, 0x20, 3}, {modeAbsolute, A24, 0x20, 5},
	{modeIndirectZeroPage, A16, 0x5C, 3}, {modeIndirectZeroPage, A24, 0x5C, 5}, // modeAbsoluteIndexedIndirect 
	{modeIndirect, A16, 0x5C, 3}, {modeIndirect, A24, 0x5C, 5},
}
var opRTI = []opcode {
	{modeImplicit, A16, 0x40, 1}, {modeImplicit, A24, 0x40, 2},
}
var opRTS = []opcode {
	{modeImplicit, A16, 0x60, 1}, {modeImplicit, A24, 0x60, 2},
}
var opPHA = []opcode {
	{modeImplicit, R08, 0x48, 1}, {modeImplicit, R16, 0x48, 2}, {modeImplicit, R24, 0x48, 2},
}
var opPHP = []opcode {
	{modeImplicit, R08, 0x08, 1},
}
var opPLA = []opcode {
	{modeImplicit, R08, 0x68, 1}, {modeImplicit, R16, 0x68, 2}, {modeImplicit, R24, 0x68, 2},
}
var opPLP = []opcode {
	{modeImplicit, R08, 0x28, 1},
}
var opPHX = []opcode {
	{modeImplicit, R08, 0xda, 1}, {modeImplicit, R16, 0xda, 2}, {modeImplicit, R24, 0xda, 2},
}
var opPHY = []opcode {
	{modeImplicit, R08, 0x5a, 1}, {modeImplicit, R16, 0x5a, 2}, {modeImplicit, R24, 0x5a, 2},
}
var opPLX = []opcode {
	{modeImplicit, R08, 0xfa, 1}, {modeImplicit, R16, 0xfa, 2}, {modeImplicit, R24, 0xfa, 2},
}
var opPLY = []opcode {
	{modeImplicit, R08, 0x7a, 1}, {modeImplicit, R16, 0x7a, 2}, {modeImplicit, R24, 0x7a, 2},
}
var opORA = []opcode {
	{modeImmediate, R08, 0x09, 2}, {modeImmediate, R16, 0x09, 4}, {modeImmediate, R24, 0x09, 5},
	{modeZeroPage, R08, 0x05, 2}, {modeZeroPage, R16, 0x05, 3}, {modeZeroPage, R24, 0x05, 3},
	{modeZeroPageX, R08, 0x15, 2}, {modeZeroPageX, R16, 0x15, 3}, {modeZeroPageX, R24, 0x15, 3},
	{modeAbsolute, A16, 0x0D, 3}, {modeAbsolute, A24, 0x0D, 5},
		{modeAbsolute, R16, 0x0D, 4}, {modeAbsolute, R24, 0x0D, 4},
		{modeAbsolute, W16, 0x0D, 5}, {modeAbsolute, W24, 0x0D, 5},
	{modeAbsoluteX, A16, 0x1D, 3}, {modeAbsoluteX, A24, 0x1D, 5},
		{modeAbsoluteX, R16, 0x1D, 4}, {modeAbsoluteX, R24, 0x1D, 4},
		{modeAbsoluteX, W16, 0x1D, 5}, {modeAbsoluteX, W24, 0x1D, 5},
	{modeAbsoluteY, A16, 0x19, 3}, {modeAbsoluteY, A24, 0x19, 5},
		{modeAbsoluteY, R16, 0x19, 4}, {modeAbsoluteY, R24, 0x19, 4},
		{modeAbsoluteY, W16, 0x19, 5}, {modeAbsoluteY, W24, 0x19, 5},
	{modeIndexedIndirectX, A16, 0x01, 2}, {modeIndexedIndirectX, A24, 0x01, 3},
		{modeIndexedIndirectX, R16, 0x01, 3}, {modeIndexedIndirectX, R24, 0x01, 3},
		{modeIndexedIndirectX, W16, 0x01, 3}, {modeIndexedIndirectX, W24, 0x01, 3},
	{modeIndirectIndexedY, A16, 0x11, 2}, {modeIndirectIndexedY, A24, 0x11, 3},
		{modeIndirectIndexedY, R16, 0x11, 3}, {modeIndirectIndexedY, R24, 0x11, 3},
		{modeIndirectIndexedY, W16, 0x11, 3}, {modeIndirectIndexedY, W24, 0x11, 3},
	{modeIndirectZeroPage, A16, 0x12, 2}, {modeIndirectZeroPage, A24, 0x12, 3},
		{modeIndirectZeroPage, R16, 0x12, 3}, {modeIndirectZeroPage, R24, 0x12, 3},
		{modeIndirectZeroPage, W16, 0x12, 3}, {modeIndirectZeroPage, W24, 0x12, 3},
}
var opAND = []opcode {
	{modeImmediate, R08, 0x29, 2}, {modeImmediate, R16, 0x29, 4}, {modeImmediate, R24, 0x29, 5},
	{modeZeroPage, R08, 0x25, 2}, {modeZeroPage, R16, 0x25, 3}, {modeZeroPage, R24, 0x25, 3},
	{modeZeroPageX, R08, 0x35, 2}, {modeZeroPageX, R16, 0x35, 3}, {modeZeroPageX, R24, 0x35, 3},
	{modeAbsolute, A16, 0x2D, 3}, {modeAbsolute, A24, 0x2D, 5},
		{modeAbsolute, R16, 0x2D, 4}, {modeAbsolute, R24, 0x2D, 4},
		{modeAbsolute, W16, 0x2D, 5}, {modeAbsolute, W24, 0x2D, 5},
	{modeAbsoluteX, A16, 0x3D, 3}, {modeAbsoluteX, A24, 0x3D, 5},
		{modeAbsoluteX, R16, 0x3D, 4}, {modeAbsoluteX, R24, 0x3D, 4},
		{modeAbsoluteX, W16, 0x3D, 5}, {modeAbsoluteX, W24, 0x3D, 5},
	{modeAbsoluteY, A16, 0x39, 3}, {modeAbsoluteY, A24, 0x39, 5},
		{modeAbsoluteY, R16, 0x39, 4}, {modeAbsoluteY, R24, 0x39, 4},
		{modeAbsoluteY, W16, 0x39, 5}, {modeAbsoluteY, W24, 0x39, 5},
	{modeIndexedIndirectX, A16, 0x21, 2}, {modeIndexedIndirectX, A24, 0x21, 3},
		{modeIndexedIndirectX, R16, 0x21, 3}, {modeIndexedIndirectX, R24, 0x21, 3},
		{modeIndexedIndirectX, W16, 0x21, 3}, {modeIndexedIndirectX, W24, 0x21, 3},
	{modeIndirectIndexedY, A16, 0x31, 2}, {modeIndirectIndexedY, A24, 0x31, 3},
		{modeIndirectIndexedY, R16, 0x31, 3}, {modeIndirectIndexedY, R24, 0x31, 3},
		{modeIndirectIndexedY, W16, 0x31, 3}, {modeIndirectIndexedY, W24, 0x31, 3},
	{modeIndirectZeroPage, A16, 0x32, 2}, {modeIndirectZeroPage, A24, 0x32, 3},
		{modeIndirectZeroPage, R16, 0x32, 3}, {modeIndirectZeroPage, R24, 0x32, 3},
		{modeIndirectZeroPage, W16, 0x32, 3}, {modeIndirectZeroPage, W24, 0x32, 3},
}
var opEOR = []opcode {
	{modeImmediate, R08, 0x49, 2}, {modeImmediate, R16, 0x49, 4}, {modeImmediate, R24, 0x49, 5},
	{modeZeroPage, R08, 0x45, 2}, {modeZeroPage, R16, 0x45, 3}, {modeZeroPage, R24, 0x45, 3},
	{modeZeroPageX, R08, 0x55, 2}, {modeZeroPageX, R16, 0x55, 3}, {modeZeroPageX, R24, 0x55, 3},
	{modeAbsolute, A16, 0x4D, 3}, {modeAbsolute, A24, 0x4D, 5},
		{modeAbsolute, R16, 0x4D, 4}, {modeAbsolute, R24, 0x4D, 4},
		{modeAbsolute, W16, 0x4D, 5}, {modeAbsolute, W24, 0x4D, 5},
	{modeAbsoluteX, A16, 0x5D, 3}, {modeAbsoluteX, A24, 0x5D, 5},
		{modeAbsoluteX, R16, 0x5D, 4}, {modeAbsoluteX, R24, 0x5D, 4},
		{modeAbsoluteX, W16, 0x5D, 5}, {modeAbsoluteX, W24, 0x5D, 5},
	{modeAbsoluteY, A16, 0x59, 3}, {modeAbsoluteY, A24, 0x59, 5},
		{modeAbsoluteY, R16, 0x59, 4}, {modeAbsoluteY, R24, 0x59, 4},
		{modeAbsoluteY, W16, 0x59, 5}, {modeAbsoluteY, W24, 0x59, 5},
	{modeIndexedIndirectX, A16, 0x41, 2}, {modeIndexedIndirectX, A24, 0x41, 3},
		{modeIndexedIndirectX, R16, 0x41, 3}, {modeIndexedIndirectX, R24, 0x41, 3},
		{modeIndexedIndirectX, W16, 0x41, 3}, {modeIndexedIndirectX, W24, 0x41, 3},
	{modeIndirectIndexedY, A16, 0x51, 2}, {modeIndirectIndexedY, A24, 0x51, 3},
		{modeIndirectIndexedY, R16, 0x51, 3}, {modeIndirectIndexedY, R24, 0x51, 3},
		{modeIndirectIndexedY, W16, 0x51, 3}, {modeIndirectIndexedY, W24, 0x51, 3},
	{modeIndirectZeroPage, A16, 0x52, 2}, {modeIndirectZeroPage, A24, 0x52, 3},
		{modeIndirectZeroPage, R16, 0x52, 3}, {modeIndirectZeroPage, R24, 0x52, 3},
		{modeIndirectZeroPage, W16, 0x52, 3}, {modeIndirectZeroPage, W24, 0x52, 3},
}
var opADC = []opcode {
	{modeImmediate, R08, 0x69, 2}, {modeImmediate, R16, 0x69, 4}, {modeImmediate, R24, 0x69, 5},
	{modeZeroPage, R08, 0x65, 2}, {modeZeroPage, R16, 0x65, 3}, {modeZeroPage, R24, 0x65, 3},
	{modeZeroPageX, R08, 0x75, 2}, {modeZeroPageX, R16, 0x75, 3}, {modeZeroPageX, R24, 0x75, 3},
	{modeAbsolute, A16, 0x6D, 3}, {modeAbsolute, A24, 0x6D, 5},
		{modeAbsolute, R16, 0x6D, 4}, {modeAbsolute, R24, 0x6D, 4},
		{modeAbsolute, W16, 0x6D, 5}, {modeAbsolute, W24, 0x6D, 5},
	{modeAbsoluteX, A16, 0x7D, 3}, {modeAbsoluteX, A24, 0x7D, 5},
		{modeAbsoluteX, R16, 0x7D, 4}, {modeAbsoluteX, R24, 0x7D, 4},
		{modeAbsoluteX, W16, 0x7D, 5}, {modeAbsoluteX, W24, 0x7D, 5},
	{modeAbsoluteY, A16, 0x79, 3}, {modeAbsoluteY, A24, 0x79, 5},
		{modeAbsoluteY, R16, 0x79, 4}, {modeAbsoluteY, R24, 0x79, 4},
		{modeAbsoluteY, W16, 0x79, 5}, {modeAbsoluteY, W24, 0x79, 5},
	{modeIndexedIndirectX, A16, 0x61, 2}, {modeIndexedIndirectX, A24, 0x61, 3},
		{modeIndexedIndirectX, R16, 0x61, 3}, {modeIndexedIndirectX, R24, 0x61, 3},
		{modeIndexedIndirectX, W16, 0x61, 3}, {modeIndexedIndirectX, W24, 0x61, 3},
	{modeIndirectIndexedY, A16, 0x71, 2}, {modeIndirectIndexedY, A24, 0x71, 3},
		{modeIndirectIndexedY, R16, 0x71, 3}, {modeIndirectIndexedY, R24, 0x71, 3},
		{modeIndirectIndexedY, W16, 0x71, 3}, {modeIndirectIndexedY, W24, 0x71, 3},
	{modeIndirectZeroPage, A16, 0x72, 2}, {modeIndirectZeroPage, A24, 0x72, 3},
		{modeIndirectZeroPage, R16, 0x72, 3}, {modeIndirectZeroPage, R24, 0x72, 3},
		{modeIndirectZeroPage, W16, 0x72, 3}, {modeIndirectZeroPage, W24, 0x72, 3},
}
var opSBC = []opcode {
	{modeImmediate, R08, 0xE9, 2}, {modeImmediate, R16, 0xE9, 4}, {modeImmediate, R24, 0xE9, 5},
	{modeZeroPage, R08, 0xE5, 2}, {modeZeroPage, R16, 0xE5, 3}, {modeZeroPage, R24, 0xE5, 3},
	{modeZeroPageX, R08, 0xF5, 2}, {modeZeroPageX, R16, 0xF5, 3}, {modeZeroPageX, R24, 0xF5, 3},
	{modeAbsolute, A16, 0xED, 3}, {modeAbsolute, A24, 0xED, 5},
		{modeAbsolute, R16, 0xED, 4}, {modeAbsolute, R24, 0xED, 4},
		{modeAbsolute, W16, 0xED, 5}, {modeAbsolute, W24, 0xED, 5},
	{modeAbsoluteX, A16, 0xFD, 3}, {modeAbsoluteX, A24, 0xFD, 5},
		{modeAbsoluteX, R16, 0xFD, 4}, {modeAbsoluteX, R24, 0xFD, 4},
		{modeAbsoluteX, W16, 0xFD, 5}, {modeAbsoluteX, W24, 0xFD, 5},
	{modeAbsoluteY, A16, 0xF9, 3}, {modeAbsoluteY, A24, 0xF9, 5},
		{modeAbsoluteY, R16, 0xF9, 4}, {modeAbsoluteY, R24, 0xF9, 4},
		{modeAbsoluteY, W16, 0xF9, 5}, {modeAbsoluteY, W24, 0xF9, 5},
	{modeIndexedIndirectX, A16, 0xE1, 2}, {modeIndexedIndirectX, A24, 0xE1, 3},
		{modeIndexedIndirectX, R16, 0xE1, 3}, {modeIndexedIndirectX, R24, 0xE1, 3},
		{modeIndexedIndirectX, W16, 0xE1, 3}, {modeIndexedIndirectX, W24, 0xE1, 3},
	{modeIndirectIndexedY, A16, 0xF1, 2}, {modeIndirectIndexedY, A24, 0xF1, 3},
		{modeIndirectIndexedY, R16, 0xF1, 3}, {modeIndirectIndexedY, R24, 0xF1, 3},
		{modeIndirectIndexedY, W16, 0xF1, 3}, {modeIndirectIndexedY, W24, 0xF1, 3},
	{modeIndirectZeroPage, A16, 0xF2, 2}, {modeIndirectZeroPage, A24, 0xF2, 3},
		{modeIndirectZeroPage, R16, 0xF2, 3}, {modeIndirectZeroPage, R24, 0xF2, 3},
		{modeIndirectZeroPage, W16, 0xF2, 3}, {modeIndirectZeroPage, W24, 0xF2, 3},
}
var opBIT = []opcode {
	{modeImmediate, R08, 0x89, 2}, {modeImmediate, R16, 0x89, 4}, {modeImmediate, R24, 0x89, 5},
	{modeZeroPage, R08, 0x24, 2}, {modeZeroPage, R16, 0x24, 3}, {modeZeroPage, R24, 0x24, 3},
	{modeZeroPageX, R08, 0x34, 2}, {modeZeroPageX, R16, 0x34, 3}, {modeZeroPageX, R24, 0x34, 3},
	{modeAbsolute, A16, 0x2C, 3}, {modeAbsolute, A24, 0x2C, 5},
		{modeAbsolute, R16, 0x2C, 4}, {modeAbsolute, R24, 0x2C, 4},
		{modeAbsolute, W16, 0x2C, 5}, {modeAbsolute, W24, 0x2C, 5},
	{modeAbsoluteX, A16, 0x3C, 3}, {modeAbsoluteX, A24, 0x3C, 5},
		{modeAbsoluteX, R16, 0x3C, 4}, {modeAbsoluteX, R24, 0x3C, 4},
		{modeAbsoluteX, W16, 0x3C, 5}, {modeAbsoluteX, W24, 0x3C, 5},
}
var opCMP = []opcode {
	{modeImmediate, R08, 0xC9, 2}, {modeImmediate, R16, 0xC9, 4}, {modeImmediate, R24, 0xC9, 5},
	{modeZeroPage, R08, 0xC5, 2}, {modeZeroPage, R16, 0xC5, 3}, {modeZeroPage, R24, 0xC5, 3},
	{modeZeroPageX, R08, 0xD5, 2}, {modeZeroPageX, R16, 0xD5, 3}, {modeZeroPageX, R24, 0xD5, 3},
	{modeAbsolute, A16, 0xCD, 3}, {modeAbsolute, A24, 0xCD, 5},
		{modeAbsolute, R16, 0xCD, 4}, {modeAbsolute, R24, 0xCD, 4},
		{modeAbsolute, W16, 0xCD, 5}, {modeAbsolute, W24, 0xCD, 5},
	{modeAbsoluteX, A16, 0xDD, 3}, {modeAbsoluteX, A24, 0xDD, 5},
		{modeAbsoluteX, R16, 0xDD, 4}, {modeAbsoluteX, R24, 0xDD, 4},
		{modeAbsoluteX, W16, 0xDD, 5}, {modeAbsoluteX, W24, 0xDD, 5},
	{modeAbsoluteY, A16, 0xD9, 3}, {modeAbsoluteY, A24, 0xD9, 5},
		{modeAbsoluteY, R16, 0xD9, 4}, {modeAbsoluteY, R24, 0xD9, 4},
		{modeAbsoluteY, W16, 0xD9, 5}, {modeAbsoluteY, W24, 0xD9, 5},
	{modeIndexedIndirectX, A16, 0xC1, 2}, {modeIndexedIndirectX, A24, 0xC1, 3},
		{modeIndexedIndirectX, R16, 0xC1, 3}, {modeIndexedIndirectX, R24, 0xC1, 3},
		{modeIndexedIndirectX, W16, 0xC1, 3}, {modeIndexedIndirectX, W24, 0xC1, 3},
	{modeIndirectIndexedY, A16, 0xD1, 2}, {modeIndirectIndexedY, A24, 0xD1, 3},
		{modeIndirectIndexedY, R16, 0xD1, 3}, {modeIndirectIndexedY, R24, 0xD1, 3},
		{modeIndirectIndexedY, W16, 0xD1, 3}, {modeIndirectIndexedY, W24, 0xD1, 3},
	{modeIndirectZeroPage, A16, 0xD2, 2}, {modeIndirectZeroPage, A24, 0xD2, 3},
		{modeIndirectZeroPage, R16, 0xD2, 3}, {modeIndirectZeroPage, R24, 0xD2, 3},
		{modeIndirectZeroPage, W16, 0xD2, 3}, {modeIndirectZeroPage, W24, 0xD2, 3},
}
var opCPX = []opcode {
	{modeImmediate, R08, 0xE0, 2}, {modeImmediate, R16, 0xE0, 4}, {modeImmediate, R24, 0xE0, 5},
	{modeZeroPage, R08, 0xE4, 2}, {modeZeroPage, R16, 0xE4, 3}, {modeZeroPage, R24, 0xE4, 3},
	{modeAbsolute, A16, 0xEC, 3}, {modeAbsolute, A24, 0xEC, 5},
		{modeAbsolute, R16, 0xEC, 4}, {modeAbsolute, R24, 0xEC, 4},
		{modeAbsolute, W16, 0xEC, 5}, {modeAbsolute, W24, 0xEC, 5},
}
var opCPY = []opcode {
	{modeImmediate, R08, 0xC0, 2}, {modeImmediate, R16, 0xC0, 4}, {modeImmediate, R24, 0xC0, 5},
	{modeZeroPage, R08, 0xC4, 2}, {modeZeroPage, R16, 0xC4, 3}, {modeZeroPage, R24, 0xC4, 3},
	{modeAbsolute, A16, 0xCC, 3}, {modeAbsolute, A24, 0xCC, 5},
		{modeAbsolute, R16, 0xCC, 4}, {modeAbsolute, R24, 0xCC, 4},
		{modeAbsolute, W16, 0xCC, 5}, {modeAbsolute, W24, 0xCC, 5},
}
var opROL = []opcode {
	{modeImplicit, R08, 0x2A, 1}, {modeImplicit, R16, 0x2A, 2}, {modeImplicit, R24, 0x2A, 2},
	{modeZeroPage, R08, 0x26, 2}, {modeZeroPage, R16, 0x26, 3}, {modeZeroPage, R24, 0x26, 3},
	{modeZeroPageX, R08, 0x36, 2}, {modeZeroPageX, R16, 0x36, 3}, {modeZeroPageX, R24, 0x36, 3},
	{modeAbsolute, A16, 0x2E, 3}, {modeAbsolute, A24, 0x2E, 5},
		{modeAbsolute, R16, 0x2E, 4}, {modeAbsolute, R24, 0x2E, 4},
		{modeAbsolute, W16, 0x2E, 5}, {modeAbsolute, W24, 0x2E, 5},
	{modeAbsoluteX, A16, 0x3E, 3}, {modeAbsoluteX, A24, 0x3E, 5},
		{modeAbsoluteX, R16, 0x3E, 4}, {modeAbsoluteX, R24, 0x3E, 4},
		{modeAbsoluteX, W16, 0x3E, 5}, {modeAbsoluteX, W24, 0x3E, 5},
}
var opROR = []opcode {
	{modeImplicit, R08, 0x6A, 1}, {modeImplicit, R16, 0x6A, 2}, {modeImplicit, R24, 0x6A, 2},
	{modeZeroPage, R08, 0x66, 2}, {modeZeroPage, R16, 0x66, 3}, {modeZeroPage, R24, 0x66, 3},
	{modeZeroPageX, R08, 0x76, 2}, {modeZeroPageX, R16, 0x76, 3}, {modeZeroPageX, R24, 0x76, 3},
	{modeAbsolute, A16, 0x6E, 3}, {modeAbsolute, A24, 0x6E, 5},
		{modeAbsolute, R16, 0x6E, 4}, {modeAbsolute, R24, 0x6E, 4},
		{modeAbsolute, W16, 0x6E, 5}, {modeAbsolute, W24, 0x6E, 5},
	{modeAbsoluteX, A16, 0x7E, 3}, {modeAbsoluteX, A24, 0x7E, 5},
		{modeAbsoluteX, R16, 0x7E, 4}, {modeAbsoluteX, R24, 0x7E, 4},
		{modeAbsoluteX, W16, 0x7E, 5}, {modeAbsoluteX, W24, 0x7E, 5},
}
var opASL = []opcode {
	{modeImplicit, R08, 0x0A, 1}, {modeImplicit, R16, 0x0A, 2}, {modeImplicit, R24, 0x0A, 2},
	{modeZeroPage, R08, 0x06, 2}, {modeZeroPage, R16, 0x06, 3}, {modeZeroPage, R24, 0x06, 3},
	{modeZeroPageX, R08, 0x16, 2}, {modeZeroPageX, R16, 0x16, 3}, {modeZeroPageX, R24, 0x16, 3},
	{modeAbsolute, A16, 0x0E, 3}, {modeAbsolute, A24, 0x0E, 5},
		{modeAbsolute, R16, 0x0E, 4}, {modeAbsolute, R24, 0x0E, 4},
		{modeAbsolute, W16, 0x0E, 5}, {modeAbsolute, W24, 0x0E, 5},
	{modeAbsoluteX, A16, 0x1E, 3}, {modeAbsoluteX, A24, 0x1E, 5},
		{modeAbsoluteX, R16, 0x1E, 4}, {modeAbsoluteX, R24, 0x1E, 4},
		{modeAbsoluteX, W16, 0x1E, 5}, {modeAbsoluteX, W24, 0x1E, 5},
}
var opLSR = []opcode {
	{modeImplicit, R08, 0x4A, 1}, {modeImplicit, R16, 0x4A, 2}, {modeImplicit, R24, 0x4A, 2},
	{modeZeroPage, R08, 0x46, 2}, {modeZeroPage, R16, 0x46, 3}, {modeZeroPage, R24, 0x46, 3},
	{modeZeroPageX, R08, 0x56, 2}, {modeZeroPageX, R16, 0x56, 3}, {modeZeroPageX, R24, 0x56, 3},
	{modeAbsolute, A16, 0x4E, 3}, {modeAbsolute, A24, 0x4E, 5},
		{modeAbsolute, R16, 0x4E, 4}, {modeAbsolute, R24, 0x4E, 4},
		{modeAbsolute, W16, 0x4E, 5}, {modeAbsolute, W24, 0x4E, 5},
	{modeAbsoluteX, A16, 0x5E, 3}, {modeAbsoluteX, A24, 0x5E, 5},
		{modeAbsoluteX, R16, 0x5E, 4}, {modeAbsoluteX, R24, 0x5E, 4},
		{modeAbsoluteX, W16, 0x5E, 5}, {modeAbsoluteX, W24, 0x5E, 5},
}
var opSEC = []opcode {
	{modeImplicit, A16, 0x38, 1},
}
var opSED = []opcode {
	{modeImplicit, A16, 0xF8, 1},
}
var opSEI = []opcode {
	{modeImplicit, A16, 0x78, 1},
}
var opCLC = []opcode {
	{modeImplicit, A16, 0x18, 1},
}
var opCLD = []opcode {
	{modeImplicit, A16, 0xD8, 1},
}
var opCLI = []opcode {
	{modeImplicit, A16, 0x58, 1},
}
var opCLV = []opcode {
	{modeImplicit, A16, 0xB8, 1},
}
var opINC = []opcode {
	{modeImplicit, R08, 0x1A, 1}, {modeImplicit, R16, 0x1A, 2}, {modeImplicit, R24, 0x1A, 2},
	{modeZeroPage, R08, 0xE6, 2}, {modeZeroPage, R16, 0xE6, 3}, {modeZeroPage, R24, 0xE6, 3},
	{modeZeroPageX, R08, 0xF6, 2}, {modeZeroPageX, R16, 0xF6, 3}, {modeZeroPageX, R24, 0xF6, 3},
	{modeAbsolute, A16, 0xEE, 3}, {modeAbsolute, A24, 0xEE, 5},
		{modeAbsolute, R16, 0xEE, 4}, {modeAbsolute, R24, 0xEE, 4},
		{modeAbsolute, W16, 0xEE, 5}, {modeAbsolute, W24, 0xEE, 5},
	{modeAbsoluteX, A16, 0xFE, 3}, {modeAbsoluteX, A24, 0xFE, 5},
		{modeAbsoluteX, R16, 0xFE, 4}, {modeAbsoluteX, R24, 0xFE, 4},
		{modeAbsoluteX, W16, 0xFE, 5}, {modeAbsoluteX, W24, 0xFE, 5},
}
var opDEC = []opcode {
	{modeImplicit, R08, 0x3A, 1}, {modeImplicit, R16, 0x3A, 2}, {modeImplicit, R24, 0x3A, 2},
	{modeZeroPage, R08, 0xC6, 2}, {modeZeroPage, R16, 0xC6, 3}, {modeZeroPage, R24, 0xC6, 3},
	{modeZeroPageX, R08, 0xD6, 2}, {modeZeroPageX, R16, 0xD6, 3}, {modeZeroPageX, R24, 0xD6, 3},
	{modeAbsolute, A16, 0xCE, 3}, {modeAbsolute, A24, 0xCE, 5},
		{modeAbsolute, R16, 0xCE, 4}, {modeAbsolute, R24, 0xCE, 4},
		{modeAbsolute, W16, 0xCE, 5}, {modeAbsolute, W24, 0xCE, 5},
	{modeAbsoluteX, A16, 0xDE, 3}, {modeAbsoluteX, A24, 0xDE, 5},
		{modeAbsoluteX, R16, 0xDE, 4}, {modeAbsoluteX, R24, 0xDE, 4},
		{modeAbsoluteX, W16, 0xDE, 5}, {modeAbsoluteX, W24, 0xDE, 5},
}
var opINX = []opcode {
	{modeImplicit, R08, 0xE8, 1}, {modeImplicit, R16, 0xE8, 2}, {modeImplicit, R24, 0xE8, 2},
}
var opDEX = []opcode {
	{modeImplicit, R08, 0xCA, 1}, {modeImplicit, R16, 0xCA, 2}, {modeImplicit, R24, 0xCA, 2},
}
var opINY = []opcode {
	{modeImplicit, R08, 0xC8, 1}, {modeImplicit, R16, 0xC8, 2}, {modeImplicit, R24, 0xC8, 2},
}
var opDEY = []opcode {
	{modeImplicit, R08, 0x88, 1}, {modeImplicit, R16, 0x88, 2}, {modeImplicit, R24, 0x88, 2},
}
var opTAX = []opcode {
	{modeImplicit, R08, 0xAA, 1}, {modeImplicit, R16, 0xAA, 2}, {modeImplicit, R24, 0xAA, 2},
}
var opTXA = []opcode {
	{modeImplicit, R08, 0x8A, 1}, {modeImplicit, R16, 0x8A, 2}, {modeImplicit, R24, 0x8A, 2},
}
var opTAY = []opcode {
	{modeImplicit, R08, 0xA8, 1}, {modeImplicit, R16, 0xA8, 2}, {modeImplicit, R24, 0xA8, 2},
}
var opTYA = []opcode {
	{modeImplicit, R08, 0x98, 1}, {modeImplicit, R16, 0x98, 2}, {modeImplicit, R24, 0x98, 2},
}
var opTXS = []opcode {
	{modeImplicit, R08, 0x9A, 1}, {modeImplicit, R16, 0x9A, 2}, {modeImplicit, R24, 0x9A, 2},
}
var opTSX = []opcode {
	{modeImplicit, R08, 0xBA, 1}, {modeImplicit, R16, 0xBA, 2}, {modeImplicit, R24, 0xBA, 2},
}
var opLDA = []opcode {
	{modeImmediate, R08, 0xA9, 2}, {modeImmediate, R16, 0xA9, 4}, {modeImmediate, R24, 0xA9, 5},
	{modeZeroPage, R08, 0xA5, 2}, {modeZeroPage, R16, 0xA5, 3}, {modeZeroPage, R24, 0xA5, 3},
	{modeZeroPageX, R08, 0xB5, 2}, {modeZeroPageX, R16, 0xB5, 3}, {modeZeroPageX, R24, 0xB5, 3},
	{modeAbsolute, A16, 0xAD, 3}, {modeAbsolute, A24, 0xAD, 5},
		{modeAbsolute, R16, 0xAD, 4}, {modeAbsolute, R24, 0xAD, 4},
		{modeAbsolute, W16, 0xAD, 5}, {modeAbsolute, W24, 0xAD, 5},
	{modeAbsoluteX, A16, 0xBD, 3}, {modeAbsoluteX, A24, 0xBD, 5},
		{modeAbsoluteX, R16, 0xBD, 4}, {modeAbsoluteX, R24, 0xBD, 4},
		{modeAbsoluteX, W16, 0xBD, 5}, {modeAbsoluteX, W24, 0xBD, 5},
	{modeAbsoluteY, A16, 0xB9, 3}, {modeAbsoluteY, A24, 0xB9, 5},
		{modeAbsoluteY, R16, 0xB9, 4}, {modeAbsoluteY, R24, 0xB9, 4},
		{modeAbsoluteY, W16, 0xB9, 5}, {modeAbsoluteY, W24, 0xB9, 5},
	{modeIndexedIndirectX, A16, 0xA1, 2}, {modeIndexedIndirectX, A24, 0xA1, 3},
		{modeIndexedIndirectX, R16, 0xA1, 3}, {modeIndexedIndirectX, R24, 0xA1, 3},
		{modeIndexedIndirectX, W16, 0xA1, 3}, {modeIndexedIndirectX, W24, 0xA1, 3},
	{modeIndirectIndexedY, A16, 0xB1, 2}, {modeIndirectIndexedY, A24, 0xB1, 3},
		{modeIndirectIndexedY, R16, 0xB1, 3}, {modeIndirectIndexedY, R24, 0xB1, 3},
		{modeIndirectIndexedY, W16, 0xB1, 3}, {modeIndirectIndexedY, W24, 0xB1, 3},
	{modeIndirectZeroPage, A16, 0xB2, 2}, {modeIndirectZeroPage, A24, 0xB2, 3},
		{modeIndirectZeroPage, R16, 0xB2, 3}, {modeIndirectZeroPage, R24, 0xB2, 3},
		{modeIndirectZeroPage, W16, 0xB2, 3}, {modeIndirectZeroPage, W24, 0xB2, 3},
}
var opSTA = []opcode {
	{modeZeroPage, R08, 0x85, 2}, {modeZeroPage, R16, 0x85, 3}, {modeZeroPage, R24, 0x85, 3},
	{modeZeroPageX, R08, 0x95, 2}, {modeZeroPageX, R16, 0x95, 3}, {modeZeroPageX, R24, 0x95, 3},
	{modeAbsolute, A16, 0x8D, 3}, {modeAbsolute, A24, 0x8D, 5},
		{modeAbsolute, R16, 0x8D, 4}, {modeAbsolute, R24, 0x8D, 4},
		{modeAbsolute, W16, 0x8D, 5}, {modeAbsolute, W24, 0x8D, 5},
	{modeAbsoluteX, A16, 0x9D, 3}, {modeAbsoluteX, A24, 0x9D, 5},
		{modeAbsoluteX, R16, 0x9D, 4}, {modeAbsoluteX, R24, 0x9D, 4},
		{modeAbsoluteX, W16, 0x9D, 5}, {modeAbsoluteX, W24, 0x9D, 5},
	{modeAbsoluteY, A16, 0x99, 3}, {modeAbsoluteY, A24, 0x99, 5},
		{modeAbsoluteY, R16, 0x99, 4}, {modeAbsoluteY, R24, 0x99, 4},
		{modeAbsoluteY, W16, 0x99, 5}, {modeAbsoluteY, W24, 0x99, 5},
	{modeIndexedIndirectX, A16, 0x81, 2}, {modeIndexedIndirectX, A24, 0x81, 3},
		{modeIndexedIndirectX, R16, 0x81, 3}, {modeIndexedIndirectX, R24, 0x81, 3},
		{modeIndexedIndirectX, W16, 0x81, 3}, {modeIndexedIndirectX, W24, 0x81, 3},
	{modeIndirectIndexedY, A16, 0x91, 2}, {modeIndirectIndexedY, A24, 0x91, 3},
		{modeIndirectIndexedY, R16, 0x91, 3}, {modeIndirectIndexedY, R24, 0x91, 3},
		{modeIndirectIndexedY, W16, 0x91, 3}, {modeIndirectIndexedY, W24, 0x91, 3},
	{modeIndirectZeroPage, A16, 0x92, 2}, {modeIndirectZeroPage, A24, 0x92, 3},
		{modeIndirectZeroPage, R16, 0x92, 3}, {modeIndirectZeroPage, R24, 0x92, 3},
		{modeIndirectZeroPage, W16, 0x92, 3}, {modeIndirectZeroPage, W24, 0x92, 3},
}
var opLDX = []opcode {
	{modeImmediate, R08, 0xA2, 2}, {modeImmediate, R16, 0xA2, 4}, {modeImmediate, R24, 0xA2, 5},
	{modeZeroPage, R08, 0xA6, 2}, {modeZeroPage, R16, 0xA6, 3}, {modeZeroPage, R24, 0xA6, 3},
	{modeZeroPageY, R08, 0xB6, 2}, {modeZeroPageY, R16, 0xB6, 3}, {modeZeroPageY, R24, 0xB6, 3},
	{modeAbsolute, A16, 0xAE, 3}, {modeAbsolute, A24, 0xAE, 5},
		{modeAbsolute, R16, 0xAE, 4}, {modeAbsolute, R24, 0xAE, 4},
		{modeAbsolute, W16, 0xAE, 5}, {modeAbsolute, W24, 0xAE, 5},
	{modeAbsoluteY, A16, 0xBE, 3}, {modeAbsoluteY, A24, 0xBE, 5},
		{modeAbsoluteY, R16, 0xBE, 4}, {modeAbsoluteY, R24, 0xBE, 4},
		{modeAbsoluteY, W16, 0xBE, 5}, {modeAbsoluteY, W24, 0xBE, 5},
}
var opSTX = []opcode {
	{modeZeroPage, R08, 0x86, 2}, {modeZeroPage, R16, 0x86, 3}, {modeZeroPage, R24, 0x86, 3},
	{modeZeroPageY, R08, 0x96, 2}, {modeZeroPageY, R16, 0x96, 3}, {modeZeroPageY, R24, 0x96, 3},
	{modeAbsolute, A16, 0x8E, 3}, {modeAbsolute, A24, 0x8E, 5},
		{modeAbsolute, R16, 0x8E, 4}, {modeAbsolute, R24, 0x8E, 4},
		{modeAbsolute, W16, 0x8E, 5}, {modeAbsolute, W24, 0x8E, 5},
}
var opLDY = []opcode {
	{modeImmediate, R08, 0xA0, 2}, {modeImmediate, R16, 0xA0, 4}, {modeImmediate, R24, 0xA0, 5},
	{modeZeroPage, R08, 0xA4, 2}, {modeZeroPage, R16, 0xA4, 3}, {modeZeroPage, R24, 0xA4, 3},
	{modeZeroPageX, R08, 0xB4, 2}, {modeZeroPageX, R16, 0xB4, 3}, {modeZeroPageX, R24, 0xB4, 3},
	{modeAbsolute, A16, 0xAC, 3}, {modeAbsolute, A24, 0xAC, 5},
		{modeAbsolute, R16, 0xAC, 4}, {modeAbsolute, R24, 0xAC, 4},
		{modeAbsolute, W16, 0xAC, 5}, {modeAbsolute, W24, 0xAC, 5},
	{modeAbsoluteX, A16, 0xBC, 3}, {modeAbsoluteX, A24, 0xBC, 5},
		{modeAbsoluteX, R16, 0xBC, 4}, {modeAbsoluteX, R24, 0xBC, 4},
		{modeAbsoluteX, W16, 0xBC, 5}, {modeAbsoluteX, W24, 0xBC, 5},
}
var opSTY = []opcode {
	{modeZeroPage, R08, 0x84, 2}, {modeZeroPage, R16, 0x84, 3}, {modeZeroPage, R24, 0x84, 3},
	{modeZeroPageX, R08, 0x94, 2}, {modeZeroPageX, R16, 0x94, 3}, {modeZeroPageX, R24, 0x94, 3},
	{modeAbsolute, A16, 0x8C, 3}, {modeAbsolute, A24, 0x8C, 5},
		{modeAbsolute, R16, 0x8C, 4}, {modeAbsolute, R24, 0x8C, 4},
		{modeAbsolute, W16, 0x8C, 5}, {modeAbsolute, W24, 0x8C, 5},
}
var opBCC = []opcode {
	{modeRelative, A16, 0x90, 2}, {modeRelative, A24, 0x90, 4},
}
var opBCS = []opcode {
	{modeRelative, A16, 0xB0, 2}, {modeRelative, A24, 0xB0, 4},
}
var opBNE = []opcode {
	{modeRelative, A16, 0xD0, 2}, {modeRelative, A24, 0xD0, 4},
}
var opBEQ = []opcode {
	{modeRelative, A16, 0xF0, 2}, {modeRelative, A24, 0xF0, 4},
}
var opBPL = []opcode {
	{modeRelative, A16, 0x10, 2}, {modeRelative, A24, 0x10, 4},
}
var opBMI = []opcode {
	{modeRelative, A16, 0x30, 2}, {modeRelative, A24, 0x30, 4},
}
var opBGE = []opcode {
	{modeRelative, A16, 0xB0, 2}, {modeRelative, A24, 0xB0, 4},
}
var opBLT = []opcode {
	{modeRelative, A16, 0x90, 2}, {modeRelative, A24, 0x90, 4},
}
var opBVC = []opcode {
	{modeRelative, A16, 0x50, 2}, {modeRelative, A24, 0x50, 4},
}
var opBVS = []opcode {
	{modeRelative, A16, 0x70, 2}, {modeRelative, A24, 0x70, 4},
}
var opBRA = []opcode {
	{modeRelative, A16, 0x80, 2}, {modeRelative, A24, 0x80, 4},
}
var opSTZ = []opcode {
	{modeZeroPage, R08, 0x64, 2}, {modeZeroPage, R16, 0x64, 3}, {modeZeroPage, R24, 0x64, 3},
	{modeZeroPageX, R08, 0x74, 2}, {modeZeroPageX, R16, 0x74, 3}, {modeZeroPageX, R24, 0x74, 3},
	{modeAbsolute, A16, 0x9C, 3}, {modeAbsolute, A24, 0x9C, 5},
		{modeAbsolute, R16, 0x9C, 4}, {modeAbsolute, R24, 0x9C, 4},
		{modeAbsolute, W16, 0x9C, 5}, {modeAbsolute, W24, 0x9C, 5},
	{modeAbsoluteX, A16, 0x9E, 3}, {modeAbsoluteX, A24, 0x9E, 5},
		{modeAbsoluteX, R16, 0x9E, 4}, {modeAbsoluteX, R24, 0x9E, 4},
		{modeAbsoluteX, W16, 0x9E, 5}, {modeAbsoluteX, W24, 0x9E, 5},
}
var opTRB = []opcode {
	{modeZeroPage, R08, 0x14, 2}, {modeZeroPage, R16, 0x14, 3}, {modeZeroPage, R24, 0x14, 3},
	{modeAbsolute, A16, 0x1C, 3}, {modeAbsolute, A24, 0x1C, 5},
		{modeAbsolute, R16, 0x1C, 4}, {modeAbsolute, R24, 0x1C, 4},
		{modeAbsolute, W16, 0x1C, 5}, {modeAbsolute, W24, 0x1C, 5},
}
var opTSB = []opcode {
	{modeZeroPage, R08, 0x04, 2}, {modeZeroPage, R16, 0x04, 3}, {modeZeroPage, R24, 0x04, 3},
	{modeAbsolute, A16, 0x0C, 3}, {modeAbsolute, A24, 0x0C, 5},
		{modeAbsolute, R16, 0x0C, 4}, {modeAbsolute, R24, 0x0C, 4},
		{modeAbsolute, W16, 0x0C, 5}, {modeAbsolute, W24, 0x0C, 5},
}
var opCPU = []opcode {
	{modeImplicit, A16, 0x0F, 1},
}
var opA24 = []opcode {
	{modeImplicit, A16, 0x4F, 1},
}
var opR16 = []opcode {
	{modeImplicit, A16, 0x1F, 1},
}
var opR24 = []opcode {
	{modeImplicit, A16, 0x2F, 1},
}
var opW16 = []opcode {
	{modeImplicit, A16, 0x5F, 1},
}
var opW24 = []opcode {
	{modeImplicit, A16, 0x6F, 1},
}
var opSWS = []opcode {
	{modeImplicit, A16, 0xFC, 1},
}
var opSL8 = []opcode {
	{modeImplicit, R08, 0x0B, 1}, {modeImplicit, R16, 0x0B, 2}, {modeImplicit, R24, 0x0B, 2},
}
var opSR8 = []opcode {
	{modeImplicit, R08, 0x1B, 1}, {modeImplicit, R16, 0x1B, 2}, {modeImplicit, R24, 0x1B, 2},
}
var opXSL = []opcode {
	{modeImplicit, R08, 0x2B, 1}, {modeImplicit, R16, 0x2B, 2}, {modeImplicit, R24, 0x2B, 2},
}
var opYSL = []opcode {
	{modeImplicit, R08, 0x3B, 1}, {modeImplicit, R16, 0x3B, 2}, {modeImplicit, R24, 0x3B, 2},
}
var opADX = []opcode {
	{modeImplicit, R08, 0xDB, 1}, {modeImplicit, R16, 0xDB, 2}, {modeImplicit, R24, 0xDB, 2},
}
var opADY = []opcode {
	{modeImplicit, R08, 0xEB, 1}, {modeImplicit, R16, 0xEB, 2}, {modeImplicit, R24, 0xEB, 2},
}
var opAXY = []opcode {
	{modeImplicit, R08, 0xFB, 1}, {modeImplicit, R16, 0xFB, 2}, {modeImplicit, R24, 0xFB, 2},
}
var opTHR = []opcode {
	{modeImplicit, A16, 0x03, 1},
}
var opTHW = []opcode {
	{modeImplicit, A16, 0x13, 1},
}
var opTHY = []opcode {
	{modeImplicit, A16, 0x23, 1},
}
var opTHI = []opcode {
	{modeAbsolute, A16, 0x33, 3}, {modeAbsolute, A24, 0x33, 4},
}
var opTTA = []opcode {
	{modeImplicit, R08, 0x43, 1}, {modeImplicit, R16, 0x43, 2}, {modeImplicit, R24, 0x43, 2},
}
var opTAT = []opcode {
	{modeImplicit, R08, 0x53, 1}, {modeImplicit, R16, 0x53, 2}, {modeImplicit, R24, 0x53, 2},
}
var opTTS = []opcode {
	{modeImplicit, R08, 0x63, 1}, {modeImplicit, R16, 0x63, 2}, {modeImplicit, R24, 0x63, 2},
}
var opTST = []opcode {
	{modeImplicit, R08, 0x73, 1}, {modeImplicit, R16, 0x73, 2}, {modeImplicit, R24, 0x73, 2},
}

var mnemonics = []mnemonic {
	{":", nil, N_A}, // used to store labels
	{"nop", opNOP, N_A},
	{"brk", opBRK, N_A},
	{"jmp", opJMP, N_A},
	{"jsr", opJSR, N_A},
	{"rti", opRTI, N_A},
	{"rts", opRTS, N_A},
	{"pha", opPHA, REG_A},
	{"php", opPHP, N_A},
	{"pla", opPLA, REG_A},
	{"plp", opPLP, N_A},
	{"ora", opORA, REG_A},
	{"and", opAND, REG_A},
	{"eor", opEOR, REG_A},
	{"adc", opADC, REG_A},
	{"sbc", opSBC, REG_A},
	{"bit", opBIT, REG_A},
	{"cmp", opCMP, REG_A},
	{"cpx", opCPX, REG_X},
	{"cpy", opCPY, REG_Y},
	{"rol", opROL, REG_A},
	{"ror", opROR, REG_A},
	{"asl", opASL, REG_A},
	{"lsr", opLSR, REG_A},
	{"sec", opSEC, N_A},
	{"sed", opSED, N_A},
	{"sei", opSEI, N_A},
	{"clc", opCLC, N_A},
	{"cld", opCLD, N_A},
	{"cli", opCLI, N_A},
	{"clv", opCLV, N_A},
	{"inc", opINC, REG_A},
	{"dec", opDEC, REG_A},
	{"inx", opINX, REG_X},
	{"iny", opINY, REG_Y},
	{"dex", opDEX, REG_X},
	{"dey", opDEY, REG_Y},
	{"tax", opTAX, REG_X},
	{"tay", opTAY, REG_Y},
	{"txa", opTXA, REG_A},
	{"tya", opTYA, REG_A},
	{"txs", opTXS, N_A},
	{"tsx", opTSX, N_A},
	{"lda", opLDA, REG_A},
	{"sta", opSTA, REG_A},
	{"ldx", opLDX, REG_X},
	{"stx", opSTX, REG_X},
	{"ldy", opLDY, REG_Y},
	{"sty", opSTY, REG_Y},
	{"bcc", opBCC, N_A},
	{"bcs", opBCS, N_A},
	{"bne", opBNE, N_A},
	{"beq", opBEQ, N_A},
	{"bpl", opBPL, N_A},
	{"bmi", opBMI, N_A},
	{"bge", opBGE, N_A},
	{"blt", opBLT, N_A},
	{"bvc", opBVC, N_A},
	{"bvs", opBVS, N_A},
	// 65C02
	{"bra", opBRA, N_A},
	{"phx", opPHX, REG_X},
	{"phy", opPHY, REG_Y},
	{"plx", opPLX, REG_X},
	{"ply", opPLY, REG_Y},
	{"stz", opSTZ, N_A},
	{"trb", opTRB, N_A},
	{"tsb", opTSB, N_A},
	// 6524T8
	{"cpu", opCPU, N_A},
	{"a24", opA24, N_A},
	{"r16", opR16, N_A},
	{"r24", opR24, N_A},
	{"w16", opW16, N_A},
	{"w24", opW24, N_A},
	{"sws", opSWS, N_A},
	{"sl8", opSL8, REG_A},
	{"sr8", opSR8, REG_A},
	{"adx", opADX, REG_A},
	{"ady", opADY, REG_A},
	{"axy", opAXY, REG_A},
	{"xsl", opXSL, REG_X},
	{"ysl", opYSL, REG_Y},
	// threads
	{"thr", opTHR, N_A},
	{"thw", opTHW, N_A},
	{"thy", opTHY, N_A},
	{"thi", opTHI, N_A},
	{"tta", opTTA, REG_A},
	{"tat", opTAT, REG_A},
	{"tts", opTTS, N_A},
	{"tst", opTST, N_A},
}

// Parsed addressing arguments
type assemblyArgs struct {
	mode		int
	size		int
	symbol		string
	hasValue	bool
	value		int
}


/*
 *  Lookup the assembly language mnemonic
 */
func (p *parser) isMnemonic(mnemonic string) bool {
	// Find the table entry for the mnemonic
	for m := range mnemonics {
		if (mnemonic == mnemonics[m].name) {
			return true
		}
	}

	return false
}

/*
 *  Lookup and parse the assembly language mnemonic
 */
func (p *parser) parseMnemonic(mnemonic string) error {
	// Parse any optional width suffix
	sz := p.parseOpWidth()
	if (sz == A32) {
		return errors.New("32-bit addresses are not supported")
	} else if (sz == A48) {
		return errors.New("48-bit registers is not supported")
	} else if (sz == R32) {
		return errors.New("32-bit registers is not supported")
	}

	// Parse the/any args
	args, err := p.parseArgs()
	if (err != nil) {
		return err
	}

	// Blend the specified size with the size implied by the arguments
	args.size |= sz

	// Find the matching mnemonic, matching addressMode and size
	for m := range mnemonics {
		if (mnemonic == mnemonics[m].name) {
			return p.addInstruction(m, args.mode, args.size, args.hasValue, args.symbol, args.value)
		}
	}

	return fmt.Errorf("mnemonic '%s' is invalid", mnemonic)
}

/*
 *  Return the explit opcode width
 *  (returning the string and index)
 */
func (p *parser) parseOpWidth() int {
	// There is no '.' and thus no suffix
	if (p.peekChar() != '.') {
		return A16
	}

	// Opcode can have suffix of: .aw .a24 .a32 .a48 .b .8 .w .16 .t .24 .f .32
	p.skip(1)
	sym1 := p.peekChar()
	sym2 := p.peekAhead(1)
	sym3 := p.peekAhead(2)
	if (sym1 == 'a') && (sym2 == '1') && (sym3 == '6') {
		p.skip(3); return A16;
	} else if (sym1 == 'a') && (sym2 == 'w') {
		p.skip(2); return A24;
	} else if (sym1 == 'a') && (sym2 == '2') && (sym3 == '4') {
		p.skip(3); return A24;
	} else if (sym1 == 'a') && (sym2 == '3') && (sym3 == '2') {
		p.skip(3); return A32;
	} else if (sym1 == 'a') && (sym2 == '4') && (sym3 == '8') {
		p.skip(3); return A48;
	} else if (sym1 == 'b') || (sym1 == '8') {
		p.skip(1); return R08;
	} else if (sym1 == 'w') {
		p.skip(1); return R16;
	} else if (sym1 == '1') && (sym2 == '6') {
		p.skip(2); return R16;
	} else if (sym1 == 't') {
		p.skip(1); return R24;
	} else if (sym1 == '2') && (sym2 == '4') {
		p.skip(2); return R24;
	} else if (sym1 == 'f') {
		p.skip(1); return R32;
	} else if (sym1 == '3') && (sym2 == '2') {
		p.skip(2); return R32;
	}

	return A16
}

/*
 *  Parse the arguments after the mnemonic
 */
func (p *parser) parseArgs() (assemblyArgs, error) {
	p.skipWhitespace()

	var args assemblyArgs
	args.size = p.abWidth
	sym := p.peekChar()
	if (sym == ';') || (sym == '/') || (sym == CR) || (sym == LF) { // mmm
		args.mode = modeImplicit
		args.size = A16
	} else if (sym == '+') || (sym == '-') { // bmm +$aa or bmm -$aa or bmm +$aaaa or bmm -$aaaa
		p.skip(1)
		args.mode = modeRelative
		if p.isNextAZ() {
			symbol := strings.ToLower(p.nextAZ_az_09())

			// keyword "break" meaning 'goto end of the current block'
			if (symbol == "break") {
				symbol = strings.ToLower(p.currentCode.name + "_end")
			}

			args.symbol = symbol
			args.hasValue = false
		} else {
			value, err := p.nextValue()
			if (err != nil) {
				return args, err
			}
			if (sym == '+') {
				args.value = value
			} else {
				args.value = -value
			}
			args.hasValue = true
		}

		if (args.hasValue) {
			if (args.value >= -128) && (args.value <= 127) {
				args.size = A16
			} else if (args.value >= -32768) && (args.value <= 32767) {
				args.size = A24
			} else {
				return args, fmt.Errorf("the branch distance '%d' is too far for even 24-bits", args.value)
			}
		} else {
			args.size = A16 // short branch unless .suffix says otherwise
		}
	} else if (sym == '#') { // mmm #$aa or mmm #$aaaa
		p.skip(1)
		args.mode = modeImmediate
		if p.peekChar() == '\'' {
			p.skip(1)
			c := p.nextChar()
			if p.peekChar() != '\'' {
				return args, fmt.Errorf("no matching single quote after %c", c)
			}
			p.skip(1)
			if (p.peekChar() == 'h') || (p.peekChar() == 'H') {
				p.skip(1)
				c |= 0x80
			}
			args.value = int(c)
			args.size = R08
			args.hasValue = true
		} else if (p.peekChar() == '@') {
			p.skip(1)
			symbol := p.nextAZ_az_09()
			args.symbol = symbol
			address, size, err := p.lookupVariable(p.currentCode, symbol)
			if (err != nil) {
				return args, fmt.Errorf("unknown variable '%s'", symbol)
			}
			args.value = address
			args.size |= size
			args.hasValue = true
		} else if p.isNextAZ() {
			symbol := p.nextAZ_az_09()
			args.symbol = symbol
			value, err := p.lookupConstant(symbol)
			if (err == nil) {
				if (p.peekChar() == '+') {
					p.skip(1)
					plus, _ := p.nextValue()
					value += plus
				}
				args.value = value
				args.hasValue = true
			} else {
				d := p.lookupDataName(symbol)
				if (d != nil) {
					args.value = d.startAddr
					args.hasValue = true
				} else {
					// Resolve this later
					args.hasValue = false
				}

			}
		} else {
			value, err := p.nextValue()
			if (err != nil) {
				return args, err
			}
			args.value = value
			args.hasValue = true
		}

		if (args.hasValue) {
			args.size = valueToPrefix(args.value)
			if (args.size == R32) {
				return args, fmt.Errorf("the value '%x' is too large for 24-bits", args.value)
			}
		} else {
			args.size = R08 // assume 8-bits, but resolve later
		}

		p.skipWhitespace()
		if p.peekChar() == ',' {
			return args, errors.New("indexed immeidate #addr,X is invalid")
		}
	} else if (sym == '(') { // mmm ($aa) or mmm ($aaaa)
		p.skip(1)
		args.mode = modeIndirect // mmm ($aaaa)
		if (p.peekChar() == '@') {
			p.skip(1)
			symbol := p.nextAZ_az_09()
			args.symbol = symbol
			address, size, err := p.lookupVariable(p.currentCode, symbol)
			if (err != nil) {
				return args, fmt.Errorf("unknown variable '%s'", symbol)
			}
			args.value = address
			args.size |= size
			args.hasValue = true
		} else if p.isNextAZ() {
			symbol := p.nextAZ_az_09()
			args.symbol = symbol
			value, err := p.lookupConstant(symbol)
			if (err == nil) {
				if (p.peekChar() == '+') {
					p.skip(1)
					plus, _ := p.nextValue()
					value += plus
				}
				args.value = value
				args.hasValue = true
			} else {
				args.symbol = symbol
				args.hasValue = false
			}
		} else {
			value, err := p.nextValue()
			if (err != nil) {
				return args, err
			}
			args.value = value
			args.hasValue = true
		}

		if (args.hasValue) {
			args.size = addressToPrefix(args.value)
			if (args.value <= 0x0FF) {
				args.mode = modeIndirectZeroPage // mmm ($aa)
			}
			if (args.size == A32) {
				return args, fmt.Errorf("the address '$%x' is too large for 24-bits", args.value)
			}
		} else {
			args.size = p.abWidth // assume the default for the file, but resolve later
		}

		p.skipWhitespace()
		if p.peekChar() == ',' {
			p.skip(1)
			sym := p.peekChar()
			if (sym == 'X') || (sym == 'x') {
				if args.value <= 0x0FF {
					args.mode = modeIndexedIndirectX // mmm ($aa,X)
				} else {
					args.mode = modeAbsoluteIndexedIndirectX // mmm ($aaaa,X)
				}
			} else if (sym == 'Y') || (sym == 'y') {
				return args, errors.New("indexed indirect (addr,Y) is invalid")
			} else {
				return args, errors.New("indexed address (addr,R) but ,R not ,X or ,Y")
			}
			p.skip(1)
			if p.peekChar() != ')' {
				return args, errors.New("indexed indirect (addr,X) is missing closing ')'")
			}
		} else if p.peekChar() != ')' {
			return args, errors.New("indirect (addr) is missing closing ')'")
		}
		p.skip(1)

		p.skipWhitespace()
		if p.peekChar() == ',' {
			p.skip(1)
			p.skipWhitespace()
			sym := p.peekChar()
			if (sym == 'X') || (sym == 'x') {
				return args, errors.New("indexed indirect (addr),X is invalid")
			} else if (sym == 'Y') || (sym == 'y') {
				args.mode = modeIndirectIndexedY // mmm ($aa),Y
				p.skip(1)
			} else {
				return args, errors.New("indexed address (addr),R but ,R not ,X or ,Y")
			}
		}
	} else { // mmm $vvvv
		if (p.peekChar() == '@') {
			p.skip(1)
			symbol := p.nextAZ_az_09()
			args.symbol = symbol
			address, size, err := p.lookupVariable(p.currentCode, symbol)
			if (err != nil) {
				return args, fmt.Errorf("unknown variable '%s'", symbol)
			}
			args.value = address
			args.size |= size
			args.hasValue = true
		} else if p.isNextAZ() {
			symbol := p.nextAZ_az_09()
			args.symbol = symbol
			value, err := p.lookupConstant(symbol)
			if (err == nil) {
				if (p.peekChar() == '+') {
					p.skip(1)
					plus, _ := p.nextValue()
					value += plus
				}
				args.value = value
				args.hasValue = true
			} else {
				args.symbol = symbol
				args.hasValue = false

				// keyword "break" meaning 'goto end of the current block'
				if (strings.ToLower(symbol) == "break") {
					args.symbol = p.currentCode.name + "_end"
				}

				// optional +offset (or -offset) to the specified label
				p.skipWhitespace()
				sym := p.peekChar()
				if (sym == '+') || (sym == '-') {
					p.skip(1)
					p.skipWhitespace()
					value, err := p.nextValue()
					if (err != nil) {
						return args, err
					}
					if (sym == '-') {
						args.value -= value
					} else {
						args.value += value
					}
				}
			}
		} else {
			value, err := p.nextValue()
			if (err != nil) {
				return args, err
			}
			args.value = value
			args.hasValue = true
		}

		if (args.hasValue) {
			args.size = addressToPrefix(args.value)
			if (args.size == A32) {
				return args, fmt.Errorf("the address '$%x' is too large for 24-bits", args.value)
			}
		} else {
			args.size = p.abWidth // assume the default for the file, but resolve later
		}

		p.skipWhitespace()
		sym := p.peekChar()
		if sym == ',' {
			p.skip(1)
			p.skipWhitespace()
			sym := p.peekChar()
			if (sym == 'X') || (sym == 'x') {
				p.skip(1)
				if args.hasValue && (args.value <= 0x0FF) {
					args.mode = modeZeroPageX // mmm $aa,X
				} else {
					args.mode = modeAbsoluteX // mmm $aaaa,X
				}
			} else if (sym == 'Y') || (sym == 'y') {
				p.skip(1)
				if args.hasValue && (args.value <= 0x0FF) {
					args.mode = modeZeroPageY // mmm $aa,Y
				} else {
					args.mode = modeAbsoluteY // mmm $aaaa,Y
				}
			} else {
				return args, errors.New("indexed address mode but not ,X or ,Y")				
			}
		} else {
			if args.hasValue && (args.value <= 0x0FF) {
				args.mode = modeZeroPage // mmm $aa
			} else {
				args.mode = modeAbsolute // mmm $aaaa
			}
		}
	}

	return args, nil
}

/*
 *  Store the label as a blank instruction
 */
func (p *parser) parseLabel(label string) error {
	// Skip past the ':'
	p.skip(1)
	p.addInstructionLabel(label)
	p.skipWhitespaceAndEOL()

	return nil
}

/*
 *  Add the instruction to the (latest) block of code
 */
func (p *parser) addInstruction(mnemonic int, addressMode int, size int, hasValue bool, symbol string, value int) error {
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

	instr.hasValue = hasValue
	instr.symbol = symbol
	instr.symbolLC = strings.ToLower(symbol)

	instr.mnemonic = mnemonic
	instr.addressMode = addressMode
	instr.value = value

	// Label inherits the previous instruction address (or the block if its the first instruction)
	if mnemonic == 0 {
		instr.address = p.currentCode.endAddr
		return nil
	}

	// Find the mnemonic that matches the address mode and address/register size
	m := mnemonics[mnemonic]
	for k := range m.opcode {
		o := m.opcode[k]
		if (o.mode == addressMode) && (o.size == size) {
//@@@fmt.Printf("%s %t %x %d %x\n", m.name, instr.hasValue, instr.value, o.mode, o.size)
			instr.prefix = o.size
			instr.opcode = o.opcode
			instr.len = o.len
			instr.address = p.currentCode.endAddr
			p.currentCode.endAddr += o.len

			// Remember the size of the opcode for each of A, X, and Y registers
			switch (m.reg) {
			case REG_A:
				p.lastAsz = size
			case REG_X:
				p.lastXsz = size
			case REG_Y:
				p.lastYsz = size
			}

			return nil
		}
	}

	return fmt.Errorf("invalid address mode %s invalid for %s", addressModeStr(addressMode), strings.ToUpper(m.name))
}

/*
 *  Add a label
 */
func (p *parser) addInstructionLabel(label string) {
	p.addInstruction(0, unknownMode, A16, false, label, 0)
}


/*
 *  Return the mnemonic that matches the address mode and address/register size
 */
func lookupMnemonic(mmm string, addressMode int, size int) (int, *opcode, error) {
	for j := range mnemonics {
		if mnemonics[j].name == mmm {
			m := mnemonics[j]
			for k := range m.opcode {
				o := m.opcode[k]
				if (o.mode == addressMode) && (o.size == size) {
//@@@fmt.Printf("%s %t %x %d %x\n", m.name, instr.hasValue, instr.value, o.mode, o.size)
					return j, &o, nil
				}
			}
		}
	}
	
	return 0, nil, fmt.Errorf("mnemonic '%s' not found", mmm)
}

/*
 *  Explain the address mode in a string
 */
func unionAddressMode(mode1 int, mode2 int) int {
	switch (mode1 & R32) {
	case R08:
		switch (mode2 & R32) {
		case R08: return R08
		case R16: return R16
		case R24: return R24
		case R32: return R24 // as R32 isn't implemented
		}
	case R16:
		switch (mode2 & R32) {
		case R08: return R16
		case R16: return R16
		case R24: return R24
		case R32: return R24 // as R32 isn't implemented
		}
	case R24:
		return R24
	case R32:
		return R24 // as R32 isn't implemented
	}

	return R24
}

/*
 *  Explain the address mode in a string
 */
func addressModeStr(mode int) string {
	switch (mode) {
		default: return "unknown"
		case modeImplicit: return "mmm [implicit]"
		case modeImmediate: return "mmm #val [immediate]"
		case modeZeroPage: return "mmm $zz [zero page]"
		case modeZeroPageX: return "mmm $zz,X [zero page X]"
		case modeZeroPageY: return "mmm $zz,Y [zero page Y]"
		case modeRelative: return "bmm +dd [relative]"
		case modeAbsolute: return "mmm $aaaa [absolute]"
		case modeAbsoluteX: return "mmm $aaaa,X [absolute X]"
		case modeAbsoluteY: return "mmm $aaaa,Y [absolute Y]"
		case modeIndirect: return "mmm ($aaaa) [indirect]"
		case modeIndexedIndirectX: return "mmm ($aa,X) [indirect X]"
		case modeIndirectIndexedY: return "mmm ($aa),Y [indirect Y]"
		case modeIndirectZeroPage: return "mmm ($zz) [indirect zero page]"
		case modeAbsoluteIndexedIndirectX: return "mmm ($aaaa,X) [absolute indexed indirect X]"
	}
}
