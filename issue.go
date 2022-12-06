package main

import "fmt"

func (c *Control) issueProcess(bufPreIssue *Queue, bufPreMem *Queue, bufPreALU *Queue) {
	for _, i := range bufPreIssue.data {
		fmt.Print(i)
	}
	/*
		It can issue up to two instructions, out of order, per clock cycle.  When an instruction is issued,
		it moves out of the pre-issue buffer and into either the pre-mem buffer or the pre-ALU buffer.
		The issue unit searches from entry 0 to entry 3 (IN THAT ORDER) of the pre-issue buffer and issues instructions if:
		No structural hazards exist (there is room in the pre-mem/pre-ALU destination buffer)
		No RAW hazards exist with active instructions (all operands are ready)
		Store instructions must be issued in order
	*/

	var value, err = bufPreIssue.head()
	var errorsExist = false

	if err != -1 { //if buffer isn't empty
		var instruction = value[0].(Instruction)
		var instructionType = instruction.typeofInstruction
		var registerPipelineAdd [][]int

		//if memory instruction
		if instructionType == "D" {
			//if memory buffer can take data
			if cap(bufPreMem.data) < bufPreMem.maxSize {
				switch instruction.op {
				case "LDUR":
					errorsExist = c.findErrors(int(instruction.rn))
					registerPipelineAdd = append(registerPipelineAdd, []int{int(instruction.rn), 0})
					break
				case "STUR": // check if value that is being stored and offset register are modified
					errorsExist = c.findErrors(int(instruction.rn)) || c.findErrors(int(instruction.rd))
					break
				}

				//if no raw hazards exist,
				if !errorsExist {
					bufPreMem.enqueue(value)

					for _, i := range registerPipelineAdd {
						c.registerLocTable = append(c.registerLocTable, i)
					}
				}
			}
		} else { //if other type (ALU usage)
			//if memory buffer can take data
			if cap(bufPreALU.data) < bufPreALU.maxSize {
				switch instruction.typeofInstruction {
				case "R":
					errorsExist = c.findErrors(int(instruction.rn)) || c.findErrors(int(instruction.rd))
					registerPipelineAdd = append(registerPipelineAdd, []int{int(instruction.rn), 0})
					registerPipelineAdd = append(registerPipelineAdd, []int{int(instruction.rd), 0})
					break
				case "I":
					errorsExist = c.findErrors(int(instruction.rn))
					errorsExist = c.findErrors(int(instruction.rn)) || errorsExist
					break
				}
			}
		}
	}

	//c.runInstruction(value[0].(Instruction))*/
}
