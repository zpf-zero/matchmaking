#!/usr/bin/env bash
#
# Ubuntu 一键部署。在解压出来的包目录里执行：
#   bash deploy.sh
# （zip 不保留执行位，所以用 bash 显式调用；脚本会给其余文件补上 +x）
#
# 做的事：检查 nginx -> 停旧后端 -> 铺文件 -> 修权限 -> 启动后端 -> 自检。
# nginx 配置已正确安装时不会重复复制或 reload；uploads 只合并，不删除已有图片。
set -euo pipefail

TARGET_ROOT="${TARGET_ROOT:-/data/matchmaking}"
SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
SITE_NAME="matchmaking.conf"
SITE_AVAILABLE="/etc/nginx/sites-available/$SITE_NAME"
SITE_ENABLED="/etc/nginx/sites-enabled/$SITE_NAME"

if [ "$(id -u)" -ne 0 ]; then
  echo "需要 root 权限，请切换到 root 后执行：bash deploy.sh"
  exit 1
fi

say() { echo; echo "==> $*"; }

nginx_is_ready() {
  command -v nginx >/dev/null 2>&1 || return 1
  [ -f "$SRC_DIR/nginx/$SITE_NAME" ] || return 1
  [ -e "$SITE_ENABLED" ] || return 1
  [ ! -e /etc/nginx/sites-enabled/default ] || return 1
  cmp -s "$SRC_DIR/nginx/$SITE_NAME" "$SITE_ENABLED" || return 1
  nginx -t >/dev/null 2>&1 || return 1
}

say "检查 nginx 配置"
NGINX_READY=0
if nginx_is_ready; then
  NGINX_READY=1
  echo "nginx 站点已正确配置，本次跳过复制与配置"
else
  echo "nginx 尚未配置或配置已变化，本次将自动安装/更新"
fi

say "检查依赖"
MISSING=()
for cmd in nginx curl rsync; do
  command -v "$cmd" >/dev/null 2>&1 || MISSING+=("$cmd")
done
if [ "${#MISSING[@]}" -gt 0 ]; then
  echo "缺少 ${MISSING[*]}，正在安装"
  apt-get update -qq
  apt-get install -y nginx curl rsync
fi

# nginx 刚安装后也可能已经存在可用配置，再判断一次。
if [ "$NGINX_READY" = "0" ] && nginx_is_ready; then
  NGINX_READY=1
fi

say "创建目录 $TARGET_ROOT"
mkdir -p "$TARGET_ROOT/web" \
         "$TARGET_ROOT/server/logs" \
         "$TARGET_ROOT/server/uploads" \
         "$TARGET_ROOT/server/configs" \
         "$TARGET_ROOT/sql"
if [ "$NGINX_READY" = "0" ]; then
  mkdir -p "$TARGET_ROOT/nginx"
fi

# 必须先停旧进程再覆盖二进制，否则会得到 "Text file busy"。
# 先用旧包自带的 stop.sh，再兜底清理所有 hunjie-server 进程。
say "停止旧的 hunjie-server 进程"
if [ -f "$TARGET_ROOT/server/stop.sh" ]; then
  bash "$TARGET_ROOT/server/stop.sh" || true
fi
OLD_PIDS="$({ pgrep -x hunjie-server || true; pgrep -f "$TARGET_ROOT/server/hunjie-server" || true; } | sort -u)"
if [ -n "$OLD_PIDS" ]; then
  echo "发现残留进程：$OLD_PIDS"
  kill $OLD_PIDS 2>/dev/null || true
  for _ in $(seq 1 10); do
    ALIVE=""
    for pid in $OLD_PIDS; do
      kill -0 "$pid" 2>/dev/null && ALIVE="$ALIVE $pid"
    done
    [ -z "$ALIVE" ] && break
    sleep 1
  done
  for pid in $OLD_PIDS; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "进程 $pid 未响应 TERM，强制结束"
      kill -9 "$pid" 2>/dev/null || true
    fi
  done
fi
rm -f "$TARGET_ROOT/server/server.pid"

say "同步前端静态文件 -> $TARGET_ROOT/web"
# 前端整体替换，--delete 清掉上个版本残留的 assets
rsync -a --delete "$SRC_DIR/web/" "$TARGET_ROOT/web/"

say "同步后端 -> $TARGET_ROOT/server"
# uploads 单独处理：只增不删，服务器上用户上传的图片不能被本地包覆盖掉
rsync -a --exclude 'uploads/' --exclude 'logs/' --exclude 'server.pid' \
      "$SRC_DIR/server/" "$TARGET_ROOT/server/"
if [ -d "$SRC_DIR/server/uploads" ]; then
  rsync -a "$SRC_DIR/server/uploads/" "$TARGET_ROOT/server/uploads/"
fi

say "同步部署附属文件"
if [ "$NGINX_READY" = "0" ]; then
  cp -f "$SRC_DIR/nginx/$SITE_NAME" "$TARGET_ROOT/nginx/$SITE_NAME"
fi
[ -d "$SRC_DIR/sql" ] && rsync -a "$SRC_DIR/sql/" "$TARGET_ROOT/sql/" || true

say "设置 root 所有权与最小必要权限"
chown -R root:root "$TARGET_ROOT"
# nginx 以 www-data 身份读 web 和 uploads，父目录必须可进入。
chmod 755 /data "$TARGET_ROOT" "$TARGET_ROOT/web" "$TARGET_ROOT/server" "$TARGET_ROOT/server/uploads"
find "$TARGET_ROOT/web" -type d -exec chmod 755 {} +
find "$TARGET_ROOT/web" -type f -exec chmod 644 {} +
find "$TARGET_ROOT/server" -type d -exec chmod 755 {} +
find "$TARGET_ROOT/server" -type f -exec chmod 644 {} +
chmod 755 "$TARGET_ROOT/server/hunjie-server" "$TARGET_ROOT/server"/*.sh
[ -f "$TARGET_ROOT/server/config.prod.yaml" ] && chmod 600 "$TARGET_ROOT/server/config.prod.yaml"
find "$TARGET_ROOT/server/uploads" -type d -exec chmod 755 {} +
find "$TARGET_ROOT/server/uploads" -type f -exec chmod 644 {} +

if [ "$NGINX_READY" = "1" ]; then
  say "沿用现有 nginx 配置"
  systemctl enable --now nginx >/dev/null 2>&1 || true
else
  say "安装或更新 nginx 站点"
  cp -f "$TARGET_ROOT/nginx/$SITE_NAME" "$SITE_AVAILABLE"
  ln -sfn "$SITE_AVAILABLE" "$SITE_ENABLED"
  rm -f /etc/nginx/sites-enabled/default
  nginx -t
  systemctl enable --now nginx >/dev/null 2>&1 || true
  systemctl reload nginx
fi

say "以 root 启动新后端"
bash "$TARGET_ROOT/server/start.sh"

say "自检"
FAILED=0
if curl -fsS http://127.0.0.1:8080/health >/dev/null; then
  echo "  [OK] 后端 127.0.0.1:8080/health"
else
  echo "  [FAIL] 后端 /health 不通，看 $TARGET_ROOT/server/logs/server.log"
  FAILED=1
fi
if curl -fsS http://127.0.0.1/health >/dev/null; then
  echo "  [OK] nginx 转发 /health"
else
  echo "  [FAIL] nginx 转发 /health 不通"
  FAILED=1
fi
if curl -fsS -o /dev/null http://127.0.0.1/; then
  echo "  [OK] 前端首页"
else
  echo "  [FAIL] 前端首页打不开"
  FAILED=1
fi
SAMPLE="$(find "$TARGET_ROOT/server/uploads" -type f -name '*.png' -o -type f -name '*.jpg' 2>/dev/null | head -n1 || true)"
if [ -n "$SAMPLE" ]; then
  REL="${SAMPLE#$TARGET_ROOT/server/}"
  if curl -fsS -o /dev/null "http://127.0.0.1/$REL"; then
    echo "  [OK] 头像/图片 /$REL"
  else
    echo "  [FAIL] 图片读不到：/$REL（多半是权限问题）"
    FAILED=1
  fi
else
  echo "  [--] uploads 里还没有图片可测"
fi

echo
if [ "$FAILED" = "0" ]; then
  echo "部署完成。"
else
  echo "部署完成，但有检查项没通过，见上面 [FAIL]。"
fi
echo "  前端：      http://$(hostname -I | awk '{print $1}')/"
echo "  后端日志：  tail -f $TARGET_ROOT/server/logs/server.log"
echo "  运行状态：  $TARGET_ROOT/server/status.sh"
echo
echo "提醒：外部 MySQL 的白名单要放行这台机器的公网 IP，否则后端连不上库。"
