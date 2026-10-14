<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { chatApi } from '@/api'
import type { Message } from '@/api/types'
import { useUserStore } from '@/stores/user'
import { uploadImage } from '@/utils/request'
import { parseTime } from '@/utils/time'
import { ws } from '@/utils/ws'

const store = useUserStore()
const convId = ref(0)
const peerId = ref(0)
const peerName = ref('')
const messages = ref<Message[]>([])
const input = ref('')
const scrollTop = ref(0)
const loadingMore = ref(false)
const noMore = ref(false)
const sending = ref(false)
const peerTyping = ref(false)
const ackTimers = new Map<string, any>()

const myId = () => store.user?.id || 0

/** 配对后不知道说什么是最大流失点，给几个现成的开场白 */
const icebreakers = [
  '你好，很高兴认识你',
  '看了你的自我介绍，感觉我们挺有共同点的',
  '周末一般喜欢做什么呀？',
  '你的照片拍得真不错，是在哪里拍的？',
]

onLoad(async (q) => {
  convId.value = Number(q?.id || 0)
  peerId.value = Number(q?.peer || 0)
  peerName.value = decodeURIComponent(q?.name || '聊天')
  uni.setNavigationBarTitle({ title: peerName.value })

  await loadHistory()
  await markRead()

  ws.on('message', onWsMessage)
  ws.on('ack', onWsAck)
  ws.on('error', onWsError)
  ws.on('typing', onWsTyping)
  ws.on('recall', onWsRecall)
  ws.connect()
})

onUnload(() => {
  ws.off('message', onWsMessage)
  ws.off('ack', onWsAck)
  ws.off('error', onWsError)
  ws.off('typing', onWsTyping)
  ws.off('recall', onWsRecall)
  clearTimeout(typingHideTimer)
  ackTimers.forEach((timer) => clearTimeout(timer))
  ackTimers.clear()
})

async function loadHistory(before = 0) {
  if (loadingMore.value || noMore.value) return
  loadingMore.value = true
  try {
    const list = await chatApi.messages(convId.value, before)
    if (!list.length) {
      noMore.value = true
      return
    }
    messages.value = before ? [...list, ...messages.value] : list
    if (!before) await scrollToBottom()
  } catch {
    // 已提示
  } finally {
    loadingMore.value = false
  }
}

function loadMore() {
  if (messages.value.length) loadHistory(messages.value[0].id)
}

async function markRead() {
  try {
    await chatApi.read(convId.value)
    store.fetchUnread()
  } catch {
    // 静默
  }
}

function onWsMessage(msg: Message) {
  if (msg.conversation_id !== convId.value) return
  peerTyping.value = false // 对方消息到了，输入提示立即消失
  messages.value.push(msg)
  scrollToBottom()
  markRead()
}

// ---------- 正在输入 ----------
let typingHideTimer: any = null
let lastTypingSent = 0

function onWsTyping(data: { conversation_id: number }) {
  if (data.conversation_id !== convId.value) return
  peerTyping.value = true
  clearTimeout(typingHideTimer)
  typingHideTimer = setTimeout(() => (peerTyping.value = false), 4000)
}

function onInput() {
  // 3 秒内只发一次，避免每敲一个字都发
  const now = Date.now()
  if (now - lastTypingSent < 3000 || !ws.connected) return
  lastTypingSent = now
  ws.send('typing', { conversation_id: convId.value })
}

// ---------- 撤回 ----------
function onWsRecall(data: { conversation_id: number; message_id: number }) {
  if (data.conversation_id !== convId.value) return
  const m = messages.value.find((x) => x.id === data.message_id)
  if (m) {
    m.type = 3
    m.content = '撤回了一条消息'
  }
}

function onLongPressMsg(m: Message) {
  if (!isMine(m) || m.type === 3 || !m.id) return
  // 2 分钟限制服务端会校验，这里只是提前隐藏入口减少无效点击
  const created = parseTime(m.created_at)
  if (Number.isNaN(created) || Date.now() - created > 2 * 60 * 1000) return

  uni.showActionSheet({
    itemList: ['撤回'],
    success: async () => {
      try {
        await chatApi.recall(m.id)
        m.type = 3
        m.content = '撤回了一条消息'
      } catch {
        // 已提示
      }
    },
  })
}

// ---------- 时间分隔线 ----------
/** 与上一条间隔超过 5 分钟才显示时间，模仿微信的做法，比每条都带时间干净得多 */
function showTimeDivider(index: number): boolean {
  if (index === 0) return true
  const cur = parseTime(messages.value[index].created_at)
  const prev = parseTime(messages.value[index - 1].created_at)
  return cur - prev > 5 * 60 * 1000
}

function dividerText(t: string) {
  const ts = parseTime(t)
  if (Number.isNaN(ts)) return ''
  const d = new Date(ts)
  const now = new Date()
  const hm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  if (d.toDateString() === now.toDateString()) return hm
  const yesterday = new Date(now.getTime() - 86400000)
  if (d.toDateString() === yesterday.toDateString()) return `昨天 ${hm}`
  return `${d.getMonth() + 1}月${d.getDate()}日 ${hm}`
}

function goPeerDetail() {
  if (peerId.value) uni.navigateTo({ url: `/pages/user/detail?id=${peerId.value}` })
}

function useIcebreaker(text: string) {
  input.value = text
}

function onWsAck(data: { temp_id: string; message: Message }) {
  clearAckTimer(data.temp_id)
  const idx = messages.value.findIndex((m) => m._temp_id === data.temp_id)
  if (idx >= 0) {
    messages.value[idx] = data.message
  } else {
    messages.value.push(data.message)
  }
  scrollToBottom()
}

function onWsError(data: { message: string; temp_id?: string }) {
  if (data.temp_id) {
    clearAckTimer(data.temp_id)
    const m = messages.value.find((x) => x._temp_id === data.temp_id)
    if (m) {
      m._sending = false
      m._failed = true
    }
  }
  uni.showToast({ title: data.message, icon: 'none' })
}

function clearAckTimer(tempId: string) {
  const timer = ackTimers.get(tempId)
  if (timer) clearTimeout(timer)
  ackTimers.delete(tempId)
}

function waitForAck(tempId: string) {
  clearAckTimer(tempId)
  ackTimers.set(
    tempId,
    setTimeout(() => {
      ackTimers.delete(tempId)
      const message = messages.value.find((item) => item._temp_id === tempId)
      if (!message || !message._sending) return
      message._sending = false
      message._failed = true
      uni.showToast({ title: '消息发送超时，请重试', icon: 'none' })
    }, 10000)
  )
}

function retryMessage(message: Message) {
  if (!message._failed) return
  const index = messages.value.indexOf(message)
  if (index >= 0) messages.value.splice(index, 1)
  doSend(message.content, message.type)
}

async function send() {
  const content = input.value.trim()
  if (!content || sending.value) return
  input.value = ''
  await doSend(content, 1)
}

function chooseImage() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    success: async (res) => {
      const path = (res.tempFilePaths as string[])[0]
      uni.showLoading({ title: '上传中' })
      try {
        const url = await uploadImage(path, 'chat')
        await doSend(url, 2)
      } catch {
        // 已提示
      } finally {
        uni.hideLoading()
      }
    },
  })
}

async function doSend(content: string, msgType: number) {
  if (sending.value) return
  sending.value = true
  const tempId = `t${Date.now()}`

  // 先本地上屏，网络慢时不至于点了没反应
  const optimistic: Message = {
    id: 0,
    conversation_id: convId.value,
    from_user_id: myId(),
    type: msgType,
    content,
    is_read: 0,
    created_at: new Date().toISOString(),
    _temp_id: tempId,
    _sending: true,
  }
  messages.value.push(optimistic)
  await scrollToBottom()

  // 优先走长连接，断线时退回 HTTP
  const ok =
    ws.connected &&
    ws.send('send', {
      conversation_id: convId.value,
      msg_type: msgType,
      content,
      temp_id: tempId,
    })

  if (ok) {
    waitForAck(tempId)
  } else {
    try {
      const msg = await chatApi.send(convId.value, content, msgType)
      const idx = messages.value.findIndex((m) => m._temp_id === tempId)
      if (idx >= 0) messages.value[idx] = msg
    } catch {
      const m = messages.value.find((x) => x._temp_id === tempId)
      if (m) {
        m._sending = false
        m._failed = true
      }
    }
  }

  sending.value = false
}

function previewImage(url: string) {
  uni.previewImage({ urls: [url] })
}

async function scrollToBottom() {
  await nextTick()
  scrollTop.value = 999999 + Math.random()
}

function isMine(m: Message) {
  return m.from_user_id === myId()
}

function unmatch() {
  uni.showModal({
    title: '解除配对',
    content: '解除后将无法继续聊天，且需要重新互相喜欢才能配对',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await chatApi.closeConversation(convId.value)
        uni.showToast({ title: '已解除配对', icon: 'none' })
        setTimeout(() => uni.navigateBack(), 800)
      } catch {
        // 已提示
      }
    },
  })
}

</script>

<template>
  <view class="chat">
    <!-- 对方信息栏：正在输入提示 + 资料入口 -->
    <view class="peer-bar" @click="goPeerDetail">
      <view class="peer-left">
        <text class="peer-name">{{ peerName }}</text>
        <text v-if="peerTyping" class="typing">正在输入…</text>
      </view>
      <view v-if="peerId" class="peer-link">
        <text>查看资料</text>
        <view class="chev"></view>
      </view>
    </view>

    <scroll-view
      class="msg-area"
      scroll-y
      :scroll-top="scrollTop"
      :scroll-with-animation="false"
      @scrolltoupper="loadMore"
      upper-threshold="60"
    >
      <view class="msg-list">
        <view v-if="noMore && messages.length" class="history-tip">没有更多消息了</view>

        <view class="safe-notice">
          请勿在聊天中透露银行卡、验证码等信息。凡涉及借钱、投资、刷单，均为诈骗。
        </view>

        <!-- 空会话：破冰话题 -->
        <view v-if="noMore && !messages.length" class="icebreak">
          <view class="ib-title">配对成功，不知道说什么？可以从这些话题开始</view>
          <view
            v-for="(t, i) in icebreakers"
            :key="i"
            class="ib-item"
            @click="useIcebreaker(t)"
          >
            {{ t }}
          </view>
        </view>

        <template v-for="(m, i) in messages" :key="m.id || m._temp_id">
          <!-- 时间分隔线：与上一条超过 5 分钟才显示 -->
          <view v-if="showTimeDivider(i)" class="time-divider">{{ dividerText(m.created_at) }}</view>

          <!-- 系统消息（撤回等） -->
          <view v-if="m.type === 3" class="sys-msg">
            {{ isMine(m) ? '你' : '对方' }}{{ m.content }}
          </view>

          <view v-else class="msg" :class="{ mine: isMine(m) }">
            <image
              v-if="m.type === 2"
              :src="m.content"
              class="img-msg"
              mode="widthFix"
              @click="previewImage(m.content)"
              @longpress="onLongPressMsg(m)"
            />
            <view v-else class="bubble" @longpress="onLongPressMsg(m)">
              <text>{{ m.content }}</text>
            </view>
            <view v-if="m._sending || m._failed" class="msg-meta">
              <text v-if="m._sending" class="sending">发送中</text>
              <text v-else class="failed" @click="retryMessage(m)">发送失败，点击重试</text>
            </view>
          </view>
        </template>
      </view>
    </scroll-view>

    <view class="input-bar">
      <input
        v-model="input"
        class="input"
        placeholder="说点什么…"
        placeholder-class="ph"
        confirm-type="send"
        :adjust-position="true"
        @input="onInput"
        @confirm="send"
      />
      <view class="img-btn" @click="chooseImage">图片</view>
      <view class="send-btn" :class="{ active: input.trim() }" @click="send">发送</view>
    </view>

    <view class="more-bar">
      <text @click="unmatch">解除配对</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.chat {
  display: flex;
  flex-direction: column;
  height: 100dvh;
  background: $bg-page;
}

.peer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18rpx 24rpx;
  background: $bg-card;
  border-bottom: 1rpx solid $border;
}

.peer-left {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
}

.peer-name {
  font-size: $font-body;
  font-weight: 600;
  color: $text-main;
}

.typing {
  font-size: $font-micro;
  color: $success;
}

.peer-link {
  display: flex;
  align-items: center;
  font-size: $font-caption;
  color: $text-weak;
}

.chev {
  width: 10rpx;
  height: 10rpx;
  margin-left: 10rpx;
  border-right: 2rpx solid $text-placeholder;
  border-bottom: 2rpx solid $text-placeholder;
  transform: rotate(-45deg);
  opacity: 0.7;
}

.icebreak {
  padding: 20rpx 0;
}

.ib-title {
  text-align: center;
  font-size: $font-caption;
  color: $text-weak;
  margin-bottom: 20rpx;
  line-height: 1.6;
}

.ib-item {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius;
  padding: 20rpx 24rpx;
  margin-bottom: 14rpx;
  font-size: $font-mid;
  color: $text-sub;
  transition: transform 160ms ease, background-color 160ms ease;

  &:active {
    background: $brand-light;
    transform: scale(0.99);
  }
}

.time-divider {
  text-align: center;
  font-size: $font-micro;
  color: $text-placeholder;
  padding: 14rpx 0;
}

.sys-msg {
  text-align: center;
  font-size: $font-micro;
  color: $text-placeholder;
  padding: 10rpx 0;
  margin-bottom: 16rpx;
}

.msg-area {
  flex: 1;
  overflow: hidden;
}

.msg-list {
  padding: 24rpx;
}

.history-tip,
.safe-notice {
  text-align: center;
  font-size: $font-micro;
  color: $text-placeholder;
  padding: 16rpx 40rpx;
  line-height: 1.6;
}

.safe-notice {
  background: $warning-light;
  color: $warning;
  border-left: 4rpx solid $warning;
  border-radius: $radius-sm;
  margin-bottom: 24rpx;
}

.msg {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin-bottom: 28rpx;

  &.mine {
    align-items: flex-end;
  }
}

.bubble {
  max-width: 76%;
  padding: 20rpx 24rpx;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  background: $bg-card;
  font-size: $font-body;
  line-height: 1.5;
  color: $text-main;
  word-break: break-all;
}

.mine .bubble {
  background: $brand;
  border-color: $brand;
  color: $text-on-accent;
}

.img-msg {
  max-width: 400rpx;
  border-radius: $radius;
  background: $bg-card;
}

.msg-meta {
  margin-top: 8rpx;
  font-size: $font-micro;
  color: $text-placeholder;
}

.failed {
  color: $danger;
}

.input-bar {
  display: flex;
  align-items: center;
  padding: 16rpx 24rpx;
  background: $bg-card;
  border-top: 1rpx solid $border;
}

.input {
  flex: 1;
  height: 76rpx;
  background: $bg-subtle;
  border: 1rpx solid transparent;
  border-radius: $radius;
  padding: 0 22rpx;
  font-size: $font-body;

  &:focus {
    border-color: $brand;
    background: $bg-card;
  }
}

.ph {
  color: $text-placeholder;
}

.img-btn {
  margin-left: 16rpx;
  width: 72rpx;
  height: 72rpx;
  line-height: 72rpx;
  text-align: center;
  color: $text-sub;
  font-size: $font-micro;
  font-weight: 500;
}

.send-btn {
  margin-left: 12rpx;
  padding: 0 32rpx;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: $radius;
  background: $bg-subtle;
  color: $text-placeholder;
  font-size: $font-body;
  font-weight: 600;
  transition: transform 160ms ease, background-color 160ms ease;

  &:active {
    transform: scale(0.97);
  }

  &.active {
    background: $brand;
    color: $text-on-accent;
  }
}

.more-bar {
  background: $bg-card;
  text-align: center;
  font-size: $font-micro;
  color: $text-placeholder;
  padding: 12rpx 0 calc(12rpx + env(safe-area-inset-bottom));
}
</style>
