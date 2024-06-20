package day2_single_node

type ByteView struct {
	v []byte
}

func (b ByteView) Len() int {
	return len(b.v)
}

func (b*ByteView) ByteSlice() []byte {
	return cloneBytes(b.v)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (b ByteView) String() string {
	return string(b.v)
}



