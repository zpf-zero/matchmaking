# 部署速查（写给 AI 助手）

面向在本仓库工作的 AI 助手。人类看的说明在 [`README.md`](./README.md)，这份只记**怎么做**和**别踩什么**。

## 0. 一句话

Windows 上 `deploy\build.ps1` 一条命令产出 zip → 传到 Ubuntu → root 执行 `bash deploy.sh`。
后端是**交叉编译的 linux/amd64 二进制 + nohup 后台进程**（不是 systemd，不是 Docker）。

**前端有两个交付目标，别搞混**（这是最容易漏的一处）：

| 目标 | 构建命令 | 产物 | 去向 | 本目录是否管 |
|---|---|---|---|---|
| **H5** | `npm run build:h5` | `miniapp/dist/build/h5` | 打进 zip → nginx | ✅ `build.ps1` 自动做 |
| **微信小程序** | `npm run build:mp-weixin` | `miniapp/dist/build/mp-weixin` | 微信开发者工具导入 → 提审 | ❌ **不在打包范围内，需手动构建** |

> 所以 `build.ps1` **只覆盖 H5**。改了前端且这次要提审小程序，必须**另外**跑一次
> `npm run build:mp-weixin`（在 `miniapp/` 下），再用微信开发者工具导入 `dist/build/mp-weixin`。
> 只跑 `build.ps1` 不会更新小程序产物。

## 1. 环境事实

| 项 | 值 |
|---|---|
| 生产服务器 | `81.70.160.36`，前端 `/data/matchmaking/web`，后端 `/data/matchmaking/server`，后端端口 `8066`（8080 被同机 Docker 容器占用） |
| 数据库 | `101.35.245.54:3306/marry_link_dev`，**MySQL 5.6.50**，账号密码在 `server/config.prod.yaml` |
| 数据库共用 | 本地开发和线上是**同一个库**，改数据两边同时生效 |
| 本地 dev | 前端 vite `:5173`，后端 `:8080` |
| 微信登录 | mock 模式，openid 就是登录框里填的 code |
| 构建机 | Windows + **Windows PowerShell 5.1**，Go 1.24.13（go.mod 要求 1.25，靠 `GOTOOLCHAIN=auto` 自动下载） |

## 2. 改了什么 → 要做什么

| 改动 | 动作 |
|---|---|
| 前端 `miniapp/src/**` | ① 重跑 `build.ps1`（H5，传包 + `bash deploy.sh`）② **要提审小程序则另跑 `npm run build:mp-weixin`** |
| 后端 `server/**` | 重跑 `build.ps1`，传包，root 执行 `bash deploy.sh`（脚本会先停旧进程） |
| `deploy/template/**` | 重跑 `build.ps1`（模板是打包时才复制进产物的） |
| `*.prod.yaml` 里的地址/域名 | 改完重跑 `build.ps1`，nginx `server_name` 和数据库里的头像地址会自动跟着变 |
| 只往数据库加了用户 | 重跑 `build.ps1` 把新头像打进包，否则线上是空头像 |
| 只改 `deploy/README.md` / 本文件 | 什么都不用做 |

**不要手改 `deploy/ubuntu/` 或 zip**，那是产物，下次打包就被覆盖。要改行为改 `deploy/template/`。

## 3. 打包

```powershell
powershell -ExecutionPolicy Bypass -File deploy\build.ps1
```

产出 `deploy/ubuntu/`（目录）和 `deploy/matchmaking-ubuntu-amd64.zip`（约 9 MB），两者都 gitignore。
参数：`-SkipAvatars` 离线打包 / `-SkipDbWrite` 只生成 SQL 不写库 / `-NoZip`。

脚本内部顺序：交叉编译 → H5 构建 → 头像归集 → 组装 → 打 zip → 校验条目分隔符。
地址和数据库连接从 `miniapp/config.prod.yaml`（`base_url`）和 `server/config.prod.yaml`（`mysql.*`）读，**不要在脚本里硬编码**。

## 4. 服务器步骤

```bash
scp deploy/matchmaking-ubuntu-amd64.zip ubuntu@81.70.160.36:/tmp/
ssh root@81.70.160.36
mkdir -p /data/matchmaking && cd /data/matchmaking
unzip -o /tmp/matchmaking-ubuntu-amd64.zip
bash deploy.sh          # 注意是 bash，不是 ./
```

zip 顶层直接就是 `deploy.sh` / `web/` / `server/` / `nginx/` / `sql/`，没有多一层 `ubuntu/`。

`deploy.sh` 幂等，可重复执行；默认全程使用 root。它会先判断 nginx 配置是否与包内模板一致且语法有效，已配置好就跳过复制和 reload；随后停止全部旧 `hunjie-server`、同步文件、设置最小必要权限、启动新进程并跑四项自检。

权限规则：目录 `755`，Web 普通文件和 uploads 文件 `644`，后端二进制与 `.sh` 为 `755`，`config.prod.yaml` 为 `600`。不要把 server 下所有文件都设成 `755`，否则配置文件既带执行位又对其他用户可读，会暴露数据库口令。

进程管理（`/data/matchmaking/server/` 下）：`status.sh` / `start.sh` / `stop.sh` / `restart.sh`。

## 5. 头像机制（容易忽略的一块）

线上头像必须自托管，不能依赖 dicebear 外链，也不能留 `localhost:8080` 地址。
`build.ps1` 会扫全库头像交给 `scripts/fetch-avatars.mjs`，按情况：

- dicebear 外链 → 抓原图存 `uploads/avatar/seed/<openid>.png`
- `localhost` 地址 + 本地文件正常 → 带走文件，只换域名
- `localhost` 地址 + 文件缺失或 <512 字节（1×1 空图）→ 按 openid+gender 生成一张
- 已是生产地址 + 本地有文件 → 带走
- 已是生产地址 + 本地没有 + 在 `avatar/seed/` 下 → **按 openid 重新生成**（同 seed 出同一张图）
- 已是生产地址 + 本地没有 + 不是生成的 → 无法恢复，脚本会列出来

最后一条分支是**幂等性的关键**：第二次打包时库里已经是生产地址，如果不认这类地址，包里就会缺图而库还指着它们。改这个脚本时别把它删了。

下载的图会同时写进包和本地 `server/uploads/`，所以第二次打包起可以完全离线（走本地缓存）。

已知无法恢复的：`mock_admin` 的头像只在服务器上，`uploads/` 是只合并不删除，所以只要服务器上还在就没事。

## 6. 坑（都是实际踩过的）

| 坑 | 后果 | 已有的防护 |
|---|---|---|
| `build.ps1` 存成 UTF-8 **无 BOM** | PS 5.1 按 ANSI 解析，中文打乱引号，一堆语法错 | 文件已带 BOM，**编辑后要确认 BOM 还在** |
| 用 `Compress-Archive` 打包 | 条目名是反斜杠，Linux 解压出一堆平铺怪文件名 | 自写 `New-UnixZip` 强制 `/`，打完还校验一遍 |
| shell 脚本 CRLF 或带 BOM | Ubuntu 报 `bad interpreter: ^M` | `Copy-AsUnixText` 统一转 LF 去 BOM |
| zip 不保留执行位 | `./deploy.sh` 报 Permission denied | root 使用 `bash deploy.sh`；脚本内部给其余文件补 `+x` |
| 按 key 读 YAML | `port` 命中 `server.port`(8066) 而不是 `mysql.port`(3306) | 用 `Read-YamlSectionValue` 按段落读 |
| `rsync --delete` 用在 `uploads/` | 删掉服务器上用户上传的图片 | `web/` 才用 `--delete`，`uploads/` 只合并 |
| 不 `cd` 到 server 目录就启动 | `sensitive.Init("configs/sensitive_words.txt")` 是**相对路径**，词库加载不到（不报错，静默退回内置词库） | `start.sh` 里 `cd "$BASE_DIR"` |
| 构建完忘了还原前端配置 | 本地 `npm run dev:h5` 打到生产地址 | `build.ps1` 末尾自动 `npm run config:dev` |

## 7. 本机工具的注意事项

- **mysql 客户端**（ServBay 的 `mysql.cmd`）：必须用长选项 `--host=x --user=y`，短选项 `-h101.35.245.54` 会被拆坏（报 `Unknown MySQL server host '101'`）。密码走 `$env:MYSQL_PWD`。
- **执行 .sql 文件**：PowerShell 没有 `<` 重定向，用 `cmd /c "mysql ... < ""path"""`。别用 `Get-Content | mysql`，会把 UTF-8 搞坏。
- **mysqldump**：这个库的账号没有 PROCESS 权限，要加 `--no-tablespaces`；`--column-statistics=0` 不被支持（服务端是 5.6）。
- **PowerShell 写含中文的文件**：用 Write 工具直接写，别用 `Get-Content | Set-Content` 转存，会乱码。
- **工作目录会在多次调用间保留**，`Remove-Item` 删目录前先 `Set-Location` 出去，否则报被占用。

## 8. 验证

服务器侧（`deploy.sh` 已自动跑这四项）：

```bash
curl http://127.0.0.1:8066/health        # 后端本身
curl http://81.70.160.36/health          # 经 nginx
curl -o /dev/null http://81.70.160.36/   # 前端首页
curl -o /dev/null http://81.70.160.36/uploads/avatar/seed/seed_m01.png   # 头像
```

打包侧值得确认的：

```powershell
# zip 条目分隔符（build.ps1 已内置校验，失败会 throw）
# 前端确实打进了生产地址、没残留 localhost
Select-String -Path deploy\ubuntu\web\assets\*.js -Pattern '81\.70\.160\.36' -List
Select-String -Path deploy\ubuntu\web\assets\*.js -Pattern 'localhost:8080' -List
# shell 语法（本机 bash 在 D:\software\Git\bin\bash.exe）
& "D:\software\Git\bin\bash.exe" -n deploy\ubuntu\deploy.sh
```

本地跑 H5 验证功能时：dev server 只监听 `::1`，playwright 里要用 `http://localhost:5173`，**用 `127.0.0.1` 连不上**。本机没装 playwright 浏览器，用系统 Chrome：`chromium.launch({ channel: 'chrome' })`。测试账号在登录框填 openid 即可（`test_user_1` 女/已审核，`mock_admin` 超管，`seed_m01`…`seed_f10` 是种子数据）。

## 9. 别做的事

- 别把 `server/config.prod.yaml` 或部署包发给无关的人 —— 里面有数据库明文口令。
- 别擅自改 `jwt.secret` —— 改了所有已登录用户 token 失效。它现在还是默认值，要换得先跟用户确认。
- 别在没确认的情况下对生产库做 DELETE / UPDATE。这个库和本地开发共用。
- 别把 `deploy/ubuntu/` 提交进 git。
