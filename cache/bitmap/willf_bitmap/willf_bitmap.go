package willf_bitmap

import (
	"go_lib/cache/bitmap/bitmap_interface"

	wf "github.com/bits-and-blooms/bitset"
)

const defaultLength = 64

type WillfBitMap struct {
	b *wf.BitSet
}

func NewWillfBitMap() bitmap_interface.Bitmap {
	return &WillfBitMap{
		b: wf.New(defaultLength),
	}
}

func (m *WillfBitMap) Set(i uint) {
	m.b.Set(i)
}

func (m *WillfBitMap) Clear(i uint) {
	m.b.Clear(i)
}

func (m *WillfBitMap) Get(i int64) bool {
	return m.b.Test(uint(i))
}

func (m *WillfBitMap) Size() int64 {
	return int64(m.b.Len())
}

func (m *WillfBitMap) Reset() {
	m.b.ClearAll()
}

func (m *WillfBitMap) Clone() bitmap_interface.Bitmap {
	return &WillfBitMap{b: m.b.Clone()}
}

func (m *WillfBitMap) Equal(slave bitmap_interface.Bitmap) bool {
	_, ok := slave.(*WillfBitMap)
	if ok {
		return m.b.Equal(slave.(*WillfBitMap).b)
	}
	return false
}

func (m *WillfBitMap) Cardinality() int64 {
	return int64(m.b.Count())
}
