<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { ws } from '@/utils/ws'

/**
 * 微信隐私授权兜底。
 *
 * 微信规定：凡涉及用户个人信息的接口（getPhoneNumber / chooseAvatar / chooseImage 等），
 * 必须先在后台配置《用户隐私保护指引》，且用户同意后才能调用；
 * 否则真机直接 fail（errno:104 / api scope is not declared in the privacy agreement）。
 *
 * 这里用官方 onNeedPrivacyAuthorization 做全局兜底：一旦有接口触发隐私需求，
 * 就弹一次官方隐私授权框，用户同意后全小程序所有隐私接口解锁。
 * 不写自定义弹窗（自定义容易漏掉同步步骤，且会报 buttonId 错误）。
 *
 * 【待申请】需在小程序后台「设置 → 服务内容声明 → 用户隐私保护指引」完成声明
 *   —— docs/07-待申请事项清单.md 第 6 项。本函数只是用户侧交互，
 *   后台未声明时即使同意也调不通。
 */
function setupPrivacyGuard() {
  // #ifdef MP-WEIXIN
  const wxApi = (typeof wx !== 'undefined' ? wx : null) as any
  if (!wxApi) return

  // onNeedPrivacyAuthorization：微信在有隐私接口待授权时回调（基础库 2.32.3+）
  if (typeof wxApi.onNeedPrivacyAuthorization === 'function') {
    wxApi.onNeedPrivacyAuthorization((resolve: any) => {
      // 用官方隐私授权弹窗；selectPrivacyContract 可跳转隐私协议全文
      wxApi.requirePrivacyAuthorize({
        success: () => resolve({ event: 'agree', buttonId: 'agree-btn' }),
        fail: () => resolve({ event: 'disagree' }),
      })
    })
  }
  // #endif
}

/**
 * 全局实时通知。
 * 用户在任何页面被喜欢/配对成功都能立刻感知，
 * 而不是要等切到消息页才发现。
 */
function setupGlobalNotify(store: ReturnType<typeof useUserStore>) {
  ws.on('like', (data: any) => {
    store.fetchUnread()
    uni.showToast({
      title: `${data?.nickname || '有人'} 喜欢了你`,
      icon: 'none',
      duration: 2500,
    })
  })

  ws.on('match', (data: any) => {
    store.fetchUnread()
    const name = data?.peer?.nickname || '对方'
    uni.showModal({
      title: '配对成功',
      content: `你和 ${name} 互相喜欢，现在可以聊天了`,
      confirmText: '去聊天',
      cancelText: '稍后',
      success: (r) => {
        if (r.confirm && data?.conversation_id) {
          uni.navigateTo({
            url: `/pages/chat/index?id=${data.conversation_id}&name=${encodeURIComponent(name)}&peer=${data?.peer?.id || ''}`,
          })
        }
      },
    })
  })

  // 收到新消息刷新红点（聊天页会自己处理消息本身）
  ws.on('message', () => store.fetchUnread())
}

/**
 * 捕获邀请码。
 *
 * 三条入口都要处理：
 *   1. 好友转发卡片 → 参数在 options.query 里，如 { path, query: { invite: 'AB12CD' } }
 *   2. 小程序码（太阳码）→ 同样在 options.query.scene，是 URLEncode 过的 "i=AB12CD"
 *   3. 小程序已在后台时被卡片唤起 → onLaunch 不会再触发，只有 onShow 带参数
 *
 * ⚠️ 关于参数层级（曾踩坑）：
 * 微信 getLaunchOptionsSync() / onAppShow 给的是**原生结构**
 *   { path: string, scene: number, query: {...}, referrerInfo }
 * 业务参数一律在 `query` 下，**不在顶层**。且顶层的 `scene` 是数字场景值
 * （1047=小程序码扫描等），不是小程序码里带的那串字符。
 * uni-app 这里不做拍平，直接把原生对象转给 onLaunch/onShow。
 * 所以必须读 options.query.invite，读 options.invite 永远是 undefined。
 *
 * 只暂存到本地，不在此时绑定：绑定要等用户注册完（有 token）之后。
 * 暂存而不是立即提交，也让用户有机会在向导里手动改掉填错的码。
 */
const PENDING_INVITE_KEY = 'pending_invite_code'

function captureInvite(options: any) {
  if (!options) return
  // 业务参数在 query 下；兼容 uni-app 某些平台可能拍平到顶层的写法
  const q = options.query || {}
  let code = q.invite || q.invite_code || options.invite || options.invite_code || ''

  // 扫码进入：小程序码的参数在 query.scene，需解码，格式形如 "i=AB12CD"
  // （scene 只支持 ASCII 可见字符；页面 onLoad 里则在 options.scene，无 query 层）
  const rawScene = q.scene || (typeof options.scene === 'string' ? options.scene : '')
  if (!code && rawScene) {
    try {
      const decoded = decodeURIComponent(rawScene)
      const params: Record<string, string> = {}
      decoded.split('&').forEach((kv) => {
        const idx = kv.indexOf('=')
        if (idx > 0) params[kv.slice(0, idx)] = kv.slice(idx + 1)
      })
      code = params.i || params.invite || ''
    } catch {
      // scene 格式异常时忽略，不影响正常进入
    }
  }

  code = String(code).trim().toUpperCase()
  // 邀请码是 6 位字母数字（去掉易混淆字符的字母表）
  if (code && /^[0-9A-Z]{4,12}$/.test(code)) {
    uni.setStorageSync(PENDING_INVITE_KEY, code)
  }
}

/** 取并清除待绑定的邀请码 */
function takePendingInvite(): string {
  const code = uni.getStorageSync(PENDING_INVITE_KEY) || ''
  if (code) uni.removeStorageSync(PENDING_INVITE_KEY)
  return code
}

onLaunch(async (options) => {
  const store = useUserStore()
  captureInvite(options)
  setupPrivacyGuard()

  store.fetchConfig()
  store.fetchDict()
  setupGlobalNotify(store)

  if (!store.isLogin) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }

  try {
    await store.fetchMe()
    await flushPendingInvite(store)
    // 只有「普通用户且资料未提交」才引导去填资料；运营角色直接进各自首页。
    // 判据统一走 store.needsProfile（内含 role 判断），避免这里漏掉管理员。
    if (store.needsProfile) {
      uni.reLaunch({ url: '/pages/profile/setup' })
      return
    }
    if (store.canInteract) ws.connect()
  } catch {
    // fetchMe 失败时 request 层已按状态码做了跳转
  }
})

/**
 * 把待绑定的邀请码补交上去。
 *
 * 覆盖「老用户扫了朋友的码」和「新用户注册后回到首页」两种收尾场景。
 * 已绑定过的用户后端会直接返回「已经使用过邀请码」，静默忽略即可 ——
 * 不要打扰用户，更不要在用户没主动操作时弹窗。
 */
async function flushPendingInvite(store: ReturnType<typeof useUserStore>) {
  const code = uni.getStorageSync(PENDING_INVITE_KEY) || ''
  if (!code) return
  if (store.status === 0) return // 还在填资料，向导提交时会带上，这里不抢跑
  try {
    const { inviteApi } = await import('@/api')
    await inviteApi.bind(code)
    uni.removeStorageSync(PENDING_INVITE_KEY)
    uni.showToast({ title: '邀请码绑定成功', icon: 'none' })
  } catch {
    // 无效/已使用：清掉本地暂存，避免每次启动都重试
    uni.removeStorageSync(PENDING_INVITE_KEY)
  }
}

onShow((options: any) => {
  const store = useUserStore()

  // 小程序已在后台时，被分享卡片/小程序码唤起只会触发 onShow（不走 onLaunch），
  // 这里必须补捕一次，否则「老用户在小程序开着的情况下点分享链接」拿不到邀请码
  captureInvite(options)
  if (store.isLogin) flushPendingInvite(store)

  if (store.isLogin && store.canInteract) {
    // 小程序切后台会被断连，回前台重新接上
    ws.connect()
    store.fetchUnread()
  }
})

onHide(() => {
  ws.close()
})
</script>

<style lang="scss">
/*
 * 主题变量挂在 page 上：小程序端只认 page，不认 :root。
 * 注意不要把 :root 和 page 写进同一组选择器 —— WXSS 不支持 :root，
 * 整条规则会被丢弃，小程序端主色/背景/边框/深色模式会全部失效。
 * :root 只有 H5 需要（body 背景在 page 之外），用条件编译单独给。
 */
page {
  --color-canvas: #f4f1f2;
  --color-surface: #fcfaf9;
  --color-surface-raised: #fffdfc;
  --color-surface-subtle: #ece8e8;
  --color-surface-accent: #f7edef;
  --color-ink: #221f20;
  --color-text-secondary: #5c5558;
  --color-text-muted: #6b6266;
  --color-text-faint: #746b6f;
  --color-border: #ddd7d9;
  --color-border-strong: #8f8589;
  --color-accent: #963e58;
  --color-accent-strong: #762b43;
  --color-accent-soft: #f2e2e7;
  --color-on-accent: #fffafb;
  --color-success: #39705a;
  --color-success-soft: #e4eee9;
  --color-warning: #8a642c;
  --color-warning-soft: #f3ead8;
  --color-danger: #a43e3e;
  --color-danger-soft: #f4e3e1;
  --color-shadow: rgba(78, 43, 55, 0.055);
}

/* #ifdef H5 */
:root {
  --color-canvas: #f4f1f2;
  --color-surface: #fcfaf9;
  --color-surface-raised: #fffdfc;
  --color-surface-subtle: #ece8e8;
  --color-surface-accent: #f7edef;
  --color-ink: #221f20;
  --color-text-secondary: #5c5558;
  --color-text-muted: #6b6266;
  --color-text-faint: #746b6f;
  --color-border: #ddd7d9;
  --color-border-strong: #8f8589;
  --color-accent: #963e58;
  --color-accent-strong: #762b43;
  --color-accent-soft: #f2e2e7;
  --color-on-accent: #fffafb;
  --color-success: #39705a;
  --color-success-soft: #e4eee9;
  --color-warning: #8a642c;
  --color-warning-soft: #f3ead8;
  --color-danger: #a43e3e;
  --color-danger-soft: #f4e3e1;
  --color-shadow: rgba(78, 43, 55, 0.055);
}
/* #endif */


page {
  background-color: var(--color-canvas);
  color: var(--color-ink);
  font-size: $font-body;
  font-family: 'Avenir Next', 'SF Pro Display', 'PingFang SC', 'Microsoft YaHei',
    'Helvetica Neue', sans-serif;
  line-height: 1.5;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
}

body {
  background-color: var(--color-canvas);
}

page,
view,
text,
image,
input,
textarea,
button {
  box-sizing: border-box;
  -webkit-tap-highlight-color: transparent;
}

page,
scroll-view {
  scroll-behavior: smooth;
}

button::after {
  border: 0;
}

button,
input,
textarea {
  font: inherit;
}

button,
[role='button'] {
  cursor: pointer;
}

::selection {
  background: var(--color-accent-soft);
  color: var(--color-accent-strong);
}

input,
textarea {
  color: var(--color-ink);
  caret-color: var(--color-accent);
}

input:focus,
textarea:focus {
  outline: 2rpx solid var(--color-accent);
  outline-offset: 3rpx;
}

view:focus-visible,
button:focus-visible {
  outline: 2rpx solid var(--color-accent);
  outline-offset: 4rpx;
}

/* H5 的浏览器表面也跟随设计系统；小程序端会安全忽略这些选择器 */
::-webkit-scrollbar {
  width: 8rpx;
  height: 8rpx;
}

::-webkit-scrollbar-thumb {
  border-radius: $radius-round;
  background: var(--color-border-strong);
}

/* uni-app H5 原生导航与 TabBar 也进入同一套材质和字阶 */
.uni-page-head {
  width: 100% !important;
  max-width: 540px;
  left: 50% !important;
  right: auto !important;
  transform: translateX(-50%);
  background-color: var(--color-surface-raised) !important;
  border-bottom: 1rpx solid var(--color-border);
}

.uni-page-head__title {
  color: var(--color-ink) !important;
  font-weight: 650 !important;
  letter-spacing: -0.5rpx;
}

.uni-tabbar {
  width: 100% !important;
  max-width: 540px;
  left: 50% !important;
  right: auto !important;
  transform: translateX(-50%);
  background-color: var(--color-surface-raised) !important;
  box-shadow: 0 -10rpx 32rpx var(--color-shadow);
}

.uni-tabbar-border {
  background-color: var(--color-border) !important;
}

.uni-tabbar__label {
  font-weight: 520;
}

.uni-page-body {
  width: 100%;
  max-width: 540px;
  min-height: 100dvh;
  margin: 0 auto;
  background-color: var(--color-canvas);
}

/* 产品本身是移动端 H5，宽屏保持聚焦的单列画布，不把卡片拉成横幅 */
.page,
.setup,
.chat,
.agreement-page,
.login {
  width: 100%;
  max-width: 540px;
  margin-left: auto !important;
  margin-right: auto !important;
}

@media (min-width: 541px) {
  .uni-page-body {
    border-left: 1rpx solid var(--color-border);
    border-right: 1rpx solid var(--color-border);
  }

  .page > .actions,
  .setup > .actions,
  .chat > .input-bar {
    width: 100%;
    max-width: 540px;
    left: 50% !important;
    right: auto !important;
    transform: translateX(-50%);
  }
}

/* 支持系统主题的 H5 与微信基础库会自动切换 */
@media (prefers-color-scheme: dark) {
  page {
    --color-canvas: #191718;
    --color-surface: #211f20;
    --color-surface-raised: #282526;
    --color-surface-subtle: #302c2e;
    --color-surface-accent: #37262c;
    --color-ink: #f5eff1;
    --color-text-secondary: #d1c7ca;
    --color-text-muted: #b7acb0;
    --color-text-faint: #a79ca0;
    --color-border: #3c3638;
    --color-border-strong: #81767a;
    --color-accent: #df8aa3;
    --color-accent-strong: #eda0b5;
    --color-accent-soft: #472b35;
    --color-on-accent: #26151b;
    --color-success: #83b59d;
    --color-success-soft: #263b32;
    --color-warning: #cfaa6f;
    --color-warning-soft: #403626;
    --color-danger: #dd8a85;
    --color-danger-soft: #432c2a;
    --color-shadow: rgba(0, 0, 0, 0.22);
  }

  /* #ifdef H5 */
  :root {
    --color-canvas: #191718;
    --color-surface: #211f20;
    --color-surface-raised: #282526;
    --color-surface-subtle: #302c2e;
    --color-surface-accent: #37262c;
    --color-ink: #f5eff1;
    --color-text-secondary: #d1c7ca;
    --color-text-muted: #b7acb0;
    --color-text-faint: #a79ca0;
    --color-border: #3c3638;
    --color-border-strong: #81767a;
    --color-accent: #df8aa3;
    --color-accent-strong: #eda0b5;
    --color-accent-soft: #472b35;
    --color-on-accent: #26151b;
    --color-success: #83b59d;
    --color-success-soft: #263b32;
    --color-warning: #cfaa6f;
    --color-warning-soft: #403626;
    --color-danger: #dd8a85;
    --color-danger-soft: #432c2a;
    --color-shadow: rgba(0, 0, 0, 0.22);
  }
  /* #endif */
}

/* 全局通用类 */
.card {
  background: var(--color-surface-raised);
  border: 1rpx solid var(--color-border);
  border-radius: $radius-lg;
  padding: 28rpx;
  margin-bottom: 24rpx;
  box-shadow: $shadow-card;
}

.row {
  display: flex;
  align-items: center;
}

.row-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.btn-primary {
  background: var(--color-accent);
  color: $text-on-accent;
  border: 1rpx solid var(--color-accent);
  border-radius: $radius;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  font-size: $font-body;
  font-weight: 650;
  letter-spacing: 0.5rpx;
  white-space: nowrap;
  transition: transform 180ms cubic-bezier(0.16, 1, 0.3, 1),
    background-color 180ms ease, opacity 180ms ease;

  &:active {
    transform: scale(0.98);
    background: var(--color-accent-strong);
  }

  &.disabled {
    background: var(--color-surface-subtle);
    border-color: var(--color-surface-subtle);
    color: var(--color-text-faint);
    opacity: 0.82;
  }
}

.btn-plain {
  background: var(--color-surface-raised);
  color: var(--color-text-secondary);
  border: 1rpx solid var(--color-border-strong);
  border-radius: $radius;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  font-size: $font-body;
  font-weight: 500;
  white-space: nowrap;
  transition: transform 180ms cubic-bezier(0.16, 1, 0.3, 1),
    background-color 180ms ease;

  &:active {
    transform: scale(0.98);
    background: var(--color-surface-subtle);
  }
}

.empty {
  padding: 144rpx 48rpx;
  text-align: center;
  color: var(--color-text-muted);
  font-size: $font-body;
  line-height: 1.7;
}

.tag {
  display: inline-block;
  padding: 5rpx 12rpx;
  border-radius: $radius-sm;
  font-size: $font-micro;
  background: var(--color-surface-subtle);
  color: var(--color-text-secondary);
  margin-right: 10rpx;
}

.tag-brand {
  background: var(--color-accent-soft);
  color: var(--color-accent);
}

.num,
.count,
.uid,
.step-sub {
  font-variant-numeric: tabular-nums;
}

.skeleton {
  position: relative;
  overflow: hidden;
  background: var(--color-surface-subtle);
}

.skeleton::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.28);
  transform: translateX(-100%);
}

@media (prefers-reduced-motion: no-preference) {
  .page,
  .login,
  .setup,
  .chat,
  .agreement-page {
    animation: page-enter 520ms cubic-bezier(0.16, 1, 0.3, 1) both;
  }

  .skeleton::after {
    animation: skeleton-sweep 1.6s ease-in-out infinite;
  }

  @keyframes page-enter {
    from {
      opacity: 0;
      transform: translateY(12rpx);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes skeleton-sweep {
    to {
      transform: translateX(100%);
    }
  }
}

/* WXSS 不支持 * 通配符选择器（基础库 3.16 直接编译报错），这块只给 H5 */
/* #ifndef MP-WEIXIN */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 1ms !important;
    animation-iteration-count: 1 !important;
    scroll-behavior: auto !important;
    transition-duration: 1ms !important;
  }
}
/* #endif */
</style>
