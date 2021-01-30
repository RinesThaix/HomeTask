package util

import (
	"fmt"
	"math"
)

type (
	block struct {
		array       []int32
		size        int
		startingPos int
	}

	BlockedArray struct {
		blocks    []*block
		size      int
		blockSize int
	}
)

func NewBlockedArray(array []int32, blocks int) *BlockedArray {
	blockedArray := &BlockedArray{}
	blockedArray.Set(array, blocks)
	return blockedArray
}

func (ba *BlockedArray) Set(array []int32, blocks int) {
	if len(array) == 0 {
		return
	}
	ba.size = len(array)
	ba.blockSize = int(math.Ceil(float64(ba.size) / float64(blocks)))
	if ba.blockSize * blocks - len(array) >= ba.blockSize {
		blocks = 1 + len(array) / ba.blockSize
		if len(array) % ba.blockSize == 0 {
			blocks--
		}
	}
	ba.blocks = make([]*block, blocks)
	blockIndex, j := -1, ba.blockSize-1
	for i, value := range array {
		j++
		if j == ba.blockSize {
			j = 0
			blockIndex++
			ba.blocks[blockIndex] = ba.newBlock(i)
		}
		ba.blocks[blockIndex].array[j] = value
		ba.blocks[blockIndex].size++
	}
	if blockIndex != blocks - 1 {
		panic(fmt.Errorf("internal error"))
	}
}

func (ba *BlockedArray) Insert(pos int, value int32) {
	if pos > ba.size {
		panic(fmt.Errorf("could not insert element at position %d: the size is only %d", pos, ba.size))
	}
	blockIndex := ba.getBlockIndex(pos)
	if blockIndex == len(ba.blocks) {
		ba.blocks[blockIndex] = ba.newBlock(pos)
	}
	block := ba.blocks[blockIndex]
	inBlockPos := block.getInBlockPosition(pos)
	if inBlockPos == len(block.array) {
		block.array = append(block.array, value)
	} else {
		for i := block.size; i > inBlockPos; i-- {
			if i == len(block.array) {
				block.array = append(block.array, block.array[i - 1])
			} else {
				block.array[i] = block.array[i - 1]
			}
		}
		block.array[inBlockPos] = value
	}
	block.size++
	ba.size++
	for i := blockIndex + 1; i < len(ba.blocks); i++ {
		ba.blocks[i].startingPos++
	}
	ba.checkAndRebalance()
}

func (ba *BlockedArray) Delete(pos int) {
	if pos >= ba.size {
		panic(fmt.Errorf("could not delete element at position %d: the size is only %d", pos, ba.size))
	}
	blockIndex := ba.getBlockIndex(pos)
	block := ba.blocks[blockIndex]
	inBlockPos := block.getInBlockPosition(pos)
	for i := inBlockPos; i < block.size-1; i++ {
		block.array[i] = block.array[i+1]
	}
	block.size--
	ba.size--
	for i := blockIndex + 1; i < len(ba.blocks); i++ {
		ba.blocks[i].startingPos--
	}
	ba.checkAndRebalance()
}

func (ba *BlockedArray) Update(pos int, value int32) {
	if pos >= ba.size {
		panic(fmt.Errorf("could not update element at position %d: the size is only %d", pos, ba.size))
	}
	block := ba.blocks[ba.getBlockIndex(pos)]
	block.array[block.getInBlockPosition(pos)] = value
}

func (ba *BlockedArray) Get(pos int) int32 {
	if pos >= ba.size {
		panic(fmt.Errorf("could not get element at position %d: the size is only %d", pos, ba.size))
	}
	block := ba.blocks[ba.getBlockIndex(pos)]
	return block.array[block.getInBlockPosition(pos)]
}

func (ba *BlockedArray) GetAll() []int32 {
	array := make([]int32, ba.size)
	i := 0
	for _, block := range ba.blocks {
		for j := 0; j < block.size; j++ {
			array[i] = block.array[j]
			i++
		}
	}
	return array
}

func (ba *BlockedArray) Size() int {
	return ba.size
}

func (ba *BlockedArray) checkAndRebalance() {
	if len(ba.blocks) == 0 {
		return
	}

	min, max := 1_000_000_000, -1
	for _, block := range ba.blocks {
		if block.size < min {
			min = block.size
		}
		if block.size > max {
			max = block.size
		}
	}
	if max > (min << 1) {
		ba.rebalance()
	}
}

func (ba *BlockedArray) rebalance() {
	ba.Set(ba.GetAll(), len(ba.blocks))
}

func (ba *BlockedArray) getBlockIndex(pos int) int {
	if len(ba.blocks) == 0 {
		panic(fmt.Errorf("there are no blocks"))
	}
	if len(ba.blocks) == 1 {
		return 0
	}
	// TODO: binary search?
	for i := 1; i < len(ba.blocks); i++ {
		from, to := ba.blocks[i - 1].startingPos, ba.blocks[i].startingPos
		if pos >= from && pos < to {
			return i - 1
		}
	}
	return len(ba.blocks) - 1
}

func (ba *BlockedArray) newBlock(startingPos int) *block {
	return &block{array: make([]int32, ba.blockSize), startingPos: startingPos}
}

func (b *block) getInBlockPosition(pos int) int {
	return pos - b.startingPos
}
