#!/usr/bin/env bash
# 后台启动后端。日志追加到 logs/server.log，PID 写入 server.pid
set -euo pipefail

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$BASE_DIR"

BIN="$BASE_DIR/hunjie-server"
CONF="$BASE_DIR/config.prod.yaml"
PID_FILE="$BASE_DIR/server.pid"
LOG_FILE="$BASE_DIR/logs/server.log"

[ -x "$BIN" ] || { echo "找不到可执行文件或没有执行权限：$BIN"; exit 1; }
[ -f "$CONF" ] || { echo "找不到配置文件：$CONF"; exit 1; }

mkdir -p "$BASE_DIR/logs" "$BASE_DIR/uploads" "$BASE_DIR/configs"

# 已经在跑就不重复拉起。除了 PID 文件，也查一遍进程表，
# 防止 PID 文件被误删之后启出第二个实例来抢 8066 端口
if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo "服务已在运行，PID=$(cat "$PID_FILE")"
  exit 0
fi
RUNNING="$(pgrep -f "$BIN" || true)"
if [ -n "$RUNNING" ]; then
  echo "检测到未被 PID 文件记录的进程：$RUNNING"
  echo "先执行 ./stop.sh 再启动"
  exit 1
fi
rm -f "$PID_FILE"

# 配置里 sensitive.Init 用的是相对路径 configs/sensitive_words.txt，
# 所以必须在 BASE_DIR 下启动，不能从别的目录调用
nohup "$BIN" -c "$CONF" >> "$LOG_FILE" 2>&1 &
echo $! > "$PID_FILE"
PID="$(cat "$PID_FILE")"

# 起不来的话大多是数据库连不上，等一下再判断，直接把日志尾巴打出来省得再去翻
sleep 2
if ! kill -0 "$PID" 2>/dev/null; then
  echo "启动失败，日志最后 30 行："
  tail -n 30 "$LOG_FILE" || true
  rm -f "$PID_FILE"
  exit 1
fi

PORT="$(grep -E '^\s+port:' "$CONF" | head -n1 | tr -dc '0-9')"
PORT="${PORT:-8066}"
for _ in $(seq 1 10); do
  if curl -fsS "http://127.0.0.1:${PORT}/health" >/dev/null 2>&1; then
    echo "服务已启动，PID=$PID，健康检查通过（:${PORT}）"
    echo "日志：$LOG_FILE"
    exit 0
  fi
  sleep 1
done

echo "进程在跑（PID=$PID），但 /health 暂时没响应，日志最后 30 行："
tail -n 30 "$LOG_FILE" || true
