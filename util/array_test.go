package util

import "testing"

func TestArray(t *testing.T) {
	realArray := make([]int32, 20)
	for i := 0; i < len(realArray); i++ {
		realArray[i] = int32(i)
	}

	array := NewBlockedArray(realArray, 10)
	t.Logf("%v", array.GetAll())

	for i := 0; i < 10; i++ {
		array.Update(i, array.Get(i) + 1)
	}
	for i := 10; i < 20; i++ {
		array.Update(i, int32(i - 1))
	}
	t.Logf("%v", array.GetAll())

	array.Insert(0, 0)
	t.Logf("%v", array.GetAll())
	array.Insert(5, -5)
	t.Logf("%v", array.GetAll())
	array.Insert(15, -15)
	t.Logf("%v", array.GetAll())

	array.Delete(5)
	array.Delete(15)
	t.Logf("%v", array.GetAll())
}
