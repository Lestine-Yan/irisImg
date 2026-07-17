#!/usr/bin/env bash
# irisImg 停止脚本。
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ -f data/irisImg.pid ]]; then
  pid="$(cat data/irisImg.pid)"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid"   # SIGTERM，触发后端优雅关闭(最多 5s)
    # 轮询等待进程真正退出,最长 8s(略大于后端 5s 优雅关闭超时,留余量),
    # 避免删 pidfile 后新实例撞上未释放的 :8080 监听或 SQLite WAL 锁。
    for _ in $(seq 1 80); do
      kill -0 "$pid" 2>/dev/null || break
      sleep 0.1
    done
    if kill -0 "$pid" 2>/dev/null; then
      echo "irisImg 未在超时内退出,发送 SIGKILL" >&2
      kill -9 "$pid" 2>/dev/null || true
      sleep 0.5
    fi
    echo "irisImg stopped, pid=$pid"
  else
    echo "pid $pid not alive, removing stale pid file"
  fi
  rm -f data/irisImg.pid
else
  echo "pid file not found, is irisImg running?"
fi
