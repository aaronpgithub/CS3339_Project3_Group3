package main

func fetch() {
	// can fetch and decode 2 instructions per clock cycle
	// check if fetch is stalled bc branch/cache miss
	// check if instruction in cache
	// if instruction not in cache then grab from memory and it will be the instruction in cache
	// for the next cycle
	// instructions that are branched over need to be revisited if needed
	// writing takes the first part then reading the second
	// when BREAK is fetched no more instructions will be fetched
	// Branch, BREAK, and NOP instructions will all be fetched,
	// but will not be written into the Pre-Issue Buffer

	/*if typeofInstruction == "B" || cacheHit == false {
		fetch is stalled
	} */
	/*
		if typeofInstruction == "BREAK" {
			no more instructions
		}
	*/
}
