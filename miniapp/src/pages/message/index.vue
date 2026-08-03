<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onUnload } from '@dcloudio/uni-app'
import { chatApi } from '@/api'
import type { Conversation } from '@/api/types'
import { useUserStore } from '@/stores/user'
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

function timeText(t: string | null) {
  if (!t) return ''
  const d = new Date(t.replace(/-/g, '/').replace(/\..*$/, ''))
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
      <view class="entry" @click="goInteraction(0)">
        <view class="entry-icon like">♥</view>
        <text class="entry-text">谁喜欢我</text>
        <view v-if="store.unread.liked_me" class="badge">{{ store.unread.liked_me }}</view>
      </view>
      <view class="entry" @click="goInteraction(1)">
        <view class="entry-icon mine">☆</view>
        <text class="entry-text">我喜欢的</text>
      </view>
      <view class="entry" @click="goInteraction(2)">
        <view class="entry-icon view">◎</view>
        <text class="entry-text">谁看过我</text>
        <view v-if="store.unread.viewed_me" class="badge">{{ store.unread.viewed_me }}</view>
      </view>
    </view>

    <view class="section-title">我的配对</view>

    <view v-if="!store.canInteract" class="empty">
      资料审核通过后，与他人互相喜欢即可开始聊天
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

    <view v-else-if="loaded" class="empty">
      还没有配对，去首页看看有没有心动的人
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 20rpx 24rpx 40rpx;
}

.entries {
  display: flex;
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx 0;
  margin-bottom: 28rpx;
}

.entry {
  flex: 1;
  text-align: center;
  position: relative;
}

.entry-icon {
  width: 88rpx;
  height: 88rpx;
  line-height: 88rpx;
  margin: 0 auto 14rpx;
  border-radius: 50%;
  font-size: 38rpx;

  &.like {
    background: #fff0f3;
    color: #e8506d;
  }

  &.mine {
    background: #fff7e8;
    color: #ff8f1f;
  }

  &.view {
    background: #eef5ff;
    color: #3f7ef5;
  }
}

.entry-text {
  font-size: 24rpx;
  color: #646a73;
}

.badge {
  position: absolute;
  top: -6rpx;
  right: 50rpx;
  min-width: 32rpx;
  height: 32rpx;
  line-height: 32rpx;
  padding: 0 8rpx;
  border-radius: 16rpx;
  background: #f53f3f;
  color: #fff;
  font-size: 20rpx;
  text-align: center;
}

.section-title {
  font-size: 28rpx;
  font-weight: 600;
  padding: 8rpx 8rpx 20rpx;
}

.conv-list {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.conv {
  display: flex;
  align-items: center;
  padding: 24rpx;
  border-bottom: 1rpx solid #f2f3f5;

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
  border-radius: 50%;
  background: #f5f6f8;
}

.online {
  position: absolute;
  right: 2rpx;
  bottom: 2rpx;
  width: 20rpx;
  height: 20rpx;
  border-radius: 50%;
  background: #00b578;
  border: 3rpx solid #fff;
}

.unread {
  position: absolute;
  top: -6rpx;
  right: -10rpx;
  min-width: 32rpx;
  height: 32rpx;
  line-height: 32rpx;
  padding: 0 8rpx;
  border-radius: 16rpx;
  background: #f53f3f;
  color: #fff;
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
  font-weight: 500;
}

.conv-time {
  font-size: 22rpx;
  color: #bbbfc4;
}

.conv-msg {
  font-size: 26rpx;
  color: #8a8f99;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
