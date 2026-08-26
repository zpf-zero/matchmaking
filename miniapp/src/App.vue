<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { ws } from '@/utils/ws'

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

onLaunch(async () => {
  const store = useUserStore()
  store.fetchConfig()
  store.fetchDict()
  setupGlobalNotify(store)

  if (!store.isLogin) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }

  try {
    await store.fetchMe()
    // 资料没填完直接引导去填，否则首页是空的，用户不知道该干什么
    if (store.status === 0) {
      uni.reLaunch({ url: '/pages/profile/setup' })
      return
    }
    if (store.canInteract) ws.connect()
  } catch {
    // fetchMe 失败时 request 层已按状态码做了跳转
  }
})

onShow(() => {
  const store = useUserStore()
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
page,
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

page {
  background-color: var(--color-canvas);
  color: var(--color-ink);
  font-size: 28rpx;
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
  border-radius: 8rpx;
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
  page,
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
}

/* 全局通用类 */
.card {
  background: var(--color-surface-raised);
  border: 1rpx solid var(--color-border);
  border-radius: 24rpx;
  padding: 28rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 2rpx rgba(78, 43, 55, 0.018),
    0 14rpx 34rpx var(--color-shadow);
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
  border-radius: 12rpx;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  font-size: 29rpx;
  font-weight: 650;
  letter-spacing: 1rpx;
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
  border-radius: 12rpx;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  font-size: 29rpx;
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
  font-size: 28rpx;
  line-height: 1.7;
}

.tag {
  display: inline-block;
  padding: 5rpx 12rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
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
</style>
