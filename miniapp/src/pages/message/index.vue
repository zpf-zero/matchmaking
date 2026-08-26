<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onUnload } from '@dcloudio/uni-app'
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

onShow(() => {
  if (!store.canInteract) {
    loaded.value = true
    return
  }
  load()
  store.fetchUnread()
  ws.on('message', onNewMessage)
  ws.on('match', onNewMessage)
})

onUnload(() => {
  ws.off('message', onNewMessage)
  ws.off('match', onNewMessage)
})

async function load() {
  try {
    list.value = await chatApi.conversations()
  } catch {
    // 已提示
  } finally {
    loaded.value = true
  }
}

function goChat(c: Conversation) {
  uni.navigateTo({
    url: `/pages/chat/index?id=${c.id}&name=${encodeURIComponent(c.peer_nickname)}&peer=${c.peer_id}`,
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

    <view v-if="!loaded && store.canInteract" class="conversation-skeleton" aria-label="正在加载消息">
      <view v-for="n in 3" :key="n" class="skeleton-conv">
        <view class="skeleton skeleton-conv-avatar"></view>
        <view class="skeleton-conv-copy">
          <view class="skeleton skeleton-conv-line strong"></view>
          <view class="skeleton skeleton-conv-line"></view>
        </view>
      </view>
    </view>

    <view v-if="!store.canInteract" class="empty-state">
      <view class="empty-mark">审</view>
      <view class="empty-title">通过审核后即可聊天</view>
      <view class="empty-copy">你仍然可以先浏览推荐，找到想认真认识的人。</view>
      <view class="empty-action" @click="goBrowse">先看看推荐</view>
    </view>

    <view v-else-if="list.length" class="conv-list">
      <view v-for="c in list" :key="c.id" class="conv" @click="goChat(c)">
        <view class="avatar-wrap">
          <image :src="c.peer_avatar" class="avatar" mode="aspectFill" />
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
  font-size: 29rpx;
  color: $text-main;
  font-weight: 650;
}

.entry-note {
  font-size: 22rpx;
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
  font-size: 20rpx;
  text-align: center;
}

.section-title {
  font-size: 34rpx;
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

.conv {
  display: flex;
  align-items: center;
  padding: 26rpx 24rpx;
  border-bottom: 1rpx solid $border;
  transition: background-color 180ms ease;

  &:active {
    background: $bg-subtle;
  }

  &:last-child {
    border-bottom: none;
  }
}

.avatar-wrap {
  position: relative;
  flex-shrink: 0;
}

.avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: $radius;
  background: $bg-subtle;
}

.online {
  position: absolute;
  right: 2rpx;
  bottom: 2rpx;
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: $success;
  border: 3rpx solid $bg-card;
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
  font-size: 20rpx;
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
  font-size: 30rpx;
  font-weight: 600;
  color: $text-main;
}

.conv-time {
  font-size: 22rpx;
  color: $text-placeholder;
  font-variant-numeric: tabular-nums;
}

.conv-msg {
  font-size: 26rpx;
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
  font-size: 29rpx;
  font-weight: 650;
}

.empty-title {
  margin-bottom: 10rpx;
  color: $text-main;
  font-size: 30rpx;
  font-weight: 600;
}

.empty-copy {
  max-width: 480rpx;
  margin: 0 auto 30rpx;
  color: $text-weak;
  font-size: 25rpx;
  line-height: 1.7;
}

.empty-action {
  display: inline-block;
  min-width: 196rpx;
  height: 70rpx;
  line-height: 70rpx;
  padding: 0 24rpx;
  border: 1rpx solid $border-strong;
  border-radius: $radius-sm;
  background: $bg-raised;
  color: $text-main;
  font-size: 25rpx;
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
