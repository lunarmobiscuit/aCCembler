# aCCembler

*More than an assembler, less than a compiler...*

A programming language that mixes the mnemonics of assembly with some of the syntax of C, allowing for structured code with well-defined subroutines, higher-level constructs like IF, LOOP, DO...WHILE, BREAK, CONTINUE, and RETURN that compile efficiently down into the minimal machine code, plus constants, globals, and local variables.

Or in short, this is what happens when a C-coder is faced with writing thousands of lines of 6502 assembly.  C compilers for the 6502 are too high level, losing the beauty of that CPUs addressing modes.  At the same time, macro assemblers are not good enough, providing some higher-level constructs but not the structure that C provides.

## 6502 Mnemonics

All of the tradtional 6502 mnemonics are available, unchaged from the 1970s.

Syntactially, the one big change from traditional assemblers is that mnemonics can start in the first character of a line.  There doesn't need to be the obligitory space or tab to differeniate menonics from labels.

## Addressing Modes

All of the tradtional 6502 addressing modes are avaialble, unchaged from the 1970s, except for one change in syntax.  Branch distances (numeric or symbolic) must be prefixed by a '+'' or '-'.  E.g. `BCC -loop` to branch back to the `loop:` label or `bne +skip` to branch forward to the yet-defined `skip:` label.

The assembler doesn't actually care whether you use a '+'' or '-' prefix, but it is useful later when reading and maintaining the code to help see the pattern of the branching.

## Labels

Labels can be included on any line.  Labels MUST end in a ':'.

Syntactially, the one big change from traditional assemblers is that labels don't have to start on the first character of a line.  Labels can be prefixed by spaces or tabs when that makes the code easier to read.

Given the subroutine blocks and C-like constructs, labels are less common than in tradtiional assembly.

Labels are local to the current block.  I.e., JSR, JMP, and Bxx from one block can not jump to labels in another block.

## CONST *name* = *value*

Constants are defined with the CONST keyword.  These can be defined anywhere in the file, but they must be defined before they are used.  The name is any alphanumeric string (a string starting with A-Za-z_ then a string of A-Za-z0-9_ characters).  The value is either decimal (no prefix) or hexidemical (prefixed either with $ or 0x).

## GLOBAL *name* = @*address*[.width]

Global variables are named addresses.  These can be defined anywhere in the file outside of another block, but they must be defined before they are used.  The name is any alphanumeric string (no repeats with constants).  The value is prefixed by a `@` (which in the aCCembler denotes an address) followed by either decimal (no prefix) or hexidemical (prefixed either with $ or 0x).

The *.width* suffix to the address is optional.  See below for the 65C2402 augmentation to the 6502 capabilities.

## SUB *name* [*address*] { ... }

Blocks of code are defined with the keyword SUB, followed by the name of the subroutine, optionally followed by an address, then a '{', the code, and ended with a closing '}'.  E.g. `SUB main @$1000 { RTS }`, but with newlines after the '{' and after the `RTS` (at least until the parser is perfected)

The name acts like a label (but specified without the trailing ':'), usable in JSR, JMP, and Bxx menomics.

The address replaces the tradtional ORG assembler directive.  If the address is not specified, then the new SUB block's address beigns at the next byte following the previous SUB block.  (The first block's address defaults to $0000 if no address is specified)

## VAR *name* = @*address*[.width]

Local variables are named addresses within a code block.  These act the same as GLOBAL variables, but are only accessible within the block where they are defined, or any sub-block therein.

## IF *bool* { ... } [ELSE { ... }]

Rather than the assembly pattern of `Bxx +skip`, where bxx is the opposite of the pattern being checked and the `skip:` label pointing to the code not being run, `IF` is specified like the `if` keyword in C.

The boolean expression can be as simple as `==` or or `!=` or `-`, compiled directly into a machine `BEQ`, `BNE`, or `BMI` instruction, using the current flags, without adding any `CMP` instruction.  Or the boolean can compare against a value, e.g. `== 0` or `< $A0`, which with compile into a `CMP` followed by the proper `Bxx`.

The optional ELSE clause is just like `else` in C, with the code in that block run only if the boolean is not true.

The aCCembler generates labels in the .lst output that are used in the compiled logic.  The actual code generated is the same as hand-assembly, reversing the logic and jumping past the block of code.  The difference is that anyone reading the code need only see the pattern of *IF true { do this }* with no labels to parse.

## LOOP { ... }

The code in the block runs in a perpetual loop.  The compiled code generates a label for the start of the loop and inserts a `BRA -loop` instruction at the end of the loop, the same as hand-assembly.  The difference is that the keyword LOOP makes it obvious where the loop starts and stops.

The loop can be exited with a Bxx or JMP mnemonic, but better done with an IF that includes a BREAK keyword.

## BREAK

Break out of the most recent loop.  Most recent as LOOP, FOR, DO, etc. can be nested.  The compiled code simply inserts a label after the last instruction in the most recent loop and generates a `BRA +end` instruction to jump out of the loop.

TODO: Add BREAK 2, BREAK 3, BREAK n, to allow jumping out of nested loops.

## CONTINUE

Jump back to the top of the most recent loop.  Most recent as LOOP, FOR, DO, etc. can be nested.  The compiled code generates a `BRA +start` instruction to the top of the loop, to the label it had already generated as part of the loop.

The typical use case is to have a series of IF's inside a loop, using CONTINUE to avoid checking the other IFs once one is found that triggers a behavior.

TODO: Add CONTINUE 2, CONTINUE 3, CONTINUE n, to allow jumping up through nested loops.

## DO { ... } WHILE (*bool*)

Similar to LOOP, except with a boolean test at the end of the loop.  The boolean works like in an IF, including syntax like `WHILE (!=)` to check the CPU flags.  The last line in the DO block can be handcoded CMP or CMX or CMY, or that can be generated automatically if the boolean test is more complicated.  E.g. `WHILE (@V < 255)` will generate the instructions for `LDA @V` followed by `CMP 255` and `BCC -loop`.

TODO: Allow for more complex boolean expressions, including compound expressions using && and ||.  For now, if you need more tests to break out of the loop, use IF and BREAK.

## FOR *reg/var* = *start* [DOWN] TO *end* { ... }

Similar to FOR in BASIC, except the loop variable can be specified as `X` or `Y` to use the X or Y register, or any previously defined GLOBAL or VAR.  E.g. `FOR X = 0 TO 255` or `FOR @I = 10 DOWN TO 1`.

In the case of iterating over X or Y, the generated code is the same as hand-coded assembly.  For variables, the generated code is as tight as possible, but without stomping on X or Y, even if that would be more efficient.

## 65C2402

Beyond adding struture to assembly code, the other reason the aCCembler was written was that there was no assembler or compiler availalbe for the mythical 65C2402 CPU (https://github.com/lunarmobiscuit/verilog-65C2402-fsm and https://github.com/lunarmobiscuit/iz6502).

This is a mythical series of variations of the 65C02 that starts by adding a 24-bit address bus, then growing the A/X/Y/S registers from 8-bit to 16-bit and 24-bit.  The 65C24T8 then adds hardward threads.

It was not too difficult to get DASM (https://github.com/lunarmobiscuit/dasm) to handle 24-bit addresses, but it was just as easy to write an assembler from scratch than to re-write half of DASM to handle registers that are not 8-bit wide.

## Prefix codes

The 65C2402 extends the capabilities of the 6502 using "prefix codes".  These are a 1-byte opcodse that by themselves do nothing.  It simply informs the CPU that the next instruction will include or manage a 24-bit address, or that the next instruction shoudl treat the A/X/Y register as 16-bit or 24-bit wide.

The idea of prefix codes comes from the Z80.  They adds one extra byte to the resulting machine code but do not consume large swathes of precious opcode space adding any new addressing modes or new registers.  This code expansion is more than made up for by the new capabilities, typically shrinking the overall code by 15%-20%.

Using prefix codes is far simpler than the collection of modes used in the 65C816.  All registers are separately 8-bit or 16-bit or 24-bit for each line of assembly code, and all addresses can be specified as one byte (zero page), two bytes (<= 64K), or 3 bytes (24-bit).

And with prefix codes, the 65C2402 is 100% backward compatible with the 65C02 (not including the Rockwell opcodes), allowing legacy code to run completely unchanged.

## 16-bit/24-bit addresses

24-bit addresses are enabled by using the A24 prefix code before any other opcode.  The aCCembler inserts this automatically whenever you specify an address wider than 16-bits, or whenever you access a variable that is specified with an `.a24` suffix.  E.g. `GLOBAL HIGH = @$10000` or `VAR TF = @$200.a24`.

## 8-bit/16-bit/24-bit registers

The A, X, and Y registers are treated as 8-bit, 16-bit, or 24-bit per instruction.  These can be specified with the R16 and R24 prefix codes (no prefix code exists for R08 as by default the registers are 8-bits wide), but the aCCembler inserts this automatically whenever you specify a value wider than 8-bits, or when you specify a menmonic with a `.b`, `.w`, `.16`, `.t` or `.24` suffix.  E.g. `LDA #$1234` prepends the machine code with R16 without any explicit suffix on the LDA.  E.g. `LDA #$112233` prepends the R24 suffix.

The explicit suffixes are needed when processing data in the registers.  E.g. `LDA.w $200` loads 16-bits (little endian) from address $200, then `ADC.w #3` requires that `.w` suffix to specify that the results of the addition are still 16-bits wide.

If you forget the suffix when processing the values, the results in the registers WILL get truncated.  Or in other words, the registers do not keep the bits above the width required by an opcode.  Those bits are zeroed out.

This is the tricky part about variable width registers, but also their great flexibility.  You can LDA.w a 16-bit value and then STA.t three bytes, with confidence that the top 8 bits are all zero.  Or you can LDA.t a 24-bit value and if you only need the compare the bottommost 8-bits, CMP.b (or CMP with no prefix) will do that.

With wider regisers, you can use X or Y to loop up to 16,777,216 times, instead of just 256 times.  You can also use X or Y to hold an entire address, instead of having to move addresses around one byte at a time.

## A work in progress

The aCCembler is very much a work in progress.  Its features are being written as-needed, to match the code required to create an emulated Apple II4, a mythical computer that should have been between the IIplus and IIe, with the 24-bit addresses (avoiding all the IIe nonsense with a dozen swappable pages of RAM and ROM).

# Example

The markdown indents are much too deep, but the following nonsense code shows off the structure that this syntax allows along with the use of varialbes and use of .w and .t suffixes:

```
const FIVE = 5
global HERE = @$200
global THERE = @$300

sub Start @$FF0000 {
	lda.w @HERE
	sta.w @HERE
	FOR X = 0 TO 8 {
		lda.w @HERE,X
		IF == {
			lsr.w
			lsr.w
			lsr.w
			lsr.w
			clc
			adc.t #FIVE
			sta.t @THERE
			BREAK
		}
		ELSE {
			LOOP {
				lda.t @THERE
				IF - {
					BREAK
				}
			}
			dec.t @THERE
		}
	}
}
```

