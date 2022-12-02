package main

func (c *Control) issueProcess(bufPreIssue *Queue) {

}

/*
Issue Unit:
It can issue up to two instructions, out of order, per clock cycle.  When an instruction is issued, it moves out of the pre-issue buffer and into either the pre-mem buffer or the pre-ALU buffer.  The issue unit searches from entry 0 to entry 3 (IN THAT ORDER) of the pre-issue buffer and issues instructions if:
No structural hazards exist (there is room in the pre-mem/pre-ALU destination buffer)
No RAW hazards exist with active instructions (all operands are ready)
Store instructions must be issued in order
*/
