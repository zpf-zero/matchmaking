#!/usr/bin/env bash
# 停止后端。先 TERM 等它自己退，超时才 KILL
set -euo pipefail

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN="$BASE_DIR/hunjie-server"
PID_FILE="$BASE_DIR/server.pid"

stop_pid() {
  local pid="$1"
  kill -0 "$pid" 2>/dev/null || return 1
  kill "$pid" 2>/dev/null || true
  for _ in $(seq 1 10); do
    kill -0 "$pid" 2>/dev/null || { echo "已停止，PID=$pid"; return 0; }
    sleep 1
  done
  echo "PID=$pid 未响应 TERM，强制结束"
  kill -9 "$pid" 2>/dev/null || true
  return 0
}

STOPPED=0
if [ -f "$PID_FILE" ]; then
  stop_pid "$(cat "$PID_FILE")" && STOPPED=1
  rm -f "$PID_FILE"
fi

# PID 文件丢了也要能收拾干净，否则下次 start.sh 会被残留进程挡住
for pid in $(pgrep -f "$BIN" || true); do
  stop_pid "$pid" && STOPPED=1
done

[ "$STOPPED" = "1" ] || echo "服务未在运行"
