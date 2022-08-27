package pomme

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
    "strings"
)

// Structure to hold the parsed data
type parser struct {
	b 			[]uint8			// file buffer
	end			int				// buffer length-1
	i			int				// index into the file
	n			int				// line number

	abWidth		int 			// A16 for lowest code address @<$FFFF or A24 @>=10000

	cnst		*cnst			// linked list of constants
	lastCnst	*cnst

	global		*vrbl			// linked list of constants
	lastGlobal	*vrbl

	code		*codeBlock		// linked list of subroutine blocks
	lastCode	*codeBlock

	lastAsz		int				// the size the last time A was touched
	lastXsz		int				//                      ^ X
	lastYsz		int				//                      ^ Y

	data		*dataBlock		// linked list of data items
	lastData	*dataBlock
}

// Linked list of constants
type cnst struct {
	next		*cnst			// next in the linked list
	name		string
	nameLC		string
	value		int
}

// Linked list of global variables
type vrbl struct {
	next		*vrbl			// next in the linked list
	name		string
	nameLC		string
	address		int
	size		int
}

// Linked list of code blocks
type codeBlock struct {
	next		*codeBlock		// next in the linked list
	prev		*codeBlock		// previous in the linked list
	startAddr	int
	endAddr		int
	name		string
	nameLC		string

	vrbl		*vrbl			// linked list of local-to-the-block variables
	lastVrbl	*vrbl

	instr		*instruction	// linked list of instructions (parsed but not yet encoded)
	lastInstr	*instruction
}

// Linked list of instructions
type instruction struct {
	next		*instruction	// next in the linked list
	prev		*instruction	// previous in the linked list
	hasValue	bool
	symbol		string
	// code
	mnemonic	int
	addressMode	int
	prefix		int
	opcode		int
	value		int
	len			int
	address		int
	// optional comment
	comment		*comment
	// optional expression
	expr		*expression
	// optional block from keyword
	ifloop		*ifloop
}

const (
	IS_IF = iota
	IS_LOOP
	IS_FOR
	IS_DO
)

// Sub-block created by a keyword
type ifloop struct {
	keyword		int
	upDown		bool
	startAddr	int
	endAddr		int
}

// Unit within an instruction
type comment struct {
	comment		string
}

// Unit within an expression
type eunit struct {
	location	int
	addrval		int
	size		int
}

// Details of an expressions
type expression struct {
	dest		eunit
	src1		eunit
	src2		eunit
	equalOp		int
	op			int
}

// Linked list of code blocks
type dataBlock struct {
	next		*dataBlock		// next in the linked list
	prev		*dataBlock		// previous in the linked list
	startAddr	int
	endAddr		int
	name		string
	nameLC		string

	data		*data			// linked list of data entries
	lastData	*data
}

// Linked list of data
type data struct {
	next		*data		// next in the linked list
	prev		*data		// previous in the linked list
	size		int
	value		int
	string		string
	len			int
	address		int
}
const DSTRING = -1 // size of data when the value is a string


/*
 *  Start of the assembler/compiler
 *
 *  -flags input1[ input2 ... inputN]
 */
func Pomme() {
	// Parse the flags
	oflag := flag.String("o", "", "filename of the compiled code")
	lflag := flag.String("l", "", "filename of the compiled listing")

	flag.Parse()

	// The list of input files comes after the flags
	filename := flag.Arg(0)
	if filename == "" {
		fmt.Printf("ERROR: No file was specified\n");
		return
	}

	// Generate the output name from the first filename (if not specified)
	outname := *oflag
	if outname == "" {
		dot := strings.LastIndex(filename, ".")
		if (dot < 0) {
			outname = filename + ".out"
		} else {
			outname = filename[:dot] + ".out"
		}
	}

	// Generate the output name from the first filename (if not specified)
	listname := *lflag
	if listname == "" {
		dot := strings.LastIndex(filename, ".")
		if (dot < 0) {
			listname = filename + ".lst"
		} else {
			listname = filename[:dot] + ".lst"
		}
	}

	// Load all the files into memory
	filenames := flag.Args()
	files := make([][]uint8, len(filenames), len(filenames))
	for i := range filenames {
		fmt.Printf("READ %s\n", filenames[i])

		var err error
		files[i], err = readFile(filenames[i])
		if err != nil {
			fmt.Printf("ERROR: %v\n", err);
			return
		}
	}

	// Generate the output file
	fmt.Printf("CREATE %s\n", outname)
	out, err := os.Create(outname)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err);
		return
	}

	// Generate the listing file
	fmt.Printf("CREATE %s\n", listname)
	listing, err := os.Create(listname)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err);
		return
	}

	// Parse each file
	var p parser
	p.abWidth = A24				// default is 24-bit addresses
	for i := range files { 
		err := p.parseFile(filenames[i], files[i])
		if err != nil {
			return
		}
	}

	// Output the machine code and listing
	err = p.generateCode(out, listing)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err);
		return
	}

	fmt.Printf("ASSEMBLY COMPLETE\n")
	defer listing.Close()
	defer out.Close()
}

/*
 *  Read the whole file into memory
 */
func readFile(filename string) ([]uint8, error) {
	var file io.Reader
	diskFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer diskFile.Close()
	file = diskFile

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}	

	return data, nil
}
