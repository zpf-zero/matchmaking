# 部署目录说明

Windows 上打包 → Ubuntu 上一键部署，前端是 **H5**，后端是 **交叉编译出来的 linux/amd64 可执行文件 + nohup 后台进程**。

> AI 助手请看 [`AGENTS.md`](./AGENTS.md)，那份是操作速查 + 踩坑清单；也可以直接用 `/redeploy`。

## 目录结构

```text
deploy/
├── build.ps1                 本地打包入口（Windows PowerShell）
├── scripts/
│   └── fetch-avatars.mjs     把头像变成自托管文件，并生成地址改写 SQL
├── template/                 随包分发的静态文件，改这里而不是改产物
│   ├── deploy.sh             服务器端一键部署脚本
│   ├── nginx/matchmaking.conf
│   └── server/{start,stop,restart,status}.sh
├── AGENTS.md                 给 AI 助手的操作速查与踩坑清单
├── ubuntu/                   ← 产物，git 忽略，每次打包重建
└── matchmaking-ubuntu-amd64.zip  ← 产物
```

`ubuntu/` 和 zip 都是生成的，不要手改，改了下次打包就没了。要调整部署行为改 `template/` 下的文件。

## 一、本地打包

```powershell
powershell -ExecutionPolicy Bypass -File deploy\build.ps1
```

一条命令做完四件事：

1. **交叉编译后端** — `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`，带 `-trimpath -ldflags "-s -w"`。依赖全是纯 Go（MySQL 驱动是 `go-sql-driver`），不需要 CGO，Windows 上直接就能编出 Linux 二进制。
2. **构建 H5** — `npm run build:h5`，产物从 `miniapp/dist/build/h5` 复制到包里。构建完会自动把 `src/config.generated.ts` 还原成 dev 配置，不会影响本地接着 `npm run dev:h5`。
3. **归集头像** — 见下一节。
4. **组装打包** — shell 脚本统一转成 LF 无 BOM（CRLF 会让 Ubuntu 报 `bad interpreter: ^M`），nginx 模板里的 `__SERVER_NAME__` 用实际域名替换，最后压成 zip。

地址和数据库连接不在脚本里写死，从这两个文件读：

| 文件 | 提供 |
|---|---|
| `miniapp/config.prod.yaml` → `base_url` | 前端 API 地址、nginx `server_name`、头像域名 |
| `server/config.prod.yaml` → `mysql.*` | 打包时连库改写头像地址用 |

换域名时改这两个文件再重新打包即可，其余地方不用动。

常用参数：

| 参数 | 作用 |
|---|---|
| `-SkipAvatars` | 不连库、不下载头像（离线打包） |
| `-SkipDbWrite` | 头像照常下载打包，但不执行 UPDATE，SQL 留在 `ubuntu/sql/` 里 |
| `-NoZip` | 只出目录不打包 |

## 二、头像是怎么带上的

线上要能显示头像，不能依赖外网图床，也不能留着指向开发机的地址。打包时会扫一遍数据库，分三种情况处理：

| 数据库里的地址 | 处理 |
|---|---|
| `https://api.dicebear.com/...`（测试数据的占位头像） | 把原图下载下来存进包里 `uploads/avatar/seed/<openid>.png`，画面和原来一致 |
| `http://localhost:8080/...` 且本地文件正常 | 文件原样打进包，只把域名换成生产地址 |
| `http://localhost:8080/...` 但文件缺失或只有几十字节的空图 | 按性别生成一张占位头像补上 |
| 已经是生产地址 | 不动。如果本地没有对应文件会打印告警——那张图只在服务器上，包里带不走 |

改写语句会写进 `ubuntu/sql/fix-avatars.sql` 留档，同时直接执行到数据库（用 `-SkipDbWrite` 可以只生成不执行）。

nginx 侧 `/uploads/` 直接读磁盘（`root /data/matchmaking/server`），文件不在才回源给后端，比全部回源快，也不怕后端没起来时头像全挂。

## 三、服务器部署

把 zip 传上去，解压，跑一个脚本：

```bash
scp deploy/matchmaking-ubuntu-amd64.zip ubuntu@<服务器IP>:/tmp/
ssh ubuntu@<服务器IP>

sudo mkdir -p /data/matchmaking && cd /data/matchmaking
sudo unzip -o /tmp/matchmaking-ubuntu-amd64.zip
sudo bash deploy.sh
```

> zip 格式不保留 Unix 执行位，所以第一步用 `bash` 显式调用；`deploy.sh` 会给 `hunjie-server` 和其余 `.sh` 补上 `+x`。
> 另外 `Compress-Archive` 在 PowerShell 5.1 下写出的条目名是反斜杠，Linux 解压会得到一堆平铺的怪文件名，
> 所以 `build.ps1` 自己写 zip 条目并强制正斜杠，打完还会校验一遍。

`deploy.sh` 会：建目录 → 停旧进程 → 铺文件 → 装 nginx 站点 → 修权限 → 起后端 → 自检（后端 /health、nginx 转发、前端首页、随便挑一张图片）。可以重复执行。

两个要注意的点：

- **uploads 只合并不删除。** 前端目录是 `rsync --delete` 整体替换（清掉上个版本残留的 assets），但 `uploads/` 单独用不带 `--delete` 的 rsync，服务器上用户上传的图片不会被本地包覆盖掉。
- **后端不用 root 跑。** 脚本自己需要 root（建 `/data`、装 nginx），但会用 `$SUDO_USER` 把文件属主和后端进程切回你自己的账号。

## 四、日常运维

```bash
cd /data/matchmaking/server

./status.sh      # 进程 / 端口 / 健康检查 / 日志尾巴，一次看全
./restart.sh     # 重启
./stop.sh        # 停止（先 TERM，10 秒不退再 KILL）
./start.sh       # 启动（起完自动等 /health 通过才报成功）

tail -f logs/server.log
```

`start.sh` 除了 PID 文件还会查一遍进程表，PID 文件被误删也不会启出第二个实例来抢 8080。

## 五、前置条件与注意事项

- 云安全组开放 **22** 和 **80**；8080 不用对公网开放，nginx 从 `127.0.0.1:8080` 转发。
- **外部 MySQL 的白名单必须放行服务器公网 IP**，否则后端起不来。日志里会是连接超时。
- `server/config.prod.yaml` 里有数据库明文口令，包和仓库都不要外发。
- `jwt.secret` 目前还是 `hunjie-dev-secret-please-change-in-production`。换掉更安全，但换了之后所有已登录用户的 token 会失效，需要重新登录。
- 微信登录还是 mock 模式（`wechat.mock: true`），openid 就是前端传的 code。

## 六、换域名 / 上 HTTPS

1. 改 `miniapp/config.prod.yaml` 的 `base_url`。
2. 改 `server/config.prod.yaml` 的 `upload.base_url`。
3. 重新 `deploy\build.ps1` 打包（nginx `server_name` 会自动跟着变，头像地址也会一并改写）。
4. 上传部署。证书相关配置加在 `template/nginx/matchmaking.conf` 里。
