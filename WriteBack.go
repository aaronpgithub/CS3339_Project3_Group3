package main

import "fmt"

func (c *Control) writeBack(queueMEM *Queue, queueALU *Queue) {
	fmt.Print(queueMEM)
	fmt.Print(queueALU)
	var memData, errMEM = queueMEM.dequeue()
	var aluData, errALU = queueALU.dequeue()

	var memInstruction Instruction
	var aluInstruction Instruction

	//if MEM is not empty
	if errMEM != -1 {
		memInstruction = memData[0].(Instruction)
		c.registers[memInstruction.rm] = memData[1].(int64)
	}

	//if ALU is not empty
	if errALU != -1 {
		aluInstruction = aluData[0].(Instruction)
		c.registers[aluInstruction.rm] = aluData[1].(int64)
	}
}
