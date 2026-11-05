package bitwise

/*
乘除取模  ←  左/右移 + 掩码
置零置一  ←  与/或
翻转判断  ←  异或
最值绝对  ←  算术右移掩码
压缩对齐  ←  位与
哈希分片  ←  掩码
打包校验  ←  位移组合

位运算符号（以 C/Go 语法为例）：
&  位与（AND）
|  位或（OR）
^  位异或（XOR）
<< 左移（LEFT SHIFT）
>> 右移（RIGHT SHIFT）
~  位取反（NOT，一元运算符）
共 6 种基本符号。
*/

// SetBit 将指定位置的位设置为1
// n: 原始数字
// pos: 要设置的位位置 (0-indexed)
// 返回设置指定位后的数字
func SetBit(n int, pos uint) int {
	return n | (1 << pos)
}

// ClearBit 将指定位置的位设置为0
// n: 原始数字
// pos: 要清除的位位置 (0-indexed)
// 返回清除指定位后的数字
func ClearBit(n int, pos uint) int {
	return n &^ (1 << pos)
}

// ToggleBit 切换指定位置的位（0变1，1变0）
// n: 原始数字
// pos: 要切换的位位置 (0-indexed)
// 返回切换指定位后的数字
func ToggleBit(n int, pos uint) int {
	return n ^ (1 << pos)
}

// CheckBit 检查指定位置的位是否为1
// n: 原始数字
// pos: 要检查的位位置 (0-indexed)
// 返回true如果指定位为1，否则返回false
func CheckBit(n int, pos uint) bool {
	return (n & (1 << pos)) != 0
}

// GetBit 获取指定位置的位值
// n: 原始数字
// pos: 要获取的位位置 (0-indexed)
// 返回指定位的值（0或1）
func GetBit(n int, pos uint) int {
	return (n >> pos) & 1
}

// SetBits 将多个位置的位设置为1
// n: 原始数字
// positions: 要设置的位位置列表
// 返回设置指定位后的数字
func SetBits(n int, positions ...uint) int {
	for _, pos := range positions {
		n = SetBit(n, pos)
	}
	return n
}

// ClearBits 将多个位置的位设置为0
// n: 原始数字
// positions: 要清除的位位置列表
// 返回清除指定位后的数字
func ClearBits(n int, positions ...uint) int {
	for _, pos := range positions {
		n = ClearBit(n, pos)
	}
	return n
}

// CountBits 计算数字中为1的位的数量（汉明重量）
// n: 输入数字
// 返回为1的位的数量
func CountBits(n int) int {
	count := 0
	for n != 0 {
		count += n & 1
		n >>= 1
	}
	return count
}

// CountBitsFast 使用Brian Kernighan算法快速计算为1的位数
// n: 输入数字
// 返回为1的位的数量
func CountBitsFast(n int) int {
	count := 0
	for n != 0 {
		n &= n - 1 // 清除最右边的1
		count++
	}
	return count
}

// IsPowerOfTwo 检查一个数是否是2的幂
// n: 输入数字
// 返回true如果是2的幂，否则返回false
func IsPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

// ReverseBits 反转一个数字的位（假设为32位整数）
// n: 输入数字
// 返回位反转后的数字
func ReverseBits(n uint32) uint32 {
	var result uint32
	for i := 0; i < 32; i++ {
		result <<= 1
		result |= n & 1
		n >>= 1
	}
	return result
}

// LeftRotate 循环左移指定位数
// n: 原始数字
// bits: 移动的位数
// width: 数字的位宽（通常为32或64）
// 返回循环左移后的数字
func LeftRotate(n uint32, bits, width int) uint32 {
	bits %= width
	return (n << bits) | (n >> (width - bits))
}

// RightRotate 循环右移指定位数
// n: 原始数字
// bits: 移动的位数
// width: 数字的位宽（通常为32或64）
// 返回循环右移后的数字
func RightRotate(n uint32, bits, width int) uint32 {
	bits %= width
	return (n >> bits) | (n << (width - bits))
}

// GetLowestSetBit 获取最低位的1的位置
// n: 输入数字
// 返回最低位1的位置（0-indexed），如果n为0则返回-1
func GetLowestSetBit(n int) int {
	if n == 0 {
		return -1
	}
	pos := 0
	for (n & 1) == 0 {
		n >>= 1
		pos++
	}
	return pos
}

// IsEven 检查一个数是否是偶数（使用位运算）
// n: 输入数字
// 返回true如果是偶数，否则返回false
func IsEven(n int) bool {
	return (n & 1) == 0
}

// IsOdd 检查一个数是否是奇数（使用位运算）
// n: 输入数字
// 返回true如果是奇数，否则返回false
func IsOdd(n int) bool {
	return (n & 1) != 0
}

// MultiplyByPowerOfTwo 乘以2的幂（使用位运算）
// n: 原始数字
// power: 2的幂指数
// 返回 n * 2^power
func MultiplyByPowerOfTwo(n, power int) int {
	return n << power
}

// DivideByPowerOfTwo 除以2的幂（使用位运算）
// n: 原始数字
// power: 2的幂指数
// 返回 n / 2^power
func DivideByPowerOfTwo(n, power int) int {
	return n >> power
}
