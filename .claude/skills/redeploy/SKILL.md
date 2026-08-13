---
name: redeploy
description: 重新打包并部署本项目到 Ubuntu 服务器（H5 前端 + Go 后端）。当用户要求部署、发布、上线、更新服务器、重新打包、出部署包，或改完前后端代码问"怎么发到服务器"时使用。也用于排查部署失败、头像在线上不显示、后端起不来等部署相关问题。
---

# 重新部署到 Ubuntu

## 第一步：先读完整速查

```
deploy/AGENTS.md
```

**必须先读它再动手。** 里面有环境事实、改动对照表、头像自托管机制、以及一堆实际踩过的坑（PowerShell 5.1 的 BOM 要求、`Compress-Archive` 反斜杠 bug、shell 脚本行尾、YAML 段落读取、mysql 客户端长选项等）。跳过这一步大概率会重犯已经修过的问题。

配套文件：

- `deploy/README.md` — 人类看的说明，结构和参数更详细
- `docs/04-Ubuntu部署.md` — 操作流程与排查表
- `deploy/build.ps1` — 打包入口（**UTF-8 with BOM，编辑后确认 BOM 还在**）
- `deploy/template/` — 随包分发的脚本与 nginx 配置，改部署行为改这里
- `deploy/scripts/fetch-avatars.mjs` — 头像自托管处理

## 第二步：判断要不要重新打包

对照 `deploy/AGENTS.md` 第 2 节。简单说：动了 `miniapp/src`、`server/`、`deploy/template/`、任一 `*.prod.yaml`，或往数据库加了用户（头像要打进包），都要重新打包。

## 第三步：打包

```powershell
powershell -ExecutionPolicy Bypass -File deploy\build.ps1
```

产出 `deploy/matchmaking-ubuntu-amd64.zip`（约 9 MB）。打包的是 `deploy/ubuntu/` 的内容，不是整个 `deploy/`。
脚本自带校验，中途 throw 就按报错处理，别跳过。

## 第四步：给用户部署步骤

打包机是 Windows，`scp` / `ssh` 到生产服务器通常要用户自己执行。把下面这段给他：

```bash
scp deploy/matchmaking-ubuntu-amd64.zip ubuntu@81.70.160.36:/tmp/
ssh ubuntu@81.70.160.36
sudo mkdir -p /data/matchmaking && cd /data/matchmaking
sudo unzip -o /tmp/matchmaking-ubuntu-amd64.zip
sudo bash deploy.sh
```

`deploy.sh` 幂等，会自己停旧进程、铺文件、修权限、装 nginx、起后端，最后打印四项自检。四项全 OK 即成功。

如果需要用户自己在终端执行，提示他可以用 `! <命令>` 前缀直接在会话里跑。

## 常见故障对照

| 现象 | 方向 |
|---|---|
| `/health` 502 | 后端没起来，看 `server/logs/server.log`，最常见是外部 MySQL 白名单没放行服务器 IP |
| 头像 403 | uploads 权限，重跑 `sudo bash deploy.sh` 会修成 755/644 |
| 头像 404 | 图片没打进包，见 `deploy/AGENTS.md` 第 5 节头像机制 |
| `bad interpreter: ^M` | 脚本被 CRLF 污染，重新用 `build.ps1` 打包 |
| 解压出一堆带反斜杠的怪文件名 | zip 是用 `Compress-Archive` 打的，必须用 `build.ps1` 里的 `New-UnixZip` |
| 覆盖二进制报 `Text file busy` | 先 `./stop.sh` |
