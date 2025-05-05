# LCache：LRU 缓存模块

这是一个使用 Go 实现的轻量级 LRU（Least Recently Used，最近最少使用）内存缓存模块。支持线程安全、TTL（过期时间）、内存限制与自动清理机制，适合作为分布式缓存系统的底层存储模块。

---

## ✅ 功能特性

- 基于 `sync.RWMutex` 实现的并发读写安全
- 使用 `container/list` 实现 LRU 淘汰机制
- 支持为每个键设置过期时间（TTL）
- 支持按最大内存字节数自动淘汰缓存项
- 后台协程定时清理过期数据
- 可选淘汰回调函数 `OnEvicted`
- 简洁清晰的 API 设计

---

## 📦 基本使用

### 自定义 Value 类型（需实现 Len() 方法）：

```go
type String string

func (s String) Len() int {
    return len(s)
}
```

### 示例代码：

```go
cache := newLRUCache(Options{
    MaxBytes:        1024,
    CleanupInterval: time.Minute,
})

cache.Set("key", String("value"))

if val, ok := cache.Get("key"); ok {
    fmt.Println("获取成功:", val)
}
```

---

## 🧪 单元测试

测试文件位于 `store/lru_test.go`，涵盖以下内容：

- Get / Set 基本操作
- TTL 过期测试
- LRU 淘汰机制验证
- Delete / Clear 功能
- 过期时间的续期（UpdateExpiration）

运行测试命令：

```bash
go test ./store -v
```

---

## 📁 项目结构

```
LCache/
├── go.mod
├── main.go
├── README.md
└── store/
    ├── lru.go
    └── lru_test.go
```

---

## 🚧 后续计划

- 增加 benchmark 性能测试
- 增加二级缓存支持（LRU-2）
- 添加 CLI 或 HTTP 示例服务
- 与统一 `Store` 接口集成
