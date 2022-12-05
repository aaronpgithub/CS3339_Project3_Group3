package main

type Block struct {
	valid int
	dirty int
	tag   uint
	word1 int
	word2 int
}

type Cache struct {
	//var address1, address2, dataWord, addressLocal int

	/*
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
		}*/

	CacheSets      [4][2]Block
	JustMissedList []int
	LruBits        [4]int

	//alignment check
	/*
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
		}*/

}

func cacheAddressConversion(value uint) (uint, uint) {
	var blockIndex = value & 0x0004
	var setIndex = value & 0x0003

	return blockIndex, setIndex
}

func (c Cache) checkCache(address uint) Block {
	//var memoryAddress = strconv.FormatInt(int64(address), 2)
	var blockIndex, setIndex = cacheAddressConversion(address)
	var tag = (address & 0xFFF0) >> 3
	var addressTemp uint

	//cache hit
	if c.CacheSets[setIndex][blockIndex].tag == tag {
		return c.CacheSets[setIndex][blockIndex]
	} else { //cache miss
		//if aligned
		if address%8 == 0 { // aligned good
			addressTemp = address + 4
			c.CacheSets[setIndex][blockIndex].word1 = int(addressTemp)
		}
		if address%8 != 0 { //aligned not good
			addressTemp = address - 4
			//CacheSets[][].word2 = addressLocal
		}
	}

}
