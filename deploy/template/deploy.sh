#!/usr/bin/env bash
#
# Ubuntu 一键部署。在解压出来的包目录里执行：
#   sudo bash deploy.sh
# （zip 不保留执行位，所以用 bash 显式调用；脚本会给其余文件补上 +x）
#
# 做的事：建目录 -> 铺文件 -> 装 nginx 站点 -> 修权限 -> 重启后端 -> 自检。
# 可重复执行；uploads 只做合并，不会删掉服务器上已有的图片。
set -euo pipefail

TARGET_ROOT="${TARGET_ROOT:-/data/matchmaking}"
SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
SITE_NAME="matchmaking.conf"

# 后端不该用 root 跑。用 sudo 调进来的话回到原来那个用户
RUN_USER="${SUDO_USER:-$(id -un)}"
RUN_GROUP="$(id -gn "$RUN_USER")"

if [ "$(id -u)" -ne 0 ]; then
  echo "需要 root 权限，请用：sudo ./deploy.sh"
  exit 1
fi

say() { echo; echo "==> $*"; }

say "检查依赖"
for cmd in nginx curl rsync; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "缺少 $cmd，正在安装…"
    apt-get update -qq
    apt-get install -y nginx curl rsync
    break
  fi
done

say "创建目录 $TARGET_ROOT"
mkdir -p "$TARGET_ROOT/web" \
         "$TARGET_ROOT/server/logs" \
         "$TARGET_ROOT/server/uploads" \
         "$TARGET_ROOT/server/configs" \
         "$TARGET_ROOT/nginx" \
         "$TARGET_ROOT/sql"

# 后端在跑就先停掉，否则覆盖二进制会得到 "Text file busy"
if [ -x "$TARGET_ROOT/server/stop.sh" ]; then
  say "停止正在运行的后端"
  sudo -u "$RUN_USER" "$TARGET_ROOT/server/stop.sh" || true
fi

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

say "同步 nginx 模板与 SQL"
cp -f "$SRC_DIR/nginx/$SITE_NAME" "$TARGET_ROOT/nginx/$SITE_NAME"
[ -d "$SRC_DIR/sql" ] && rsync -a "$SRC_DIR/sql/" "$TARGET_ROOT/sql/" || true

say "设置权限（运行用户：$RUN_USER）"
chown -R "$RUN_USER:$RUN_GROUP" "$TARGET_ROOT"
chmod +x "$TARGET_ROOT/server/hunjie-server" \
         "$TARGET_ROOT/server"/*.sh
# nginx 以 www-data 身份读 web 和 uploads，父目录也得可进入
chmod 755 /data "$TARGET_ROOT" "$TARGET_ROOT/web" "$TARGET_ROOT/server" "$TARGET_ROOT/server/uploads"
find "$TARGET_ROOT/web" -type d -exec chmod 755 {} +
find "$TARGET_ROOT/web" -type f -exec chmod 644 {} +
find "$TARGET_ROOT/server/uploads" -type d -exec chmod 755 {} +
find "$TARGET_ROOT/server/uploads" -type f -exec chmod 644 {} +

say "安装 nginx 站点"
cp -f "$TARGET_ROOT/nginx/$SITE_NAME" "/etc/nginx/sites-available/$SITE_NAME"
ln -sfn "/etc/nginx/sites-available/$SITE_NAME" "/etc/nginx/sites-enabled/$SITE_NAME"
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl enable --now nginx >/dev/null 2>&1 || true
systemctl reload nginx

say "启动后端"
sudo -u "$RUN_USER" "$TARGET_ROOT/server/start.sh"

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
