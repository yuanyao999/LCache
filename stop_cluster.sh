#!/bin/bash

echo "🛑 正在停止所有 LCache 节点和 etcd..."

# 停止 etcd
ETCD_PID=$(pgrep -f "etcd")
if [ -n "$ETCD_PID" ]; then
  kill "$ETCD_PID"
  echo "✅ 已停止 etcd (PID: $ETCD_PID)"
else
  echo "⚠️ 未检测到 etcd 正在运行"
fi

# 停止节点 A/B/C
for NODE in A B C; do
  PID=$(pgrep -f "main.go.*-node=$NODE")
  if [ -n "$PID" ]; then
    kill "$PID"
    echo "✅ 已停止 节点 $NODE (PID: $PID)"
  else
    echo "⚠️ 节点 $NODE 未在运行"
  fi
done

# 强杀端口绑定进程（修复多 PID 问题）
for PORT in 8001 8002 8003; do
  echo "🔍 检查端口 $PORT 是否被占用..."
  for pid in $(lsof -ti tcp:$PORT); do
    echo "⚠️ 端口 $PORT 被进程 PID=$pid 占用，尝试 kill -9"
    kill -9 "$pid"
    echo "✅ 已强制释放 PID=$pid"
  done
done

echo "🧹 所有节点和占用端口处理完毕。"

# 清空 logs 日志
echo "🧹 清空 logs/ 目录下旧日志..."
rm -f logs/*.log
echo "✅ 日志清理完毕"
