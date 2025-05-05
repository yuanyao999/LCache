package store

import (
	"container/list"
	"sync"
	"time"
)

// Value 接口：缓存的值必须实现 Len() 方法，用于估算内存占用
type Value interface {
	Len() int
}

// Options 用于配置 lruCache（由外部传入）
type Options struct {
	MaxBytes        int64
	OnEvicted       func(key string, value Value)
	CleanupInterval time.Duration
}

// lruCache 是基于标准库 list 的 LRU 缓存实现
type lruCache struct {
	mu              sync.RWMutex                  // 读写锁，保护并发访问
	list            *list.List                    // 双向链表，记录访问顺序，前面是最近访问的
	items           map[string]*list.Element      // 快速通过 key 找到链表中的元素
	expires         map[string]time.Time          // 每个 key 的过期时间（TTL）
	maxBytes        int64                         // 最大允许使用的内存（字节）
	usedBytes       int64                         // 当前已使用的内存（字节）
	onEvicted       func(key string, value Value) // 淘汰时的回调函数
	cleanupInterval time.Duration                 // 定时清理过期键的间隔时间
	cleanupTicker   *time.Ticker                  // 清理协程用的 ticker 定时器
	closeCh         chan struct{}                 // 用于优雅关闭清理协程
}

// lruEntry 是 LRU 缓存中链表元素的数据部分,供 map 和 list 共同使用。
type lruEntry struct {
	key   string // 缓存项的键
	value Value  // 缓存项的值，必须实现 Value 接口
}

// newLRUCache 创建一个新的 LRU 缓存实例
func newLRUCache(opts Options) *lruCache {
	// 设置默认清理间隔
	cleanupInterval := opts.CleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = time.Minute
	}

	c := &lruCache{
		list:            list.New(),
		items:           make(map[string]*list.Element),
		expires:         make(map[string]time.Time),
		maxBytes:        opts.MaxBytes,
		onEvicted:       opts.OnEvicted,
		cleanupInterval: cleanupInterval,
		closeCh:         make(chan struct{}),
	}

	// 启动定期清理协程
	c.cleanupTicker = time.NewTicker(c.cleanupInterval)
	go c.cleanupLoop()

	return c
}

// Get 获取缓存项，如果存在且未过期则返回
func (c *lruCache) Get(key string) (Value, bool) {
	// 加读锁（RLock），并尝试从 items 哈希表中查找该 key 对应的链表节点
	c.mu.RLock()
	elem, ok := c.items[key]
	// 如果没找到，立刻释放读锁并返回 false，表示未命中。
	if !ok {
		c.mu.RUnlock()
		return nil, false
	}

	// 判断是否设置了过期时间 expires[key]，并且当前时间已经超出 expTime
	if expTime, hasExp := c.expires[key]; hasExp && time.Now().After(expTime) {
		// 如果是：当前 key 已过期，释放读锁
		c.mu.RUnlock()
		// 异步删除（防止阻塞其他读操作）
		go c.Delete(key)

		return nil, false
	}

	// 如果没有过期，key有效：提取链表节点中的值 value
	entry := elem.Value.(*lruEntry)
	value := entry.value
	c.mu.RUnlock() // 释放读锁（读操作完成）

	// 更新 LRU 位置需要写锁
	c.mu.Lock()
	// 再次检查元素是否仍然存在（可能在获取写锁期间被其他协程删除）
	if _, ok := c.items[key]; ok {
		c.list.MoveToBack(elem) // 最近访问的放在链表尾部，表示最近使用
	}
	c.mu.Unlock()

	return value, true
}

// Set 添加或更新缓存项
func (c *lruCache) Set(key string, value Value) error {
	return c.SetWithExpiration(key, value, 0)
}

// SetWithExpiration 添加或更新缓存项，并设置过期时间
func (c *lruCache) SetWithExpiration(key string, value Value, expiration time.Duration) error {
	// 如果 value 为空，相当于删除该项
	if value == nil {
		c.Delete(key)
		return nil
	}

	// 加锁，保证线程安全
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置或清除过期时间
	var expTime time.Time
	if expiration > 0 { // 记录过期时间（当前时间 + TTL）
		expTime = time.Now().Add(expiration)
		c.expires[key] = expTime
	} else { // 否则永不过期，把该 key 从 expires map 中清除掉
		delete(c.expires, key)
	}

	if elem, ok := c.items[key]; ok {
		oldEntry := elem.Value.(*lruEntry)
		c.usedBytes += int64(value.Len() - oldEntry.value.Len()) // 更新内存占用
		oldEntry.value = value
		c.list.MoveToBack(elem) // 更新访问顺序（最近使用）
		return nil
	}

	// 如果 key 不存在：插入新元素
	entry := &lruEntry{key, value}
	elem := c.list.PushBack(entry) // 插入到链表尾部
	c.items[key] = elem
	c.usedBytes += int64(len(key) + value.Len())

	// 检查是否需要淘汰旧项
	c.evict()

	return nil
}

// Delete 从缓存中删除指定键的项
func (c *lruCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
		return true
	}
	return false
}

// Clear 清空缓存
func (c *lruCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果设置了回调函数，遍历所有项调用回调
	if c.onEvicted != nil {
		for _, elem := range c.items {
			entry := elem.Value.(*lruEntry)
			c.onEvicted(entry.key, entry.value)
		}
	}

	c.list.Init()                            // 重新初始化链表，相当于清空
	c.items = make(map[string]*list.Element) // 清空 items
	c.expires = make(map[string]time.Time)   // 清空 expires
	c.usedBytes = 0                          // 重置 usedBytes，表示缓存占用归零
}

// Len 返回缓存中的项数
func (c *lruCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

// removeElement 从缓存中删除元素，调用此方法前必须持有锁
func (c *lruCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*lruEntry)
	c.list.Remove(elem)                                      // 从链表中删除这个节点
	delete(c.items, entry.key)                               // 从 items 哈希表中删除该 key
	delete(c.expires, entry.key)                             // 从 expires 中也移除 TTL 设置（如果有）
	c.usedBytes -= int64(len(entry.key) + entry.value.Len()) // 减去该条目的内存占用

	// 如果用户设置了淘汰回调函数，就调用它（用于记录、上报、日志等）
	if c.onEvicted != nil {
		c.onEvicted(entry.key, entry.value)
	}
}

// evict 清理过期和超出内存限制的缓存，调用此方法前必须持有锁
func (c *lruCache) evict() {
	// 先清理过期项
	now := time.Now()
	for key, expTime := range c.expires {
		if now.After(expTime) {
			if elem, ok := c.items[key]; ok {
				c.removeElement(elem)
			}
		}
	}

	// 再根据内存限制清理最久未使用的项
	for c.maxBytes > 0 && c.usedBytes > c.maxBytes && c.list.Len() > 0 {
		elem := c.list.Front() // 获取最久未使用的项（链表头部）
		if elem != nil {
			c.removeElement(elem)
		}
	}
}

// cleanupLoop 定期清理过期缓存的协程
func (c *lruCache) cleanupLoop() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.mu.Lock()
			c.evict()
			c.mu.Unlock()
		case <-c.closeCh:
			return
		}
	}
}

// Close 关闭缓存，停止清理协程
func (c *lruCache) Close() {
	if c.cleanupTicker != nil {
		c.cleanupTicker.Stop()
		close(c.closeCh)
	}
}

// GetWithExpiration 获取缓存项及其剩余过期时间
func (c *lruCache) GetWithExpiration(key string) (Value, time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, 0, false
	}

	// 检查是否过期
	now := time.Now()
	if expTime, hasExp := c.expires[key]; hasExp {
		if now.After(expTime) {
			// 已过期
			return nil, 0, false
		} // 与 Get() 不同的是这里没有 go c.Delete(key)，可以认为这是个纯查询接口。

		// 计算剩余过期时间
		ttl := expTime.Sub(now)
		c.list.MoveToBack(elem)
		return elem.Value.(*lruEntry).value, ttl, true
	}

	// 无过期时间
	c.list.MoveToBack(elem)
	return elem.Value.(*lruEntry).value, 0, true
}

// GetExpiration 获取键的过期时间
func (c *lruCache) GetExpiration(key string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expTime, ok := c.expires[key]
	return expTime, ok
}

// UpdateExpiration 动态更新过期时间
func (c *lruCache) UpdateExpiration(key string, expiration time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; !ok {
		return false
	}

	if expiration > 0 {
		c.expires[key] = time.Now().Add(expiration)
	} else {
		delete(c.expires, key)
	}

	return true
}

// UsedBytes 返回当前使用的字节数
func (c *lruCache) UsedBytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.usedBytes
}

// MaxBytes 返回最大允许字节数
func (c *lruCache) MaxBytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxBytes
}

// SetMaxBytes 设置最大允许字节数并触发淘汰
func (c *lruCache) SetMaxBytes(maxBytes int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxBytes = maxBytes
	if maxBytes > 0 {
		c.evict()
	}
}
