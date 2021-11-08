package bitmap_interface

// 接口定义
type Bitmap interface {
	Set(uint)       // 将位置1的元素值设置为1
	Clear(uint)     // 将位置1的元素值设置为0
	Get(int64) bool // 查询位置1的元素
	Size() int64    // 查询bitmap的长度
	Reset()         // 清空bitmap元素
	Clone() Bitmap  // 拷贝该bitmap
	//Copy(Bitmap) Bitmap // 将该bitmap拷贝到指定的bitmap中，但不会对原有bitmap进行扩容
	Equal(Bitmap) bool  // 比较和另一个Bitmap是否相等
	Cardinality() int64 // 已设置值为1的元素个数

	// bitmap的位运算
	//And(...Bitmap) Bitmap    // 与模式
	//Or(...Bitmap) Bitmap     // 或模式
	//Xor(...Bitmap) Bitmap    // 异或
	//AndNot(...Bitmap) Bitmap // 与非
	//Not() Bitmap             // 非
}
