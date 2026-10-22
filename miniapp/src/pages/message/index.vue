<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onShow, onUnload } from '@dcloudio/uni-app'
import UserAvatar from '@/components/UserAvatar.vue'
import SwipeCell from '@/components/SwipeCell.vue'
import { chatApi } from '@/api'
import type { Conversation } from '@/api/types'
import { useUserStore } from '@/stores/user'
import { parseTime } from '@/utils/time'
import { ws } from '@/utils/ws'

const store = useUserStore()
const list = ref<Conversation[]>([])
const loaded = ref(false)

function onNewMessage() {
  load()
}

/**
 * 和首页同一类冷启动时序问题：本页 onShow 早于 onLaunch 里的 `await fetchMe()`，
 * 那一刻 canInteract 一定是 false。旧写法直接 `return` 并把 loaded 置 true，
 * 冷启动深链进本页会错误地显示「通过审核后即可聊天」。
 * 这里改为「等待中不做判断」，等 store 就绪后由 watch 补一次。
 */
function ensureLoad() {
  if (!store.canInteract) {
    // 用户确实还没通过审核 —— 这才是真的空状态
    if (store.meLoaded) loaded.value = true
    return
  }
  load()
  store.fetchUnread()
}

onShow(() => {
  ensureLoad()
  ws.on('message', onNewMessage)
  ws.on('match', onNewMessage)
})

// store 就绪后补一次加载（冷启动首次 onShow 时 canInteract 恒为 false）
watch(
  () => store.canInteract,
  (ok) => {
    if (ok) ensureLoad()
  },
)

onUnload(() => {
  ws.off('message', onNewMessage)
  ws.off('match', onNewMessage)
})

/**
 * 骨架屏条件：还在等 store 就绪、或已就绪但会话列表还没回来。
 * 少了前一项，冷启动等待期会先闪一下「没有配对」的空状态；
 * 但 store 明确失败时不能再转圈，要交给错误态。
 */
const showSkeleton = computed(
  () => !loaded.value && !store.meError && (store.canInteract || !store.meLoaded),
)

/** 会话列表加载失败：给重试入口，别伪装成「没有新的配对」 */
const loadFailed = ref(false)

async function load() {
  try {
    list.value = await chatApi.conversations()
    loadFailed.value = false
  } catch {
    // 已提示。必须标记失败：否则会落到「还没有新的配对」空态，
    // 把「没加载成功」说成「确实没有」，用户会以为真没人找过自己
    loadFailed.value = true
  } finally {
    loaded.value = true
  }
}

async function retryLoad() {
  loadFailed.value = false
  loaded.value = false
  // store 也没拿到就一并补上，否则重试只拉会话列表仍然进不去
  if (!store.meLoaded) {
    try {
      await store.fetchMe()
    } catch {
      loaded.value = true
      return
    }
  }
  await load()
}

function goChat(c: Conversation) {
  uni.navigateTo({
    url: `/pages/chat/index?id=${c.id}&name=${encodeURIComponent(c.peer_nickname)}&peer=${c.peer_id}`,
  })
}

/** 左滑解除配对。危险操作，仍然走二次确认 */
function onUnmatch(c: Conversation) {
  uni.showModal({
    title: '解除配对',
    content: `解除后你和 ${c.peer_nickname} 将无法继续聊天，需要重新互相喜欢才能配对`,
    confirmText: '解除',
    confirmColor: '#a43e3e',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await chatApi.closeConversation(c.id)
        list.value = list.value.filter((item) => item.id !== c.id)
        uni.showToast({ title: '已解除配对', icon: 'none' })
      } catch {
        // 已提示
      }
    },
  })
}

function goInteraction(tab: number) {
  uni.navigateTo({ url: `/pages/interaction/list?tab=${tab}` })
}

function goBrowse() {
  uni.switchTab({ url: '/pages/index/index' })
}

function timeText(t: string | null) {
  const ts = parseTime(t)
  if (Number.isNaN(ts)) return ''
  const d = new Date(ts)
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  if (sameDay) {
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
  const diff = (now.getTime() - d.getTime()) / 86400000
  if (diff < 2) return '昨天'
  if (diff < 7) return `${Math.floor(diff)}天前`
  return `${d.getMonth() + 1}/${d.getDate()}`
}
</script>

<template>
  <view class="page">
    <!-- 互动入口 -->
    <view class="entries">
      <view class="entry primary" @click="goInteraction(0)">
        <view class="entry-copy">
          <text class="entry-text">谁喜欢我</text>
          <text class="entry-note">看看最近对你表达心意的人</text>
        </view>
        <view v-if="store.unread.liked_me" class="badge">{{ store.unread.liked_me }}</view>
      </view>
      <view class="entry" @click="goInteraction(1)">
        <text class="entry-text">我喜欢的</text>
        <text class="entry-note">回看表达过的心意</text>
      </view>
      <view class="entry" @click="goInteraction(2)">
        <text class="entry-text">谁看过我</text>
        <text class="entry-note">了解最近的访客</text>
        <view v-if="store.unread.viewed_me" class="badge">{{ store.unread.viewed_me }}</view>
      </view>
    </view>

    <view class="section-title">我的配对</view>

    <view v-if="showSkeleton" class="conversation-skeleton" aria-label="正在加载消息">
      <view v-for="n in 3" :key="n" class="skeleton-conv">
        <view class="skeleton skeleton-conv-avatar"></view>
        <view class="skeleton-conv-copy">
          <view class="skeleton skeleton-conv-line strong"></view>
          <view class="skeleton skeleton-conv-line"></view>
        </view>
      </view>
    </view>

    <view v-else-if="store.meError" class="empty-state">
      <view class="empty-mark">缘</view>
      <view class="empty-title">没能加载出来</view>
      <view class="empty-copy">可能是网络不太稳定，点下面重试一下。</view>
      <view class="empty-action" @click="retryLoad">重新加载</view>
    </view>

    <view v-else-if="!store.canInteract" class="empty-state">
      <view class="empty-mark">审</view>
      <view class="empty-title">通过审核后即可聊天</view>
      <view class="empty-copy">你仍然可以先浏览推荐，找到想认真认识的人。</view>
      <view class="empty-action" @click="goBrowse">先看看推荐</view>
    </view>

    <view v-else-if="list.length" class="conv-list">
      <SwipeCell
        v-for="(c, i) in list"
        :key="c.id"
        class="conv-cell"
        :class="{ first: i === 0, last: i === list.length - 1 }"
        @action="onUnmatch(c)"
      >
        <template #action>解除</template>

        <view class="conv" @click="goChat(c)">
          <view class="avatar-wrap">
            <UserAvatar :src="c.peer_avatar" :name="c.peer_nickname" :size="96" />
            <view v-if="c.peer_online" class="online"></view>
            <view v-if="c.unread" class="unread">{{ c.unread > 99 ? '99+' : c.unread }}</view>
          </view>
          <view class="conv-main">
            <view class="conv-top">
              <text class="conv-name">{{ c.peer_nickname }}</text>
              <text class="conv-time">{{ timeText(c.last_message_at) }}</text>
            </view>
            <view class="conv-msg">
              {{ c.last_message || '你们已配对，打个招呼吧' }}
            </view>
          </view>
        </view>
      </SwipeCell>
      <view class="swipe-hint">左滑会话可解除配对</view>
    </view>

    <view v-else-if="loadFailed" class="empty-state">
      <view class="empty-mark">缘</view>
      <view class="empty-title">没能加载出来</view>
      <view class="empty-copy">可能是网络不太稳定，点下面重试一下。</view>
      <view class="empty-action" @click="retryLoad">重新加载</view>
    </view>

    <view v-else-if="loaded" class="empty-state">
      <view class="empty-mark">缘</view>
      <view class="empty-title">还没有新的配对</view>
      <view class="empty-copy">互相喜欢后，你们就可以从这里开始聊天。</view>
      <view class="empty-action" @click="goBrowse">去看看推荐</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  min-height: calc(100dvh - 188rpx);
  padding: 24rpx 24rpx 48rpx;
}

.entries {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16rpx;
  margin-bottom: 36rpx;
}

.entry {
  min-height: 138rpx;
  padding: 26rpx;
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: left;
  position: relative;
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  box-shadow: $shadow-card;
  transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1),
    border-color 200ms ease, box-shadow 200ms ease;

  &:active {
    transform: scale(0.975);
    border-color: $border-strong;
    box-shadow: none;
  }

  &.primary {
    grid-column: 1 / -1;
    min-height: 118rpx;
    display: flex;
    align-items: flex-start;
    text-align: left;
    background: $bg-accent;
    border-color: $brand-light;
  }
}

.entry-copy {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.entry-text {
  font-size: $font-body;
  color: $text-main;
  font-weight: 650;
}

.entry-note {
  font-size: $font-micro;
  color: $text-weak;
  line-height: 1.55;
}

.badge {
  position: absolute;
  top: 18rpx;
  right: 18rpx;
  min-width: 32rpx;
  height: 32rpx;
  line-height: 32rpx;
  padding: 0 7rpx;
  border-radius: $radius-sm;
  background: $danger;
  color: $text-on-accent;
  font-size: $font-micro;
  text-align: center;
}

.section-title {
  font-size: $font-section;
  font-weight: 650;
  letter-spacing: -0.5rpx;
  padding: 0 4rpx 20rpx;
}

.conv-list {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  overflow: hidden;
  box-shadow: $shadow-card;
}

/* 单元格自己带背景，圆角和分隔线得由外层容器裁 */
.conv-cell {
  border-bottom: 1rpx solid $border;

  &.last {
    border-bottom: 0;
  }
}

.conv {
  display: flex;
  align-items: center;
  padding: 26rpx 24rpx;
  transition: background-color 180ms ease;

  &:active {
    background: $bg-subtle;
  }
}

/* SwipeCell 的操作区在右侧，底边线要跨过它才连得上 */
.conv-cell.last .swipe-action,
.conv-cell.last .conv {
  border-bottom: 0;
}

.swipe-hint {
  text-align: center;
  font-size: $font-micro;
  color: $text-placeholder;
  padding: 20rpx 0 24rpx;
  border-top: 1rpx solid $border;
}

.avatar-wrap {
  position: relative;
  flex-shrink: 0;
}

.online {
  position: absolute;
  right: 2rpx;
  bottom: 2rpx;
  width: 20rpx;
  height: 20rpx;
  border-radius: $radius-round;
  background: $success;
  border: 3rpx solid $bg-raised;
}

.unread {
  position: absolute;
  top: -6rpx;
  right: -10rpx;
  min-width: 32rpx;
  height: 32rpx;
  line-height: 32rpx;
  padding: 0 8rpx;
  border-radius: $radius-sm;
  background: $danger;
  color: $text-on-accent;
  font-size: $font-micro;
  text-align: center;
}

.conv-main {
  flex: 1;
  margin-left: 24rpx;
  min-width: 0;
}

.conv-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10rpx;
}

.conv-name {
  font-size: $font-lead;
  font-weight: 600;
  color: $text-main;
}

.conv-time {
  font-size: $font-micro;
  color: $text-placeholder;
  font-variant-numeric: tabular-nums;
}

.conv-msg {
  font-size: $font-mid;
  color: $text-weak;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-skeleton {
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  background: $bg-card;
  overflow: hidden;
}

.empty-state {
  padding: 92rpx 40rpx 64rpx;
  text-align: center;
}

.empty-mark {
  width: 76rpx;
  height: 76rpx;
  line-height: 76rpx;
  margin: 0 auto 26rpx;
  border-radius: $radius-lg;
  background: $brand-light;
  color: $brand;
  font-size: $font-body;
  font-weight: 650;
}

.empty-title {
  margin-bottom: 10rpx;
  color: $text-main;
  font-size: $font-lead;
  font-weight: 600;
}

.empty-copy {
  max-width: 480rpx;
  margin: 0 auto 30rpx;
  color: $text-weak;
  font-size: $font-mid;
  line-height: 1.7;
}

.empty-action {
  display: inline-block;
  min-width: 196rpx;
  height: 70rpx;
  line-height: 70rpx;
  padding: 0 24rpx;
  border: 1rpx solid $border-strong;
  border-radius: $radius;
  background: $bg-raised;
  color: $text-main;
  font-size: $font-mid;
  font-weight: 600;
  transition: transform 180ms ease, background-color 180ms ease;

  &:active {
    transform: scale(0.98);
    background: $bg-subtle;
  }
}

.skeleton-conv {
  display: flex;
  align-items: center;
  gap: 24rpx;
  padding: 26rpx 24rpx;
  border-bottom: 1rpx solid $border;

  &:last-child {
    border-bottom: 0;
  }
}

.skeleton-conv-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: $radius;
  flex-shrink: 0;
}

.skeleton-conv-copy {
  flex: 1;
}

.skeleton-conv-line {
  width: 72%;
  height: 20rpx;
  border-radius: $radius-sm;
  margin-top: 16rpx;

  &.strong {
    width: 42%;
    height: 26rpx;
    margin-top: 0;
  }
}
</style>
