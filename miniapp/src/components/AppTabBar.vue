<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'

/**
 * 自绘底部导航。
 *
 * 为什么不用 pages.json 的原生 tabBar：原生 tabBar 的 list 是编译期静态的，
 * 微信没有提供「运行时隐藏某一项」的 API（hideTabBar 只能整条隐藏）。
 * 私聊开关是远程控制的，只能在页面里自己画一条，再把原生那条藏掉。
 *
 * 用法：四个 tab 页各自在页面根部放一次，并传入 active；
 * 页面 onShow 里调用 uni.hideTabBar() 把原生那条压掉。
 */
type TabKey = 'index' | 'search' | 'message' | 'mine'

const props = defineProps<{ active: TabKey }>()

const store = useUserStore()

interface TabItem {
  key: TabKey
  text: string
  icon: string
  path: string
}

const ALL_TABS: TabItem[] = [
  { key: 'index', text: '推荐', icon: 'home', path: '/pages/index/index' },
  { key: 'search', text: '筛选', icon: 'search', path: '/pages/search/index' },
  { key: 'message', text: '消息', icon: 'message', path: '/pages/message/index' },
  { key: 'mine', text: '我的', icon: 'mine', path: '/pages/mine/index' },
]

/**
 * 私聊开关关闭时「消息」整项不渲染，而不是只把页面内容换成「暂未开放」。
 * 默认值是 true：配置还没拉回来之前先按开放处理，避免冷启动三项闪成四项。
 */
const tabs = computed(() =>
  ALL_TABS.filter((t) => t.key !== 'message' || store.privateChatEnabled),
)

function iconPath(t: TabItem, active: boolean, dark = false) {
  return `/static/${dark ? 'tabbar-dark' : 'tabbar'}/${t.icon}${active ? '-active' : ''}.png`
}

function go(t: TabItem) {
  if (t.key === props.active) return
  uni.switchTab({ url: t.path, fail: () => {} })
}

/**
 * 用户正停在「消息」页时开关被远程关掉：这一页已经没有入口了，退回推荐，
 * 否则会停留在一条没有高亮项的 tab 上。
 */
onMounted(() => {
  if (props.active === 'message' && !store.privateChatEnabled) {
    uni.switchTab({ url: '/pages/index/index', fail: () => {} })
  }
})
</script>

<template>
  <view class="app-tabbar">
    <view
      v-for="t in tabs"
      :key="t.key"
      class="app-tabbar__item"
      :class="{ 'is-active': t.key === active }"
      @click="go(t)"
    >
      <image
        class="app-tabbar__icon only-light"
        :src="iconPath(t, t.key === active)"
        mode="aspectFit"
      />
      <image
        class="app-tabbar__icon only-dark"
        :src="iconPath(t, t.key === active, true)"
        mode="aspectFit"
      />
      <text class="app-tabbar__text">{{ t.text }}</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
/*
 * 外观对齐原生 tabBar：100rpx 内容高 + 底部安全区，浅色 #fcfaf9 / 深色 #211f20
 * （对应 theme.json 里的 tabBg）。渐变分隔线用 --color-border，深色下自动跟随。
 */
.app-tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 200;
  display: flex;
  align-items: center;
  padding-bottom: env(safe-area-inset-bottom);
  background: var(--color-surface);
  border-top: 1rpx solid var(--color-border);
}

.app-tabbar__item {
  flex: 1;
  height: 100rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
  transition: transform 160ms ease;

  &:active {
    transform: scale(0.94);
  }
}

.app-tabbar__icon {
  width: 46rpx;
  height: 46rpx;
}

.app-tabbar__text {
  font-size: 20rpx;
  line-height: 1.2;
  color: var(--color-text-muted);
}

.is-active .app-tabbar__text {
  color: var(--color-accent);
  font-weight: 600;
}

/* 原生 tabBar 已隐藏，item 图标要自己按系统深浅色各挂一张 */
@media (prefers-color-scheme: dark) {
  .only-light {
    display: none;
  }
}

@media (prefers-color-scheme: light) {
  .only-dark {
    display: none;
  }
}
</style>
