package quenex

import (
	"sync/atomic"
	"unsafe"
)

const retryMax = 8

type node[T any] struct {
	data T
	next unsafe.Pointer // *node[T]
}

// Queue 是无锁多生产者-单消费者链表
type Queue[T any] struct {
	head unsafe.Pointer // *node[T] ; 生产者只动 head
	tail unsafe.Pointer // *node[T] ; 消费者只动 tail
}

func NewMPSC[T any]() *Queue[T] {
	stub := &node[T]{} // 哨兵节点，next=nil
	stub.next = nil
	return &Queue[T]{
		head: unsafe.Pointer(stub),
		tail: unsafe.Pointer(stub),
	}
}

// Push（多生产者，CAS 循环）
func (q *Queue[T]) Push(v T) {
	n := &node[T]{data: v}
	// 1. 原子交换：把新节点设为 head，返回旧 head
	prev := (*node[T])(atomic.SwapPointer(&q.head, unsafe.Pointer(n)))
	// 2. 旧 head.next 指向新节点（此时链表暂时断开，读端会重试）
	prev.next = unsafe.Pointer(n)
}

// Pop（单消费者，无 CAS）
func (q *Queue[T]) Pop() (T, bool) {
	tail := (*node[T])(q.tail)
	next := (*node[T])(tail.next)
	if next != nil {
		q.tail = unsafe.Pointer(next)
		return next.data, true
	}
	for i := 0; i < retryMax; i++ {
		next = (*node[T])(tail.next)
		if next != nil {
			q.tail = unsafe.Pointer(next)
			return next.data, true
		}
	}
	var zero T
	return zero, false
}
