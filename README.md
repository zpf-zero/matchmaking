# 缘来婚恋 · 婚介小程序

严肃婚恋交友小程序。核心闭环：**微信登录 → 完善资料 → 人工审核 → 条件筛选 → 互相喜欢 → 站内聊天**。

- 前端：uni-app + Vue3 + TypeScript + Pinia（**C 端与管理端同一个小程序**）
- 后端：Go 1.25 + Gin + GORM + MySQL 5.6+（**C 端与管理端同一套服务**）

设计文档见 `docs/`。

---

## 功能一览

**C 端**
- 微信登录 + 手机号绑定（凭证到位前走 Mock 模式）
- 4 步资料填写向导（**草稿自动保存**，中途退出重进可恢复）、择偶要求、相册管理
- 首页推荐流（异性 + 年龄/身高/学历/收入/婚史/子女/期望城市过滤 + 双向匹配度展示，按活跃排序）
- 多条件筛选（9 个维度，**条件自动记忆**）
- 喜欢 / 不喜欢，互相喜欢即配对；谁喜欢我 / 我喜欢谁 / 谁看过我
- 站内聊天：WebSocket 实时收发、图片消息、**2 分钟内撤回**、**「正在输入」提示**、时间分隔线、空会话破冰话题、发送超时提示与点击重试
- **全局实时通知**：任意页面被喜欢弹 toast、配对成功弹窗直达聊天；消息 Tab 未读角标
- 举报 / 拉黑（双向不可见）/ 账号注销
- 敏感词 + 联系方式检测（防私下引流与诈骗）
- 用户协议 / 隐私政策草案页、登录勾选、TabBar 图标和主要空状态

**管理端**（同一个小程序内，`role > 1` 可见）
- 数据概览、资料审核（含相册一屏审完）、举报处理、用户管理（封禁/解封）
- 未提交的资料草稿不会进入审核队列；已通过用户修改资料或新增照片会自动进入复审

---

## 关键业务规则

- **审核队列**：新用户只有正式提交审核后才出现；已通过用户修改资料或新增照片时进入复审，待审照片不对外展示。
- **推荐结果**：仅展示状态正常、资料审核通过、异性且双方未拉黑的用户；推荐流额外排除我已喜欢或跳过的人。
- **活跃排序**：鉴权请求最多每 5 分钟更新一次活跃时间；年龄字段每天自动批量修正一次。
- **推荐分页**：喜欢后立即移除本地卡片并重新拉取第一页，避免 OFFSET 分页因结果集变化漏数据。
- **聊天可靠性**：WebSocket 未连接时走 HTTP 发送；发送后 10 秒未收到 ACK 会显示失败，可点击重试。
- **开发阶段降级**：AppID/AppSecret 未到位时保留 Mock 登录、手填手机号和本地文本过滤，不启用必须依赖真实微信凭证的能力。

---

## 目录结构

```
matchmaking/
├── docs/                    设计文档（产品 / 数据库 / 接口）
├── server/                  Go 后端
│   ├── cmd/main.go
│   ├── config.dev.yaml      开发配置（默认）
│   ├── config.prod.yaml     生产配置（域名和凭证申请后填写）
│   ├── configs/             敏感词库
│   ├── migrations/          建表脚本
│   ├── uploads/             图片存储
│   └── internal/
│       ├── config/ db/ model/ dto/
│       ├── service/         业务逻辑（直接用 gorm，无 repository 层）
│       ├── handler/         HTTP 处理
│       ├── middleware/      Auth / RequireBrowse / RequireActive / RequireRole
│       ├── router/ ws/
│       └── pkg/             wechat / jwt / response / upload / sensitive / logger
└── miniapp/                 uni-app 小程序
    ├── config.dev.yaml      开发接口地址
    ├── config.prod.yaml     生产接口地址
    ├── scripts/             构建前生成前端配置
    └── src/
        ├── config.generated.ts  YAML 自动生成的接口地址
        ├── api/ stores/ utils/ components/
        ├── static/tabbar/       TabBar 图标
        └── pages/
            ├── login/ profile/ index/ search/ message/ chat/ mine/ user/ ...
            ├── common/      状态页、用户协议、隐私政策
            └── admin/       管理页面（role > 1 才显示入口）
```

---

## 快速开始

### 1. 数据库

当前使用**云数据库**（MySQL 5.6，连接信息见 `mysql连接方式.txt`），`server/config.dev.yaml` 已配置好，表结构与测试数据已同步，无需额外操作。

如需重建或换库，执行建表脚本即可：

```powershell
# 示例：在目标库上建表（本地/云均适用）
mysql --host=<HOST> --user=<USER> -p <DATABASE> < server/migrations/001_init.sql
```

> 注意：建表脚本兼容 MySQL 5.6 ~ 8.x。云库是 5.6，本地是 8.4，勿用 8.x 的 mysqldump 直接导结构到 5.6（会带不兼容语法），结构统一以 `migrations/001_init.sql` 为准。

### 2. 后端

```powershell
cd server
go mod tidy
go run ./cmd -c config.dev.yaml
# 服务跑在 http://localhost:8080
```

验证：`curl http://localhost:8080/health`

### 3. 小程序

```powershell
cd miniapp
npm install
npm run dev:mp-weixin
npm run dev:h5
```

不同命令会在构建前读取对应 YAML 并生成 `src/config.generated.ts`：

| 命令 | 配置文件 | 输出 |
|---|---|---|
| `npm run dev:mp-weixin` | `config.dev.yaml` | `dist/dev/mp-weixin` |
| `npm run build:mp-weixin` | `config.prod.yaml` | `dist/build/mp-weixin` |
| `npm run dev:h5` | `config.dev.yaml` | H5 开发服务 |
| `npm run build:h5` | `config.prod.yaml` | H5 构建目录 |

然后用**微信开发者工具**导入对应的微信小程序输出目录。`miniapp/dist` 是构建产物，已从源码版本管理中移除。

⚠️ 开发者工具需勾选 **详情 → 本地设置 → 不校验合法域名**（`manifest.json` 里已配 `urlCheck: false`，但工具端也要确认）。

---

## Mock 模式（当前默认）

AppID / AppSecret 尚未到位，`server/config.dev.yaml` 中 `wechat.mock: true`，此时：

| 功能 | Mock 行为 |
|---|---|
| 登录 | 登录页手填一个字符串作为 openid，**同一个值 = 同一个账号** |
| 手机号 | 资料向导第一步手填，不走微信授权 |
| 文本内容安全 | 走本地敏感词库 + 联系方式正则 |
| 图片内容安全 | 头像和相册走人工审核；聊天图片保留举报处理能力，真实微信图片安全接口后续接入 |

**测试账号**：登录页填 `mock_admin` 即可获得超级管理员（配置在 `server/config.dev.yaml` 的 `admin.super_admin_openids`）；填 `test_user_1`、`test_user_2` 等模拟普通用户。

### 拿到微信凭证后如何切换

在部署使用的 `server/config.prod.yaml` 中填写以下三项，重启服务即可，**业务代码不用动**：

```yaml
wechat:
  mock: false
  app_id: "wx..."
  app_secret: "..."
```

再把 `miniapp/src/manifest.json` 的 `mp-weixin.appid` 填上，并在 `miniapp/config.prod.yaml` 中填写正式 HTTPS 服务地址。

---

## 角色与权限

**管理员不是独立账号体系**，就是 `user.role > 1` 的普通用户，同样微信登录。

| role | 名称 | 权限（向上包含） |
|---|---|---|
| 1 | 普通用户 | C 端功能 |
| 2 | 审核员 | + 资料审核、举报处理、数据概览 |
| 3 | 管理员 | + 用户封禁/解封 |
| 4 | 超级管理员 | 同上 |

- **首个超管**：服务端 YAML 的 `admin.super_admin_openids` 白名单，命中的 openid 登录时自动置 `role=4`
- **其他管理员**：直接改库 `UPDATE user SET role=2 WHERE id=?`
- **授权判定**：唯一判定点是 `middleware.RequireRole`。Auth 中间件每次请求回查数据库拿实时 role，**不信任 token 里的声明** —— 否则用户被降级后会有最长 7 天的越权窗口
- 前端 `role > 1` 才渲染管理入口，那只是体验优化，不是安全措施

---

## 账号状态流转

```
0 待完善 →(提交)→ 1 待审核 →(通过)→ 2 正常
                      ↓(拒绝)
                  3 已拒绝 →(重新提交)→ 1 待审核

任意状态 →(违规)→ 4 已封禁
```

status 与 role 是**两个正交维度**：管理员不填资料也能用管理功能；封禁用户即便 role 很高也一律拒绝。

接口按状态分三档拦截：

| 路由组 | 中间件 | 说明 |
|---|---|---|
| 资料填写、上传 | `Auth` | 登录即可，否则新用户没法填资料 |
| 浏览、筛选、详情 | `Auth + RequireBrowse` | 资料未完善不能看别人 |
| 喜欢、聊天 | `Auth + RequireActive` | 审核通过才能互动 |
| 管理接口 | `Auth + RequireRole(n)` | 按接口声明最低角色 |

---

## 真机调试

开发环境接口地址配置在 `miniapp/config.dev.yaml`。真机访问不到电脑的 `localhost`，改成局域网 IP：

```yaml
base_url: "http://192.168.x.x:8080"
```

重新执行 `npm run dev:mp-weixin` 后，构建前脚本会自动生成前端配置文件。生产地址后续填写到 `miniapp/config.prod.yaml`。

（`ipconfig` 查本机 IP，确保手机与电脑同一局域网，防火墙放行 8080。）

---

## Ubuntu H5 部署包

当前已经按公网 IP `81.70.160.36` 生成可直接上传的 Ubuntu AMD64 部署包：

- `deploy/ubuntu/web` → `/data/matchmaking/web`
- `deploy/ubuntu/server` → `/data/matchmaking/server`
- `deploy/ubuntu/nginx/matchmaking.conf` → Nginx 站点配置模板
- `deploy/matchmaking-ubuntu-amd64.zip` → 可直接上传并解压的完整部署包

后端使用现有外部 MySQL，H5 和图片地址暂时使用 HTTP 公网 IP。完整上传、启动、日志及 Nginx 配置步骤见 `docs/04-Ubuntu部署.md`。

---

## 上线前必须补的

1. **主体资质** — 微信「婚恋交友」类目要求企业/个体工商户主体，个人主体无法过审
2. **《用户协议》《隐私政策》最终信息** — 页面和草案已完成，需补充正式运营主体、联系方式和保存期限
3. **域名 + ICP 备案 + HTTPS**
4. **生产配置** — 填写 `server/config.prod.yaml` 与 `miniapp/config.prod.yaml`
5. **对象存储** — 图片现在存本地磁盘，生产环境单机故障即丢数据
6. **微信凭证** — 填写 AppID/AppSecret，关闭 Mock，并配置小程序隐私保护指引
