package main

func (c *Control) aluProcess(bufPreALU *Queue) {
	var aluData, errALU = bufPreALU.dequeue()

	if errALU != -1 {
		c.runInstruction(aluData[0].(Instruction))
	}
}
