# LCache 项目进度说明

## ✅ 当前已完成模块

### 1. store/ 层
- `lru.go`：实现了基础的 LRU 缓存淘汰策略。
- `lru2.go`：实现了双层缓存（L1 + L2）结构的 LRU2 算法，支持：
  - 分桶（Bucket）
  - 并发锁（每个桶一个 mutex）
  - TTL 机制
  - 手写双向链表
- `store.go`：定义统一缓存接口 `Store`，用于抽象多种缓存实现（如 LRU、LRU2）。

---

### 2. cache/ 层
- `cache.go`：对底层 store 层进行封装，支持：
  - 缓存初始化控制（懒加载）
  - 读写并发安全
  - 过期时间（TTL）
  - 命中/未命中统计
  - 缓存关闭与资源释放
  - 缓存接口如：`Add`, `AddWithExpiration`, `Get`, `Delete`, `Clear`, `Stats`

---

### 3. group/ 层
- `group.go`：实现缓存命名空间 `Group`，支持分布式缓存架构。
  - 支持三层数据加载策略：本地缓存 → 分布式节点 Peer → Getter 回源
  - 支持缓存自动同步（`syncToPeers`）
  - 支持缓存组关闭、清除、统计等接口
  - 通过 `RegisterPeers` 支持一致性哈希注册

---

### 4. singleflight/ 层
- `singleflight.go`：实现请求抖动抑制（防止缓存击穿）
  - 针对相同 key 的并发 `load` 请求只执行一次，其他协程等待并复用结果

---

### 5. kamacache/ 层
- `byteview.go`：定义缓存值类型 `ByteView`，封装只读视图：
  - `ByteSlice()` 返回拷贝，防止引用篡改
  - `String()` 实现字符串表示
  - `Len()` 获取数据大小（用于限流）

---

### 6. 一致性哈希模块（今日新增）
- `consistenthash.go`：
  - 实现虚拟节点机制，提高节点均匀性
  - 支持添加、删除节点、根据 key 映射节点
  - 支持哈希函数定制
  - 将被集成到分布式 peer 通信层中用于选择节点

---

## 📁 推荐项目结构
```
LCache/
├── go.mod
├── main.go                   // 启动示例
├── kamacache/
│   └── byteview.go
├── cache/
│   └── cache.go
├── group/
│   └── group.go
├── store/
│   ├── lru.go
│   ├── lru2.go
│   └── store.go
├── singleflight/
│   └── singleflight.go
├── consistenthash/
│   └── consistenthash.go
├── test/
│   └── lru_test.go
```

---

## 🔜 下一步建议开发模块

- `peers.go`: 实现 `PeerPicker` 和 `Peer` 接口，负责节点间的调用（结合一致性哈希）
- `http.go`: 使用 HTTP 实现节点间通信服务
- `example/`: 添加演示 demo 和 benchmark
- 完善测试覆盖率，涵盖 LRU2/LRU、多协程、过期淘汰、group 流程

---
