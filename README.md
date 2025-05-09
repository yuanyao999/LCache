# 🧠 LCache - 分布式本地缓存系统（仿 groupcache）

LCache 是一个仿照 Google `groupcache` 实现的本地缓存系统，支持多种缓存策略、命名空间隔离、分布式节点同步、抖动抑制等特性。

---

## ✅ 当前完成模块

### 1. `store/`
- `lru.go`：实现了基本的 LRU 缓存淘汰策略。
- `lru2.go`：实现了带有两级缓存（L1 + L2）的 LRU2 缓存，支持 TTL、分桶、并发锁、手写双向链表。
- `store.go`：定义了统一的缓存接口 `Store`，支持 LRU 和 LRU2 的选择。

### 2. `cache/`
- `cache.go`：对底层 store 封装，提供：
    - 缓存初始化控制
    - 命中/未命中统计
    - TTL 过期控制
    - 缓存清理、关闭操作

支持核心操作：
```go
Add(key, value)
Get(ctx, key)
Delete(key)
Clear()
Stats()
```

### 3. `group/`
- `group.go`：实现命名空间 `Group`，用于支持分布式缓存设计。

功能包括：
- 本地缓存查找 → Peers 获取 → 回源加载
- 分布式同步：`syncToPeers`
- 支持过期时间配置、关闭、清除、统计信息等

### 4. `singleflight/`
- `singleflight.go`：实现请求合并机制，防止缓存击穿。
    - 多个并发请求同一个 key，仅发起一次实际加载。

### 5. `kamacache/`
- `byteview.go`：定义只读的 `ByteView` 缓存视图结构，避免底层数据被篡改。

---

## 📁 项目结构

```
LCache/
├── go.mod
├── main.go                 # 示例启动逻辑（可选）
├── group/                  # Group命名空间 & 分布式逻辑
│   └── group.go
├── cache/                  # 封装 store 的缓存层
│   └── cache.go
├── store/                  # 缓存核心策略层
│   ├── lru.go
│   ├── lru2.go
│   └── store.go
├── singleflight/           # 请求抖动抑制机制
│   └── singleflight.go
├── kamacache/              # 缓存只读视图定义
│   └── byteview.go
└── test/                   # 单元测试（可选）
    └── lru_test.go
```

---

## 🔜 下一步计划

- 实现 `PeerPicker` 和 `Peer` 接口：支持分布式节点注册、选择、请求
- 编写 `http` 或 `gRPC` 通信模块，用于组间远程数据同步
- 编写更多单元测试，测试 `group`、`cache`、`byteview` 等模块
- 编写 CLI 启动示例或 Web 接口演示

---

## 🧠 技术亮点

- 两级缓存（L1/L2）+ TTL + 双向链表 + 分桶并发锁
- 支持模块解耦，良好的接口设计（`Store`、`Getter`、`GroupOption`）
- 可拓展的分布式缓存方案
- 请求合并（`singleflight`）防止并发击穿
- 高可读性与测试性设计，便于工程扩展

---

## 📜 License

MIT License - 由你本人开发与学习使用
