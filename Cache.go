package main

type Block struct {
	valid int
	dirty int
	tag   uint
	word1 interface{}
	word2 interface{}
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

func cacheAddressConversion(value uint) uint {
	var setIndex = value & 0x0003

	return setIndex
}

func checkAlignment(address uint) int {
	if address%8 == 0 {
		return 0
	} else {
		return 1
	}
}

func (c *Control) readDisk(address uint) interface{} {
	var valueAdr1 interface{}
	var addressToIndex = (address - uint(c.programCntStart)) / 4

	valueAdr1 = c.memory[addressToIndex]

	return valueAdr1
}

func (c *Control) cacheMiss(address uint, setIndex uint, inBlock int, alignment int, tag uint) {
	var addressTemp uint
	var cache = &c.cache
	var value1, value2 interface{}

	if alignment == 0 { // aligned good
		addressTemp = address + 4
		value1 = c.readDisk(address)
		value2 = c.readDisk(addressTemp)
	} else { //aligned not good
		addressTemp = address - 4
		value1 = c.readDisk(addressTemp)
		value2 = c.readDisk(address)
	}

	if inBlock == 1 {
		cache.LruBits[setIndex] = 0
	} else {
		cache.LruBits[setIndex] = 1
	}

	cache.CacheSets[setIndex][inBlock].word1 = value1
	cache.CacheSets[setIndex][inBlock].word2 = value2
	cache.CacheSets[setIndex][inBlock].tag = tag
	cache.CacheSets[setIndex][inBlock].valid = 1
}

func (c *Control) checkCache(address uint) (Block, int) {
	//var memoryAddress = strconv.FormatInt(int64(address), 2)
	var cache = &c.cache
	var setIndex = cacheAddressConversion(address)
	var tag = (address & 0xFFF0) >> 3
	var i = 0
	var inBlock = -1
	var alignment = checkAlignment(address)

	for i < 2 {
		if cache.CacheSets[setIndex][i].tag == tag {
			inBlock = i
			break
		}
		i += 1
	}

	//cache hit
	if inBlock != -1 {
		//if valid bit is valid
		if cache.CacheSets[setIndex][inBlock].valid == 1 {
			return cache.CacheSets[setIndex][inBlock], alignment
		} else {
			c.cacheMiss(address, setIndex, cache.LruBits[setIndex], alignment, tag)
		}
	} else { //cache miss
		c.cacheMiss(address, setIndex, cache.LruBits[setIndex], alignment, tag)
	}

	return Block{}, -1
}
