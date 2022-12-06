package main

func (c *Control) fetch(bufPreIssue *Queue) {
	// can fetch and decode 2 instructions per clock cycle
	// check if fetch is stalled bc  branch/cache miss
	//		fetch will be stalled if branch or cache did not find data
	// check if instruction in cache DONE
	// if instruction not in cache then grab from memory and it will be the instruction in cache
	// for the next cycle DONE
	// instructions that are branched over need to be revisited if needed
	//if read a branch, stop reading
	// writing takes the first part then reading the second
	//write to data then read data
	// when BREAK is fetched no more instructions will be fetched
	//same as branch
	// Branch, BREAK, and NOP instructions will all be fetched,
	// but will not be written into the Pre-Issue Buffer
	//special case for those three

	/*if typeofInstruction == "B" || cacheHit == true {
		fetch is stalled
	} */

	var branchOperation = false

	//fetch is not paused and pre-issue buffer is not full
	if !c.fetchPaused && cap(bufPreIssue.data) < bufPreIssue.maxSize {
		//check which cache block holds value
		var cacheBlock, alignment = c.checkCache(uint(c.programCnt))
		var instructionArray []interface{}

		//cache found value
		if alignment != -1 {
			//check which word holds value
			if alignment == 0 { //in word 1
				instructionArray = append(instructionArray, cacheBlock.word1)
			} else { //in word 2
				instructionArray = append(instructionArray, cacheBlock.word2)
			}

			var instruction = instructionArray[0].(Instruction)

			//check if break, branch, or NOP
			switch instruction.op {
			case "B": //branch by immediate
				c.programCnt += int(instruction.offset * 4)
				branchOperation = true
				break
			case "BREAK": //do nothing
				break
			case "NOP":
				break
			default:
				bufPreIssue.enqueue(instructionArray)
				break
			}

			if !branchOperation {
				c.programCnt += 4
			}
		}
	}
}
