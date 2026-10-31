# Ubuntu H5 部署

> 部署形态：前端 H5 静态文件由 nginx 托管，后端交叉编译成 linux/amd64 可执行文件、以 nohup 后台进程运行。
> 打包脚本和详细说明在 [`deploy/README.md`](../deploy/README.md)，这里是操作流程。

| 项 | 值 |
|---|---|
| 服务器公网 IP | `81.70.160.36` |
| 前端目录 | `/data/matchmaking/web` |
| 后端目录 | `/data/matchmaking/server` |
| 数据库 | 外部 MySQL `101.35.245.54:3306/marry_link_dev` |
| 微信登录 | mock 模式 |

## 一、本地打包

在项目根目录执行：

```powershell
powershell -ExecutionPolicy Bypass -File deploy\build.ps1
```

产物：

- `deploy/ubuntu/` — 完整部署目录
- `deploy/matchmaking-ubuntu-amd64.zip` — 上传用的包（约 9 MB）

包含内容：

```text
web/                        H5 静态文件
server/hunjie-server        linux/amd64 可执行文件
server/config.prod.yaml     生产配置（含数据库口令）
server/configs/             敏感词补充词库
server/uploads/             随包带的用户头像与图片
server/{start,stop,restart,status}.sh
nginx/matchmaking.conf      站点配置（server_name 已填好）
sql/fix-avatars.sql         头像地址改写语句，留档
deploy.sh                   服务器端一键部署
README.txt
```

打包会顺带把数据库里的头像地址改写成本站自托管地址，并把图片文件一起打进包，线上不依赖外网图床。细节见 `deploy/README.md` 第二节。

## 二、上传并部署

```bash
scp deploy/matchmaking-ubuntu-amd64.zip ubuntu@81.70.160.36:/tmp/
ssh ubuntu@81.70.160.36

sudo mkdir -p /data/matchmaking && cd /data/matchmaking
sudo apt update && sudo apt install -y unzip
sudo unzip -o /tmp/matchmaking-ubuntu-amd64.zip
sudo bash deploy.sh
```

zip 不保留 Unix 执行位，所以用 `bash` 显式调用，脚本会给 `hunjie-server` 和其余 `.sh` 补上 `+x`。

`deploy.sh` 依次完成：建目录 → 停旧进程 → 铺文件 → 装 nginx 站点 → 修权限 → 起后端 → 自检，最后打印检查结果。可以重复执行，`uploads/` 只合并不删除。

部署完成后打开：

```text
http://81.70.160.36
```

## 三、前置条件

- 云安全组开放 `22` 和 `80`。`8066` 不用对公网开放，nginx 从 `127.0.0.1:8066` 转发。
- **外部 MySQL 白名单要放行 `81.70.160.36`**，否则后端连不上库，`/health` 不通。
- 服务器需要 `nginx`、`curl`、`rsync`，`deploy.sh` 检测到缺失会自动 `apt install`。

## 四、日常运维

```bash
cd /data/matchmaking/server

./status.sh      # 进程 / 端口 / 健康检查 / 日志尾巴
./restart.sh     # 重启后端
./stop.sh
./start.sh

tail -f logs/server.log
```

只更新前端时，重新打包后覆盖 `/data/matchmaking/web` 即可，后端不用重启：

```bash
sudo rsync -a --delete /tmp/新包/web/ /data/matchmaking/web/
sudo systemctl reload nginx
```

只更新后端时必须先停进程，否则覆盖二进制会报 `Text file busy`：

```bash
cd /data/matchmaking/server
./stop.sh
sudo cp /tmp/新包/server/hunjie-server .
./start.sh
```

直接重跑 `sudo bash deploy.sh` 也可以，它已经把停机顺序处理好了。

## 五、排查

```bash
# 后端进程和端口
ps -ef | grep hunjie-server
sudo ss -lntp | grep -E ':(80|8066)'

# 后端本身
curl http://127.0.0.1:8066/health

# 经过 nginx
curl http://81.70.160.36/health

# nginx 日志
sudo tail -f /var/log/nginx/matchmaking.error.log
```

常见问题：

| 现象 | 原因 |
|---|---|
| `/health` 502 | 后端没起来，看 `server/logs/server.log`，多半是数据库白名单没放行 |
| 头像 403 | uploads 目录权限，重跑 `deploy.sh` 会自动修成 755/644 |
| 头像 404 | 该图片只存在于旧服务器上，包里没有；打包时会有告警列出是哪些账号 |
| 页面白屏、刷新子路由 404 | nginx 站点没生效，`sudo nginx -t && sudo systemctl reload nginx` |
| 脚本报 `bad interpreter: ^M` | 脚本被 CRLF 污染了，重新用 `build.ps1` 打包 |

## 六、换域名与 HTTPS

1. 改 `miniapp/config.prod.yaml` 的 `base_url`。
2. 改 `server/config.prod.yaml` 的 `upload.base_url`。
3. 重新执行 `deploy\build.ps1`（nginx `server_name` 和数据库里的头像地址都会跟着变）。
4. 上传部署。证书配置加在 `deploy/template/nginx/matchmaking.conf` 里，不要直接改产物。

## 七、安全提醒

`server/config.prod.yaml` 含数据库明文口令，部署包和仓库都不要外发。`jwt.secret` 已于 2026-10-26 换成随机串（与开发配置不同值），更换时所有已登录用户需要重新登录一次，现已完成。
