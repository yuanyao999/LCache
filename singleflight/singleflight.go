package singleflight

import (
	"sync"
)

// 代表一个正在进行或已结束的请求
type call struct {
	wg  sync.WaitGroup // 用于等待结果的完成
	val interface{}    // 请求的返回值
	err error          // 请求过程中的错误
}

// Group manages all kinds of calls
type Group struct {
	m sync.Map // 使用sync.Map来优化并发性能：key:string → value:*call
}

// Do 针对同一个 key，无论多少个 goroutine 调用 Do()，都只会执行一次 fn()，并将执行结果复用给所有调用者
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 检查是否已有请求在处理该 key
	if existing, ok := g.m.Load(key); ok {
		c := existing.(*call)
		c.wg.Wait()         // 等待已有的 call 完成
		return c.val, c.err // 复用结果
	}

	// If no ongoing request, create a new one
	c := &call{}
	c.wg.Add(1)
	g.m.Store(key, c) // Store the call in the map

	// 执行真正的加载逻辑
	c.val, c.err = fn()
	c.wg.Done() // Mark the request as done

	// 当前请求处理完毕, 清理 key，避免内存泄漏
	g.m.Delete(key)

	return c.val, c.err
}
