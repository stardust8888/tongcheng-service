package poolants

import ants "github.com/panjf2000/ants/v2"

// PoolFunc 线程池结构体
type PoolFunc struct {
	innerPool *ants.PoolWithFunc
}

// NewPoolWithFunc 创建一个线程池
func NewPoolWithFunc(cap int, pf func(any)) (*PoolFunc, error) {
	pool, err := ants.NewPoolWithFunc(cap, pf)
	if err != nil {
		return nil, err
	}

	return &PoolFunc{innerPool: pool}, nil
}

// Submit 提交任务
func (p *PoolFunc) Submit(arg any) error {
	return p.innerPool.Invoke(arg)
}

// Running 返回正在运行的goroutine数量
func (p *PoolFunc) Running() int {
	return p.innerPool.Running()
}

// Free 返回空闲的goroutine数量
func (p *PoolFunc) Free() int {
	return p.innerPool.Free()
}

// Cap 返回线程池容量
func (p *PoolFunc) Cap() int {
	return p.innerPool.Cap()
}

// Tune 调整线程池大小
func (p *PoolFunc) Tune(size int) {
	p.innerPool.Tune(size)
}

// Release 释放线程池
func (p *PoolFunc) Release() {
	p.innerPool.Release()
}
