package poolants

import (
	"github.com/panjf2000/ants/v2"
)

// Pool 线程池结构体
type Pool struct {
	innerPool *ants.Pool
}

// NewPool 创建一个线程池
func NewPool(cap int) (*Pool, error) {
	pool, err := ants.NewPool(cap)
	if err != nil {
		return nil, err
	}

	return &Pool{innerPool: pool}, nil
}

// Submit 提交任务
func (p *Pool) Submit(task func()) error {
	return p.innerPool.Submit(task)
}

// Running 返回正在运行的goroutine数量
func (p *Pool) Running() int {
	return p.innerPool.Running()
}

// Free 返回空闲的goroutine数量
func (p *Pool) Free() int {
	return p.innerPool.Free()
}

// Cap 返回线程池容量
func (p *Pool) Cap() int {
	return p.innerPool.Cap()
}

// Tune 调整线程池大小
func (p *Pool) Tune(size int) {
	p.innerPool.Tune(size)
}

// Release 释放线程池
func (p *Pool) Release() {
	p.innerPool.Release()
}
