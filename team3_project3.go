package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode/utf8"
)

type Instruction struct {
	typeofInstruction string
	rawInstruction    string
	linevalue         uint64
	programCnt        int
	opcode            uint64
	op                string
	rm                uint8
	rd                uint8
	rn                uint8
	im                int32
	shamt             int
	conditional       uint8
	instructionParsed string
	offset            int32
	registers         string
	address           uint16
	rawoffset         string
	shfcd             uint16
	field             uint32
	destReg           int // rd
	src1Reg           int // rn
	src2Reg           int //rm
	arg1Str           string
	arg2Str           string
	arg3Str           string
	rawValue          int64
}

type Control struct {
	programCnt       int           //program counter for next instruction to run (stored value must be multiplied by 4)
	registers        [32]int64     //array of 32 registers
	memory           []interface{} //data of both memory and instruction
	memoryDataHead   int           //program counter at start of memory data
	registerLocTable [][]int
	programCntStart  int   //start of program counter
	cache            Cache //cache storage
	fetchPaused      bool
}

var bufPreIssue = initQueue(4)
var bufPreMem = initQueue(2)
var bufPreALU = initQueue(2)
var bufPostMem = initQueue(1)
var bufPostALU = initQueue(1)

func main() {
	//flags for input output files
	var oFlag, iFlag = parseFlags()
	var oSim = *oFlag + "_sim.txt"

	//store input file data in array
	instructionList, control := ReadFile(*iFlag)
	control.programCntStart = 96
	control.programCnt = 96
	control.memory = instructionList
	control.fetchPaused = false

	//parse data and write to output file
	for i := range control.memory {

		if control.programCnt+i*4 >= control.memoryDataHead {
			break
		}

		var instruction = control.memory[i].(Instruction)

		linevalue, err := strconv.ParseUint(instruction.rawInstruction, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		if !(len(instruction.op) > 0) {
			instruction.linevalue = linevalue
			instruction.opcode, instruction.op, instruction.typeofInstruction = binToDec(instruction.linevalue)
			instruction = parse(instruction)
			readRegister(&instruction)
			control.memory[i] = instruction
		}
	}

	writeOutputFile(oFlag, instructionList, control.memoryDataHead, control.programCntStart)

	runSimulation(oSim, &control, instructionList)
}

// ***** Function Definitions *****//
// Function: ReadFile
// Usage: reads file and stores input data into an array of structs of dtype Instruction.
//
// Parameter(s):
//
//	-fileName string - name of file to read input from
//
// Returns:
//
//	-Array of structs (dtype Instruction)
func ReadFile(fileName string) ([]interface{}, Control) {

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var instructions []interface{}
	control := Control{}
	scanner := bufio.NewScanner(file)
	data := false
	i := 0
	d := -1
	for scanner.Scan() {
		instruc := scanner.Text()
		if instruc == "11111110110111101111111111100111" || data {
			if !data {
				newInstruct := Instruction{
					rawInstruction:    instruc,
					instructionParsed: "1 11111 10110 11110 11111 11111 100111",
					registers:         "",
					programCnt:        96 + (i * 4),
				}
				control.memoryDataHead = newInstruct.programCnt + 4
				instructions = append(instructions, newInstruct)
				data = true
			} else {
				newInstruct := Instruction{}
				if instruc[0:1] == "1" {
					newInstruct = Instruction{
						rawInstruction:    instruc,
						instructionParsed: instruc,
						programCnt:        96 + (i * 4),
						op:                twoCompliment(instruc),
					}

				} else {
					temp, err := strconv.ParseUint(instruc, 2, 32)
					if err != nil {
						fmt.Println(err)
					}
					newInstruct = Instruction{
						rawInstruction:    instruc,
						instructionParsed: instruc,
						programCnt:        96 + (i * 4),
						op:                fmt.Sprintf("%d", temp),
					}
				}

				newInstruct.rawValue = parse2CBinary(instruc)

				//set memory head at first memory collection
				if d == -1 {
					control.memoryDataHead = newInstruct.programCnt
				}

				instructions = append(instructions, newInstruct)
				d--
			}
		} else {
			newInstruct := Instruction{
				rawInstruction: instruc,
				programCnt:     96 + (i * 4),
			}
			instructions = append(instructions, newInstruct)

		}
		i++
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	return instructions, control
}

// Function: binToDec
// Usage: function binToDec converts given string of a binary value to
// actual binary form.
//
// Parameter(s):
//   - uint64 value of given instruction bit string
//
// Returns:
//   - opcode - 11 leftmost bits of the 32 bits
//   - instructType - R, B, IM, etc; the letter code for instruction type
//   - op - the exact operation to perform (Add, Add Immediate, Zero Branch Conditional, etc)
func binToDec(linevalue uint64) (uint64, string, string) {

	//shift by 21
	opcode := linevalue >> 21
	instructType := ""
	op := ""

	switch {
	case opcode == 0:
		op = "NOP"
		instructType = "N/A"
	case opcode <= 191 && opcode >= 160:
		op = "B"
		instructType = "B"
	case opcode == 1104:
		op = "AND"
		instructType = "R"
	case opcode == 1112:
		op = "ADD"
		instructType = "R"
	case opcode == 1160 || opcode == 1161:
		op = "ADDI"
		instructType = "I"
	case opcode == 1360:
		op = "ORR"
		instructType = "R"
	case opcode >= 1440 && opcode <= 1447:
		op = "CBZ"
		instructType = "CB"
	case opcode >= 1448 && opcode <= 1455:
		op = "CBNZ"
		instructType = "CB"
	case opcode == 1624:
		op = "SUB"
		instructType = "R"
	case opcode == 1672 || opcode == 1673:
		op = "SUBI"
		instructType = "I"
	case opcode >= 1684 && opcode <= 1687:
		op = "MOVZ"
		instructType = "IM"
	case opcode >= 1940 && opcode <= 1943:
		op = "MOVK"
		instructType = "IM"
	case opcode == 1690:
		op = "LSR"
		instructType = "R"
	case opcode == 1691:
		op = "LSL"
		instructType = "R"
	case opcode == 1984:
		op = "STUR"
		instructType = "D"
	case opcode == 1986:
		op = "LDUR"
		instructType = "D"
	case opcode == 1692:
		op = "ASR"
		instructType = "R"
	case opcode == 1872:
		op = "EOR"
		instructType = "R"
	case opcode == 2038:
		op = "BREAK"
	}

	return opcode, op, instructType //string name of instruction type (R, I, IM, B)
}

// Function: parseFlags
// Usage: parses -i and -o flags to find input and output file names.
//
// Parameter(s):
//   - none
//
// Returns:
//   - oFlag - pointer to string of output file name
//   - iFlag - pointer to string of input file name
func parseFlags() (oFlag *string, iFlag *string) {
	oFlag = flag.String("o", "", "output file")
	iFlag = flag.String("i", "", "input file")
	flag.Parse()

	if *oFlag == "" {
		log.Fatal("ERR: output file defined as ", *oFlag)
	}

	if *iFlag == "" {
		log.Fatal("ERR: input file defined as ", *oFlag)
	}

	return oFlag, iFlag
}

func writeOutputFile(oFlag *string, instructionList []interface{}, memoryDataStart int, programCountStart int) {
	//open output file
	outFile, errOut := os.Create(*oFlag + "_dis.txt")
	if errOut != nil {
		log.Fatalf("Error opening output file. err: %s", errOut)
	}
	defer outFile.Close()

	//string concatenation for printing to output file

	for i := range instructionList {

		var outputString string
		var concatString string
		var object = instructionList[i].(Instruction)
		concatString = fmt.Sprintf("%s\t", object.instructionParsed)
		outputString += concatString
		concatString = fmt.Sprintf("%s\t", strconv.Itoa(object.programCnt))
		outputString += concatString
		concatString = fmt.Sprintf("%s\t", object.op)
		outputString += concatString
		concatString = fmt.Sprintf("%s\n", object.registers)
		outputString += concatString

		if _, err2 := outFile.Write([]byte(outputString)); err2 != nil {
			panic(err2)
		}
	}
}

// ***** Function Definitions *****//
// Function: parse
// Usage: Parse the raw instruction and find the registers
//
// Parameter(s):
//   - structure 'instruct'
//
// Returns:
// structure 'instruct'
func parse(instruct Instruction) Instruction {
	var parse1, parse2, parse3, parse4, parse5 string

	switch {
	case instruct.typeofInstruction == "R":
		parse1 = instruct.rawInstruction[0:11]
		parse2 = instruct.rawInstruction[11:16]
		parse3 = instruct.rawInstruction[16:22]
		parse4 = instruct.rawInstruction[22:27]
		parse5 = instruct.rawInstruction[27:32]

		// finds the register for rm
		temp, err := strconv.ParseUint(parse2, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rm = uint8(temp)
		temp, err = strconv.ParseUint(parse3, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.shamt = int(temp)
		// finds the register for rn
		temp, err = strconv.ParseUint(parse4, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rn = uint8(temp)

		// finds the register for rd
		temp, err = strconv.ParseUint(parse5, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rd = uint8(temp)

		instruct.instructionParsed = parse1 + " " + parse2 + " " + parse3 + " " + parse4 + " " + parse5

	case instruct.typeofInstruction == "D":
		parse1 = instruct.rawInstruction[0:11]
		parse2 = instruct.rawInstruction[11:20]
		parse3 = instruct.rawInstruction[20:22]
		parse4 = instruct.rawInstruction[22:27]
		parse5 = instruct.rawInstruction[27:32]

		temp, err := strconv.ParseUint(parse2, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.address = uint16(temp)

		temp, err = strconv.ParseUint(parse4, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rn = uint8(temp)

		temp, err = strconv.ParseUint(parse5, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rd = uint8(temp)

		instruct.instructionParsed = parse1 + " " + parse2 + " " + parse3 + " " + parse4 + " " + parse5

	case instruct.typeofInstruction == "I":
		parse1 = instruct.rawInstruction[0:10]
		parse2 = instruct.rawInstruction[10:22]
		parse3 = instruct.rawInstruction[22:27]
		parse4 = instruct.rawInstruction[27:32]

		temp2 := parse2CBinary(parse2) //strconv.ParseInt(parse2, 2, 32)
		instruct.im = int32(temp2)
		instruct.rawoffset = parse2
		temp, err := strconv.ParseUint(parse3, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rn = uint8(temp)

		temp, err = strconv.ParseUint(parse4, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rd = uint8(temp)

		instruct.instructionParsed = parse1 + " " + parse2 + " " + parse3 + " " + parse4

	case instruct.typeofInstruction == "B":
		parse1 = instruct.rawInstruction[0:6]
		parse2 = instruct.rawInstruction[6:32]
		temp := parse2CBinary(parse2)
		instruct.offset = int32(temp)
		instruct.rawoffset = parse2
		instruct.instructionParsed = parse1 + " " + parse2

	case instruct.typeofInstruction == "CB":
		parse1 = instruct.rawInstruction[0:8]
		parse2 = instruct.rawInstruction[8:27]
		parse3 = instruct.rawInstruction[27:32]
		temp := parse2CBinary(parse2)
		instruct.offset = int32(temp)
		instruct.rawoffset = parse2
		temp2, err := strconv.ParseUint(parse3, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.conditional = uint8(temp2)
		instruct.instructionParsed = parse1 + " " + parse2 + " " + parse3

	case instruct.typeofInstruction == "IM":
		parse1 = instruct.rawInstruction[0:9]
		parse2 = instruct.rawInstruction[9:11]
		parse3 = instruct.rawInstruction[11:27]
		parse4 = instruct.rawInstruction[27:32]

		temp, err := strconv.ParseUint(parse4, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.rd = uint8(temp)

		temp, err = strconv.ParseUint(parse2, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.shamt = int(temp)

		temp, err = strconv.ParseUint(parse3, 2, 32)
		if err != nil {
			fmt.Println(err)
		}
		instruct.address = uint16(temp)
		instruct.instructionParsed = parse1 + " " + parse2 + " " + parse3 + " " + parse4
	case instruct.typeofInstruction == "N/A":
		instruct.instructionParsed = "00000000000000000000000000000000"
	}

	return instruct

}

func readRegister(s1 *Instruction) {
	switch {
	case s1.typeofInstruction == "B":
		if s1.rawoffset[0:1] == "1" {
			s1.registers = fmt.Sprintf("#%s", twoCompliment(s1.rawoffset))
		} else {
			s1.registers = fmt.Sprintf("#%d", s1.offset)
		}

	case s1.typeofInstruction == "R":
		switch {
		case s1.op == "LSL" || s1.op == "LSR":
			s1.registers = fmt.Sprintf("R%d, R%d, #%d", s1.rd, s1.rn, s1.shamt)
		default:
			s1.registers = fmt.Sprintf("R%d, R%d, R%d", s1.rd, s1.rn, s1.rm)
		}
	case s1.typeofInstruction == "I":
		if s1.rawoffset[0:1] == "1" {
			s1.registers = fmt.Sprintf("R%d, R%d, #%s", s1.rd, s1.rn, twoCompliment(s1.rawoffset))
		} else {
			s1.registers = fmt.Sprintf("R%d, R%d, #%d", s1.rd, s1.rn, s1.im)
		}

	case s1.typeofInstruction == "CB":
		if s1.rawoffset[0:1] == "1" {
			s1.registers = fmt.Sprintf("R%d, #%s", s1.conditional, twoCompliment(s1.rawoffset))
		} else {
			s1.registers = fmt.Sprintf("R%d, #%d", s1.conditional, s1.offset)
		}

	case s1.typeofInstruction == "IM":
		if s1.op == "MOVZ" {
			s1.registers = fmt.Sprintf("R%d, %d, LSL %d", s1.rd, s1.address, s1.shamt*16)
		} else {
			s1.registers = fmt.Sprintf("R%d, %d, LSL %d", s1.rd, s1.address, s1.shamt*16)
		}
	case s1.typeofInstruction == "D":
		s1.registers = fmt.Sprintf("R%d, [R%d, #%d]", s1.rd, s1.rn, s1.address)
	}

}

func twoCompliment(binary string) string {

	// flips the 1s to 0s and vice versa
	binaryString := trimFirstRune(binary)
	for i := 0; i < len(binaryString); i++ {
		if binaryString[i:i+1] == "0" {

			binaryString = binaryString[:i] + "1" + binaryString[i+1:]
		} else if binaryString[i:i+1] == "1" {
			binaryString = binaryString[:i] + "0" + binaryString[i+1:]
		}
	}
	temp, err := strconv.ParseUint(binaryString, 2, 32)
	if err != nil {
		fmt.Println(err)
	}
	temp += 1
	binaryString = fmt.Sprintf("-%d", temp)

	return binaryString
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func (c *Control) runInstruction(i Instruction) int64 {

	var branchOperation = false
	var returnValue int64

	if !((i.rm >= 0 && i.rm <= 31) ||
		(i.rd >= 0 && i.rd <= 31) ||
		(i.rn >= 0 && i.rn <= 31)) {

	} else {

		switch {
		case i.op == "ORR":
			returnValue = c.registers[i.rn] | c.registers[i.rm]
		case i.op == "AND":
			returnValue = c.registers[i.rn] & c.registers[i.rm]
		case i.op == "ADD":
			returnValue = c.registers[i.rn] + c.registers[i.rm]
		case i.op == "SUB":
			returnValue = c.registers[i.rn] - c.registers[i.rm]
		case i.op == "EOR":
			returnValue = c.registers[i.rn] ^ c.registers[i.rm]
		case i.op == "LSL":
			returnValue = c.registers[i.rn] << i.shamt
		case i.op == "LSR":
			returnValue = c.registers[i.rn] >> i.shamt
		case i.op == "ASR":
			var shift = 0

			if c.registers[i.rm]%2 == 0 {
				shift = shift + 1 // 1
			}

			returnValue = c.registers[i.rn] >> (16 * shift)
		case i.op == "MOVZ":
			returnValue = int64(i.address) << (i.shamt * 16)
		case i.op == "MOVK":
			returnValue = int64(uint16(c.registers[i.rd])) ^ int64(i.address)<<(i.shamt*16)
		case i.op == "LDUR":
			// fmt.Printf("Rd: %d\n Rm: %d\nValue: %d\nOffset:%d\n", i.rd, i.rm, c.registers[i.rm], i.offset)
			var registerDestValue = c.registers[i.rn]
			var memoryIndex = ((registerDestValue + int64(i.address*4)) - int64(c.memoryDataHead)) / 4

			if memoryIndex < 0 || memoryIndex > 2048 {
				break
			}

			c.memory = memoryCheck(c.memory, int(memoryIndex))

			returnValue = c.memory[memoryIndex].(int64)
		case i.op == "STUR":
			var registerDestValue = c.registers[i.rn]
			var memoryIndex = int32(int32(registerDestValue+int64(i.address*4))-int32(c.memoryDataHead)) / 4

			if memoryIndex < 0 || memoryIndex > 2048 {
				break
			}

			c.memory = memoryCheck(c.memory, int(memoryIndex))

			c.memory[memoryIndex] = c.registers[i.rd]

			return -1
		case i.op == "B": //might be obsolete
			c.programCnt += int(i.offset * 4)
			branchOperation = true
		case i.op == "CBZ":
			if c.registers[i.conditional] == 0 {
				c.programCnt += int(i.offset * 4)
				branchOperation = true
			}
		case i.op == "CBNZ":
			if c.registers[i.conditional] != 0 {
				c.programCnt += int(i.offset * 4)
				branchOperation = true
			}
		case i.op == "ADDI":
			returnValue = int64(int32(c.registers[i.rn]) + i.im)
		case i.op == "SUBI":
			returnValue = int64(int32(c.registers[i.rn]) - i.im)
		case i.op == "NOP":
			break
		}
	}

	if c.programCnt >= c.memoryDataHead || c.programCnt < c.programCntStart {
		branchOperation = false
	}

	if !branchOperation {
		c.programCnt = i.programCnt
	} else {
		c.programCnt -= 4
	}

	return returnValue
}

func runSimulation(outputFile string, c *Control, il []interface{}) {
	outFile, errOut := os.Create(outputFile)
	if errOut != nil {
		log.Fatalf("Error opening output file. err: %s", errOut)
	}

	var runControlLoop = true
	var outputString, concatString string
	var cycleNumber = 1
	//compute instruction loop
	for runControlLoop {
		var programCountPrevious = c.programCnt
		var listIndexFromPC = (c.programCnt - c.programCntStart) / 4
		var breakpoint = ((c.memoryDataHead - c.programCntStart) / 4) - 1

		if listIndexFromPC < 0 {
			listIndexFromPC = 0
		}

		var currentInstruction = il[listIndexFromPC].(Instruction)

		c.writeBack(&bufPostMem, &bufPostALU)
		c.memProcess(&bufPreMem)
		c.aluProcess(&bufPreALU)
		c.issueProcess(&bufPreIssue, &bufPreMem, &bufPreALU)
		c.fetch(&bufPreIssue)

		outputString = ""
		concatString = "====================\n"
		outputString += concatString
		concatString = fmt.Sprintf("cycle:%d\t%s - %s\t%s\n", cycleNumber, strconv.Itoa(programCountPrevious),
			currentInstruction.op, currentInstruction.registers)
		outputString += concatString
		concatString = fmt.Sprintf("Pre-Issue Buffer:\n")
		outputString += concatString

		var entryAmt = 0

		if !bufPreIssue.isEmpty() {
			for _, bufInstr := range bufPreIssue.data {
				var inst = bufInstr[0].(Instruction)
				concatString = fmt.Sprintf("\tEntry %d:\t[%d , %s]", entryAmt, inst.programCnt, inst.op)
				outputString += concatString
				entryAmt += 1
			}
		} else {
			concatString = fmt.Sprintf("EMPTY")
			outputString += concatString
		}

		concatString = fmt.Sprintf("\nPre-ALU Buffer:\n")
		outputString += concatString

		entryAmt = 0

		if !bufPreALU.isEmpty() {
			for _, bufInstr := range bufPreALU.data {
				var inst = bufInstr[0].(Instruction)
				concatString = fmt.Sprintf("\tEntry %d:\t[%d , %s]", entryAmt, inst.programCnt, inst.op)
				outputString += concatString
				entryAmt += 1
			}
		} else {
			concatString = fmt.Sprintf("EMPTY")
			outputString += concatString
		}

		concatString = fmt.Sprintf("\nPost-ALU Buffer:\n")
		outputString += concatString

		entryAmt = 0

		if !bufPostALU.isEmpty() {
			for _, bufInstr := range bufPostALU.data {
				var inst = bufInstr[0].(Instruction)
				concatString = fmt.Sprintf("\tEntry %d:\t[%d , %s]", entryAmt, inst.programCnt, inst.op)
				outputString += concatString
				entryAmt += 1
			}
		} else {
			concatString = fmt.Sprintf("EMPTY")
			outputString += concatString
		}

		concatString = fmt.Sprintf("\nPre-MEM Buffer:\n")
		outputString += concatString

		entryAmt = 0

		if !bufPreMem.isEmpty() {
			for _, bufInstr := range bufPreMem.data {
				var inst = bufInstr[0].(Instruction)
				concatString = fmt.Sprintf("\tEntry %d:\t[%d , %s]", entryAmt, inst.programCnt, inst.op)
				outputString += concatString
				entryAmt += 1
			}
		} else {
			concatString = fmt.Sprintf("EMPTY")
			outputString += concatString
		}

		concatString = fmt.Sprintf("\nPost-MEM Buffer:\n")
		outputString += concatString

		entryAmt = 0

		if !bufPostMem.isEmpty() {
			for _, bufInstr := range bufPostMem.data {
				var inst = bufInstr[0].(Instruction)
				concatString = fmt.Sprintf("\tEntry %d:\t[%d , %s]", entryAmt, inst.programCnt, inst.op)
				outputString += concatString
				entryAmt += 1
			}
		} else {
			concatString = fmt.Sprintf("EMPTY")
			outputString += concatString
		}

		//write registers and data
		var runLoop = true
		var iterator = 0
		var registerMax = 32
		var dataMax = len(c.memory)
		concatString = fmt.Sprintf("\n\nregisters:\nr00\t")
		outputString += concatString

		//write registers
		for runLoop {
			if iterator >= registerMax {
				runLoop = false
			} else {
				concatString = fmt.Sprintf("%d\t", c.registers[iterator])
				outputString += concatString

				if ((iterator+1)%8 == 0) && (iterator < registerMax-1) {
					concatString = fmt.Sprintf("\nr%02d\t", iterator+1)
					outputString += concatString
				}

				iterator++
			}
		}

		//write cache
		entryAmt = 0
		concatString = fmt.Sprintf("\nCache   [(valid bit, dirty bit, int(tag))<word0,word1>]\n")
		outputString += concatString

		for _, j := range c.cache.CacheSets {
			var entryIterator = 0
			concatString = fmt.Sprintf("Set %d: LRU=%d\n", entryAmt, c.cache.LruBits[entryAmt])
			outputString += concatString

			for _, r := range j {
				var word1String = "EMPTY"
				var word2String = "EMPTY"
				if r.word1 != nil {
					word1String = strconv.Itoa(r.word1.(Instruction).programCnt)
				}
				if r.word2 != nil {
					word2String = strconv.Itoa(r.word2.(Instruction).programCnt)
				}

				concatString = fmt.Sprintf("\tEntry %d: [(%d, %d, %d)<%s,%s>]\n", entryIterator, r.valid, r.dirty,
					r.tag, word1String, word2String)
				outputString += concatString
				entryIterator += 1
			}
		}

		iterator = (c.memoryDataHead - c.programCntStart) / 4

		//if there's data in memory to print
		if len(c.memory)-iterator > 0 {
			concatString = fmt.Sprintf("\n\ndata:\n%d\t", c.memoryDataHead)
			outputString += concatString

			runLoop = true

			for runLoop {
				if iterator >= dataMax {
					runLoop = false
				} else {
					concatString = fmt.Sprintf("%d\t", c.memory[iterator].(Instruction).rawValue)
					outputString += concatString

					if (iterator+1)%8 == 0 {
						concatString = fmt.Sprintf("\n%d\t", c.memoryDataHead+iterator*4)
						outputString += concatString
					}

					iterator++
				}
			}
		} else {
			concatString = fmt.Sprint("\n\ndata:\nEMPTY")
			outputString += concatString
		}

		concatString = "\n"
		outputString += concatString

		if _, err2 := outFile.Write([]byte(outputString)); err2 != nil {
			panic(err2)
		}

		cycleNumber++

		if listIndexFromPC >= breakpoint {
			runControlLoop = false
		}

	}

	err := outFile.Close()
	if err != nil {
		fmt.Print("ERR") //return Control{}
	}
}

func memoryCheck(list []interface{}, index int) []interface{} {
	for len(list) <= index {
		list = append(list, 0)
	}

	return list
}

func parse2CBinary(binaryString string) int64 {
	var sign = false //false = positive
	if binaryString[0] == '1' {
		sign = true
	}

	var binaryValue = binaryString[1:]
	var tempString = ""
	var tempRune = "0"
	var iterator = 0

	if sign {
		for iterator < len(binaryValue) {
			if binaryValue[iterator] == '0' {
				tempRune = "1"
			} else {
				tempRune = "0"
			}

			tempString = tempString + tempRune

			iterator++
		}

		var value, err = strconv.ParseUint(tempString, 2, 32)
		if err != nil {
			log.Fatalf("binary convert error : %d", err)
		}

		return -int64(value + 1)
	}

	var value, err = strconv.ParseUint(binaryValue, 2, 32)
	if err != nil {
		log.Fatalf("binary convert error : %d", err)
	}

	return int64(value)
}
