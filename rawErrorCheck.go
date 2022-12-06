package main

/*
Issue Unit:
It can issue up to two instructions, out of order, per clock cycle.  When an instruction is issued, it moves out of the pre-issue buffer and into either the pre-mem buffer or the pre-ALU buffer.  The issue unit searches from entry 0 to entry 3 (IN THAT ORDER) of the pre-issue buffer and issues instructions if:
No structural hazards exist (there is room in the pre-mem/pre-ALU destination buffer)
No RAW hazards exist with active instructions (all operands are ready)
Store instructions must be issued in order

make a table in control that stores all registers being modified,
*/

// findErrors finds if a register that is currently being modified is
func (c *Control) findErrors(registerIndex int) bool {
	if contains(c.registerLocTable, registerIndex) {
		return true
	} else {
		return false
	}
}

// findErrors finds if a register that is currently being modified is
func (c *Control) removeError(registerIndex int) {
	var i = 0
	for _, v := range c.registerLocTable {
		if v[0] == registerIndex {
			v = c.registerLocTable[0]
			c.registerLocTable[i] = v
			c.registerLocTable = c.registerLocTable[:len(c.registerLocTable)-1]
		}
		i += 1
	}
}

func contains(slice [][]int, value int) bool {
	for _, v := range slice {
		if v[0] == value {
			return true
		}
	}

	return false
}
