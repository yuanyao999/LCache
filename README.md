# 🚀 LCache – 高性能分布式缓存系统 (Go + gRPC + etcd)

[![MIT License](https://img.shields.io/github/license/yuanyao999/LCache)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yuanyao999/LCache)](https://goreportcard.com/report/github.com/yuanyao999/LCache)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-brightgreen)](https://golang.org/)
[![etcd](https://img.shields.io/badge/etcd-v3.5.x-blue)](https://github.com/etcd-io/etcd)

> ✨ 类似 GroupCache 的分布式缓存系统，支持多节点注册发现、跨节点拉取、LRU2 策略与缓存命中率统计。

---

## 🔧 技术架构

- 编程语言：Go 1.20+
- 通信协议：gRPC
- 服务发现：etcd
- 缓存机制：自研 LRU / LRU2
- 特色功能：一致性哈希节点路由、跨节点同步、命中率统计、日志系统

---

## ⚙️ 架构图

```
                        ┌────────────────────────┐
                        │         etcd           │
                        │ 服务注册 & 服务发现中心 │
                        └──────────┬─────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
      ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
      │  LCache A    │      │  LCache B    │      │  LCache C    │
      │ (127.0.0.1:8001) │  │ (127.0.0.1:8002) │  │ (127.0.0.1:8003) │
      └──────┬────────┘      └──────┬────────┘      └──────┬────────┘
             │                       │                       │
             │                       │                       │
             ▼                       ▼                       ▼
       ┌──────────┐           ┌──────────┐           ┌──────────┐
       │  LRU2 缓存 │           │  LRU2 缓存 │           │  LRU2 缓存 │
       └──────────┘           └──────────┘           └──────────┘
             │                       │                       │
             ▼                       ▼                       ▼
       ┌────────────┐         ┌────────────┐         ┌────────────┐
       │  Group API │ ◄────┐  │  Group API │ ◄────┐  │  Group API │ ◄────┐
       └────┬───────┘      │  └────┬───────┘      │  └────┬───────┘      │
            │              │       │              │       │              │
     ┌──────▼────────┐     │ ┌─────▼────────┐     │ ┌─────▼────────┐     │
     │ PeerPicker 一致性哈希 │────┘ │ PeerPicker 一致性哈希 │────┘ │ PeerPicker 一致性哈希 │────┘
     └──────┬────────┘       └─────┬────────┘       └─────┬────────┘
            │                      │                      │
            ▼                      ▼                      ▼
     ┌────────────┐         ┌────────────┐         ┌────────────┐
     │ gRPC Client │         │ gRPC Client │         │ gRPC Client │
     └──────┬──────┘         └──────┬──────┘         └──────┬──────┘
            │                       │                       │
            ▼                       ▼                       ▼
       (通过 gRPC 发起远程数据拉取请求到目标 LCache 节点)

```

## ✅ 架构说明要点

- 每个 LCache 节点自带一套本地缓存（LRU2），并通过 Group 管理命名空间。

- 所有节点启动时自动向 etcd 注册服务，并监听服务变更。

- 通过一致性哈希确定 key 应该由哪个节点负责。

- 本地未命中时自动通过 gRPC 请求其他节点拉取。

- 支持缓存加载器（GetterFunc）在所有节点回源。



---

## 🚀 快速开始

```bash
# 克隆项目
git clone https://github.com/yuanyao999/LCache.git && cd LCache

# 编译 protobuf
cd pb
protoc --go_out=. --go-grpc_out=. cache.proto

# 启动集群（etcd + 节点 A/B/C）
chmod +x start_cluster.sh stop_cluster.sh
./start_cluster.sh
```

---

## 🧪 功能展示

- 每个节点写入本地键（key_A / key_B / key_C）
- 自动服务发现后，每个节点尝试跨节点 `Get`
- 输出缓存命中统计 & 节点路由日志

示例日志：

```
节点A: 获取远程键 key_B 成功: 这是节点B的数据
缓存统计: map[local_hits:1 peer_hits:2 ...]
```

---

## 📁 目录结构

```
LCache/
├── cmd/main.go       # main.go 启动入口
├── consistenthash/   # 一致性哈希模块
├── logs/             # 日志输出目录
├── pb/               # Protobuf 文件
├── registry/         # etcd 注册模块
├── singleflight/     # 实现请求抖动抑制（防止缓存击穿）
├── store/            # LRU / LRU2 缓存实现
├── byteview.go       # 封装只读视图
├── cache.go          # 缓存适配层
├── group.go          # 分布式命名空间 Group
├── peers.go          # Peer 节点选择器
├── client.go         # gRPC 客户端实现，用于跨节点拉取/设置缓存数据（实现 Peer 接口）
├── server.go         # gRPC 服务端实现，提供 Get/Set/Delete 接口 + 注册到 etcd 的服务逻辑
├── start_cluster.sh
├── stop_cluster.sh
├── go.mod
└── README.md
```

---

## 🧹 停止集群

```bash
./stop_cluster.sh
```

该命令将：

- 停止所有节点进程
- 杀掉 etcd
- 清空 logs/*.log
- 释放 8001~8003 端口

---

## 📦 TODO & 可扩展方向

- [ ] RESTful API 网关（支持 HTTP 访问）
- [ ] Docker Compose 一键部署
- [ ] Prometheus 监控
- [ ] Web 控制面板
- [ ] 缓存压缩、预热、批量接口

---

## 📜 License

MIT © 2025 [@yuanyao999](https://github.com/yuanyao999)

---

## 🙌 致谢

- [GeeCache](https://github.com/geektutu/7days-golang/tree/master/gee-cache)
- [GroupCache](https://github.com/golang/groupcache)
- [etcd](https://github.com/etcd-io/etcd)
