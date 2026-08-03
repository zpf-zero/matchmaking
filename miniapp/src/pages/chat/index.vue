<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { chatApi } from '@/api'
import type { Message } from '@/api/types'
import { useUserStore } from '@/stores/user'
import { uploadImage } from '@/utils/request'
import { ws } from '@/utils/ws'

const store = useUserStore()
const convId = ref(0)
const peerName = ref('')
const messages = ref<Message[]>([])
const input = ref('')
const scrollTop = ref(0)
const loadingMore = ref(false)
const noMore = ref(false)
const sending = ref(false)

const myId = () => store.user?.id || 0

onLoad(async (q) => {
  convId.value = Number(q?.id || 0)
  peerName.value = decodeURIComponent(q?.name || '聊天')
  uni.setNavigationBarTitle({ title: peerName.value })

  await loadHistory()
  await markRead()

  ws.on('message', onWsMessage)
  ws.on('ack', onWsAck)
  ws.on('error', onWsError)
  ws.connect()
})

onUnload(() => {
  ws.off('message', onWsMessage)
  ws.off('ack', onWsAck)
  ws.off('error', onWsError)
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
  messages.value.push(msg)
  scrollToBottom()
  markRead()
}

function onWsAck(data: { temp_id: string; message: Message }) {
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
    const m = messages.value.find((x) => x._temp_id === data.temp_id)
    if (m) {
      m._sending = false
      m._failed = true
    }
  }
  uni.showToast({ title: data.message, icon: 'none' })
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

  if (!ok) {
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

function timeText(t: string) {
  const d = new Date(t.replace(/-/g, '/').replace(/\..*$/, ''))
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}
</script>

<template>
  <view class="chat">
    <scroll-view
      class="msg-area"
      scroll-y
      :scroll-top="scrollTop"
      :scroll-with-animation="false"
      @scrolltoupper="loadMore"
      upper-threshold="60"
    >
      <view class="msg-list">
        <view v-if="noMore" class="history-tip">没有更多消息了</view>

        <view class="safe-notice">
          请勿在聊天中透露银行卡、验证码等信息。凡涉及借钱、投资、刷单，均为诈骗。
        </view>

        <view v-for="m in messages" :key="m.id || m._temp_id" class="msg" :class="{ mine: isMine(m) }">
          <image
            v-if="m.type === 2"
            :src="m.content"
            class="img-msg"
            mode="widthFix"
            @click="previewImage(m.content)"
          />
          <view v-else class="bubble">
            <text>{{ m.content }}</text>
          </view>
          <view class="msg-meta">
            <text v-if="m._sending" class="sending">发送中</text>
            <text v-else-if="m._failed" class="failed">发送失败</text>
            <text v-else class="time">{{ timeText(m.created_at) }}</text>
          </view>
        </view>
      </view>
    </scroll-view>

    <view class="input-bar">
      <input
        v-model="input"
        class="input"
        placeholder="说点什么..."
        placeholder-class="ph"
        confirm-type="send"
        :adjust-position="true"
        @confirm="send"
      />
      <view class="img-btn" @click="chooseImage">🖼</view>
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
  height: 100vh;
  background: #f5f6f8;
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
  font-size: 22rpx;
  color: #bbbfc4;
  padding: 16rpx 40rpx;
  line-height: 1.6;
}

.safe-notice {
  background: #fff7e8;
  color: #ff8f1f;
  border-radius: 12rpx;
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
  border-radius: 16rpx;
  background: #fff;
  font-size: 29rpx;
  line-height: 1.5;
  color: #1f2329;
  word-break: break-all;
}

.mine .bubble {
  background: #e8506d;
  color: #fff;
}

.img-msg {
  max-width: 400rpx;
  border-radius: 16rpx;
  background: #fff;
}

.msg-meta {
  margin-top: 8rpx;
  font-size: 20rpx;
  color: #bbbfc4;
}

.failed {
  color: #f53f3f;
}

.input-bar {
  display: flex;
  align-items: center;
  padding: 16rpx 24rpx;
  background: #fff;
  border-top: 1rpx solid #ebedf0;
}

.input {
  flex: 1;
  height: 76rpx;
  background: #f5f6f8;
  border-radius: 38rpx;
  padding: 0 28rpx;
  font-size: 28rpx;
}

.ph {
  color: #bbbfc4;
}

.img-btn {
  margin-left: 16rpx;
  width: 72rpx;
  height: 72rpx;
  line-height: 72rpx;
  text-align: center;
  font-size: 34rpx;
}

.send-btn {
  margin-left: 12rpx;
  padding: 0 32rpx;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: 38rpx;
  background: #ebedf0;
  color: #bbbfc4;
  font-size: 28rpx;

  &.active {
    background: #e8506d;
    color: #fff;
  }
}

.more-bar {
  background: #fff;
  text-align: center;
  font-size: 22rpx;
  color: #bbbfc4;
  padding: 12rpx 0 calc(12rpx + env(safe-area-inset-bottom));
}
</style>
