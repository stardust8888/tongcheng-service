package quenex

import (
	"math/bits"
	"runtime"
	"sync/atomic"
)

const cacheLine = 64

// Ring 是一个无锁、单生产者-单消费者、固定容量的泛型环形队列
type Ring[T any] struct {
	_     [cacheLine]byte // 伪共享填充
	data  []T             // 存储槽
	mask  uint64          // 容量掩码 64位=8b
	_     [cacheLine]byte //64b 占一个缓存行
	write uint64          // 仅生产者写
	_     [cacheLine]byte
	read  uint64 // 仅消费者写
	_     [cacheLine]byte
}

// New 创建队列，cap 必须 2 的幂，最小 2
func NewSPSC[T any](cap int) *Ring[T] {
	if cap < 2 {
		cap = 2
	}
	// 向上取 2 的幂
	cap = 1 << (64 - leadingZeros(uint64(cap-1)))
	r := &Ring[T]{
		data: make([]T, cap),
		mask: uint64(cap - 1),
	}
	return r
}

// --------------------------------------------------------
// 生产者 API
// --------------------------------------------------------

// Write 阻塞写：队列满时自旋等待（可嵌 runtime.Gosched 降低 CPU）
func (r *Ring[T]) Write(v T) {
	next := (r.write + 1) & r.mask
	/*
			是**环形数组（RingBuffer）**里“把下标绕回开头”的核心技巧，等价于：
			next := (r.write + 1) % len(r.data)
		1. 为什么能代替取模？
		前提：容量必须是 2 的幂
		设 capacity = 2^n，则：
		mask = capacity - 1   // 低 n 位全 1
		任何整数 x 对 2^n 取模，结果就是 x 的低 n 位
		而 x & mask 正是保留低 n 位的位运算，于是：
		x % 2^n  ≡  x & (2^n - 1)
	*/
	for next == atomic.LoadUint64(&r.read) { // 空槽等待
		runtime.Gosched()
	}
	r.data[r.write] = v
	atomic.StoreUint64(&r.write, next) // 发布
}

// TryWrite 非阻塞写：成功返回 true，满返回 false
func (r *Ring[T]) TryWrite(v T) bool {
	next := (r.write + 1) & r.mask
	if next == atomic.LoadUint64(&r.read) {
		return false // 满
	}
	r.data[r.write] = v
	atomic.StoreUint64(&r.write, next)
	return true
}

// --------------------------------------------------------
// 消费者 API
// --------------------------------------------------------

// Read 阻塞读：空时自旋
func (r *Ring[T]) Read() T {
	for atomic.LoadUint64(&r.write) == r.read {
		runtime.Gosched()
	}
	v := r.data[r.read]
	r.read = (r.read + 1) & r.mask
	return v
}

// TryRead 非阻塞读：ok==false 表示空
func (r *Ring[T]) TryRead() (v T, ok bool) {
	if atomic.LoadUint64(&r.write) == r.read {
		return v, false
	}
	v = r.data[r.read]
	r.read = (r.read + 1) & r.mask
	return v, true
}

// --------------------------------------------------------
// 辅助
// --------------------------------------------------------

// Capacity 返回实际容量（必为 2 的幂）
func (r *Ring[T]) Capacity() int { return len(r.data) }

// Used 快照当前元素数（非精确，仅监控）
func (r *Ring[T]) Used() int {
	w := int(atomic.LoadUint64(&r.write))
	rd := int(atomic.LoadUint64(&r.read))
	return (w - rd) & int(r.mask)
}

// 快速 leadingZeros（Go 1.20+ 内置，兼容旧版）
func leadingZeros(x uint64) int { return 64 - bits.Len64(x) }
