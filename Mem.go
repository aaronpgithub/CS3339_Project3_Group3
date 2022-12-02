package main

//if stur, do nothing
//if ldur, send to wb

// memProcess takes the values from the pre memory buffer and calculates what to do based off of instruction.
func (c *Control) memProcess(bufPreMem *Queue) {
	var value, err = bufPreMem.dequeue()

	if err != -1 {
		c.runInstruction(value[0].(Instruction))
	}
}
