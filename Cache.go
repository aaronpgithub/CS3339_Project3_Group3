package main

func CheckCache() {
	type Block struct {
		valid int
		dirty int
		tag   int
		word1 int
		word2 int
	}

	var address1, address2, dataWord, addressLocal int32

	var Set = [2]Block{
		Block{
			valid: 0,
			dirty: 0,
			tag:   0,
			word1: 0,
			word2: 0,
		},
		Block{
			valid: 1,
			dirty: 1,
			tag:   1,
			word1: 2,
			word2: 1,
		},
	}

	var CacheSets [4][2]Block
	var JustMissedList []int
	var LruBits = [4]int{0, 0, 0, 0}

	//alignment check
	if addressLocal%8 == 0 { // alligned good
		dataWord = 0 // block0 is the address
		address1 = addressLocal
		address2 = addressLocal + 4
		//CacheSets[][].word1 = addressLocal
	}
	if addressLocal%8 != 0 { //alligned not good
		dataWord = 1 // block1 is the address
		address1 = addressLocal - 4
		address2 = addressLocal
		//CacheSets[][].word2 = addressLocal
		// 
	}

}

func Flush(c Control) {
	//Writes out all dirty blocks to memory at the conclusion of execution
	if CacheSets[][].dirty == 1 {
		c.memoryData[] = CacheSets[][]. 
	}

}

func DataMemManagement() {

}