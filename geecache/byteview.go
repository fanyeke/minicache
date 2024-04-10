package geecache

// ByteView 选择 byte 类型的数据作为缓存值, 可以储存任意的数据,例如字符串、图片等
type ByteView struct {
	b []byte // b 是只读类型, 使用 ByteSlice() 方法返回一个拷贝, 防止缓存值被外部程序修改
}

// Len 返回字节切片的长度
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个切片的拷贝, 防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 返回字符串
func (v ByteView) String() string {
	return string(v.b)
}

// cloneBytes 返回一个切片的拷贝
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
