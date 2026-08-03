# 缘来婚恋 · 婚介小程序

严肃婚恋交友小程序。核心闭环：**微信登录 → 完善资料 → 人工审核 → 条件筛选 → 互相喜欢 → 站内聊天**。

- 前端：uni-app + Vue3 + TypeScript + Pinia（**C 端与管理端同一个小程序**）
- 后端：Go 1.24 + Gin + GORM + MySQL 8（**C 端与管理端同一套服务**）

设计文档见 `docs/`，开发进度见 `plan.md`。

---

## 目录结构

```
matchmaking/
├── docs/                    设计文档（产品 / 数据库 / 接口）
├── plan.md                  开发进度与待办
├── server/                  Go 后端
│   ├── cmd/main.go
│   ├── config.yaml          配置（微信凭证、数据库、超管白名单）
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
    └── src/
        ├── api/ stores/ utils/ components/
        └── pages/
            ├── login/ profile/ index/ search/ message/ chat/ mine/ user/ ...
            └── admin/       管理页面（role > 1 才显示入口）
```

---

## 快速开始

### 1. 数据库

```powershell
mysql -uroot -proot123456 -e "CREATE DATABASE IF NOT EXISTS hunjie DEFAULT CHARACTER SET utf8mb4"
mysql -uroot -proot123456 hunjie < server/migrations/001_init.sql
```

### 2. 后端

```powershell
cd server
go mod tidy
go run ./cmd -c config.yaml
# 服务跑在 http://localhost:8080
```

验证：`curl http://localhost:8080/health`

### 3. 小程序

```powershell
cd miniapp
npm install
npm run dev:mp-weixin
```

然后用**微信开发者工具**导入 `miniapp/dist/dev/mp-weixin` 目录。

⚠️ 开发者工具需勾选 **详情 → 本地设置 → 不校验合法域名**（`manifest.json` 里已配 `urlCheck: false`，但工具端也要确认）。

---

## Mock 模式（当前默认）

AppID / AppSecret 尚未到位，`config.yaml` 中 `wechat.mock: true`，此时：

| 功能 | Mock 行为 |
|---|---|
| 登录 | 登录页手填一个字符串作为 openid，**同一个值 = 同一个账号** |
| 手机号 | 资料向导第一步手填，不走微信授权 |
| 文本内容安全 | 走本地敏感词库 + 联系方式正则 |
| 图片内容安全 | 不做机器检测，全部转人工审核 |

**测试账号**：登录页填 `mock_admin` 即可获得超级管理员（配置在 `config.yaml` 的 `admin.super_admin_openids`）；填 `test_user_1`、`test_user_2` 等模拟普通用户。

### 拿到微信凭证后如何切换

改 `server/config.yaml` 三行，重启服务即可，**业务代码不用动**：

```yaml
wechat:
  mock: false
  app_id: "wx..."
  app_secret: "..."
```

再把 `miniapp/src/manifest.json` 的 `mp-weixin.appid` 填上。

---

## 角色与权限

**管理员不是独立账号体系**，就是 `user.role > 1` 的普通用户，同样微信登录。

| role | 名称 | 权限（向上包含） |
|---|---|---|
| 1 | 普通用户 | C 端功能 |
| 2 | 审核员 | + 资料审核、举报处理、数据概览 |
| 3 | 管理员 | + 用户封禁/解封 |
| 4 | 超级管理员 | 同上 |

- **首个超管**：`config.yaml` 的 `admin.super_admin_openids` 白名单，命中的 openid 登录时自动置 `role=4`
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

`miniapp/src/utils/request.ts` 里的 `BASE_URL` 默认是 `http://localhost:8080`，真机访问不到本机。改成局域网 IP：

```ts
export const BASE_URL = 'http://192.168.x.x:8080'
```

（`ipconfig` 查本机 IP，确保手机与电脑同一局域网，防火墙放行 8080。）

---

## 上线前必须补的

1. **主体资质** — 微信「婚恋交友」类目要求企业/个体工商户主体，个人主体无法过审
2. **《用户协议》《隐私政策》** — 目前登录页/设置页是占位文本
3. **域名 + ICP 备案 + HTTPS**
4. **对象存储** — 图片现在存本地磁盘，生产环境单机故障即丢数据
5. **tabBar 图标** — 目前只有文字，没有图标
6. `config.yaml` 的 `jwt.secret` 必须换掉
7. `server.mode` 改 `release`
