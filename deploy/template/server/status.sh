#!/usr/bin/env bash
# 一眼看清进程 / 端口 / 健康检查 / 日志尾巴
set -uo pipefail

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
CONF="$BASE_DIR/config.prod.yaml"
PID_FILE="$BASE_DIR/server.pid"
LOG_FILE="$BASE_DIR/logs/server.log"

PORT="$(grep -E '^\s+port:' "$CONF" 2>/dev/null | head -n1 | tr -dc '0-9')"
PORT="${PORT:-8066}"

echo "== 进程 =="
if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo "运行中，PID=$(cat "$PID_FILE")"
else
  PIDS="$(pgrep -f "$BASE_DIR/hunjie-server" || true)"
  if [ -n "$PIDS" ]; then
    echo "运行中（PID 文件缺失），PID=$PIDS"
  else
    echo "未运行"
  fi
fi

echo
echo "== 端口 =="
ss -lntp 2>/dev/null | grep -E ":(80|${PORT})\b" || echo "没有监听 :80 / :${PORT}"

echo
echo "== 健康检查 =="
curl -fsS "http://127.0.0.1:${PORT}/health" && echo || echo "本地 /health 不通"

echo
echo "== 日志最后 20 行 =="
tail -n 20 "$LOG_FILE" 2>/dev/null || echo "还没有日志"
