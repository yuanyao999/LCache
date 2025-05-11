#!/bin/bash

# ========== 配置 ==========
MAIN_CMD="./cmd/main.go"
LOG_DIR="logs"

# ========== 初始化 ==========
mkdir -p "$LOG_DIR"

echo "🧹 清理已有 etcd 实例（如果有）..."
if pgrep -f etcd > /dev/null; then
  pkill -f etcd
  sleep 1
  echo "✅ 已停止旧 etcd 实例"
else
  echo "✅ 没有运行中的 etcd"
fi

# 清理旧数据目录
if [ -d "default.etcd" ]; then
  echo "🧹 清除旧 etcd 数据目录 default.etcd"
  rm -rf default.etcd
fi

# ========== 启动 etcd ==========
echo "🚀 启动 etcd 服务..."
etcd > "$LOG_DIR/etcd.log" 2>&1 &
sleep 2

# ========== 启动 LCache 节点 ==========
echo "🚀 启动节点 A (8001)..."
go run "$MAIN_CMD" -port=8001 -node=A > "$LOG_DIR/node_A.log" 2>&1 &

echo "🚀 启动节点 B (8002)..."
go run "$MAIN_CMD" -port=8002 -node=B > "$LOG_DIR/node_B.log" 2>&1 &

echo "🚀 启动节点 C (8003)..."
go run "$MAIN_CMD" -port=8003 -node=C > "$LOG_DIR/node_C.log" 2>&1 &

echo "✅ 所有节点已启动！日志目录：$LOG_DIR/"
