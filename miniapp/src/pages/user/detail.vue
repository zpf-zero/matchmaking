<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { commonApi, interactionApi, userApi } from '@/api'
import type { UserDetail } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const detail = ref<UserDetail | null>(null)
const loading = ref(true)
const userId = ref(0)

onLoad(async (q) => {
  userId.value = Number(q?.id || 0)
  if (!userId.value) {
    uni.showToast({ title: '参数错误', icon: 'none' })
    return
  }
  await store.fetchDict()
  await load()
})

async function load() {
  loading.value = true
  try {
    detail.value = await userApi.detail(userId.value)
  } catch {
    setTimeout(() => uni.navigateBack(), 1200)
  } finally {
    loading.value = false
  }
}

function previewPhoto(index: number) {
  if (!detail.value?.photos?.length) return
  uni.previewImage({ urls: detail.value.photos, current: index })
}

async function handleLike() {
  if (!store.canInteract) {
    uni.showToast({ title: '资料审核通过后才能表达心意', icon: 'none' })
    return
  }
  const d = detail.value!

  try {
    const res = await interactionApi.like(d.id)
    d.is_liked = true

    if (res.is_match) {
      d.is_matched = true
      d.conversation_id = res.conversation_id
      uni.showModal({
        title: '配对成功',
        content: `你和 ${d.nickname} 互相喜欢，现在可以聊天了`,
        confirmText: '去聊天',
        cancelText: '稍后',
        success: (r) => r.confirm && goChat(),
      })
    } else {
      uni.showToast({ title: '已表达心意，等待对方回应', icon: 'none' })
    }
  } catch {
    // 已提示
  }
}

async function handlePass() {
  try {
    await interactionApi.pass(detail.value!.id)
    uni.showToast({ title: '已跳过', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 600)
  } catch {
    // 已提示
  }
}

function goChat() {
  const d = detail.value!
  uni.navigateTo({
    url: `/pages/chat/index?id=${d.conversation_id}&name=${encodeURIComponent(d.nickname)}`,
  })
}

function moreActions() {
  uni.showActionSheet({
    itemList: ['举报该用户', '拉黑该用户'],
    success: (res) => {
      if (res.tapIndex === 0) doReport()
      else doBlock()
    },
  })
}

function doReport() {
  const reasons = store.dict['report_reason'] || []
  uni.showActionSheet({
    itemList: reasons.map((r) => r.label),
    success: async (res) => {
      try {
        await commonApi.report({
          target_user_id: detail.value!.id,
          reason: reasons[res.tapIndex].value,
        })
        uni.showToast({ title: '举报已提交，我们会尽快核实', icon: 'none' })
      } catch {
        // 已提示
      }
    },
  })
}

function doBlock() {
  uni.showModal({
    title: '确认拉黑',
    content: '拉黑后双方将互相不可见，已有的配对也会解除',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await commonApi.addBlacklist(detail.value!.id)
        uni.showToast({ title: '已拉黑', icon: 'none' })
        setTimeout(() => uni.navigateBack(), 800)
      } catch {
        // 已提示
      }
    },
  })
}
</script>

<template>
  <view v-if="detail" class="page">
    <!-- 头图 -->
    <view class="hero">
      <image :src="detail.avatar" class="hero-img" mode="aspectFill" @click="previewPhoto(0)" />
      <view class="hero-mask"></view>
      <view class="hero-info">
        <view class="hero-name">{{ detail.nickname }}</view>
        <view class="hero-meta">{{ detail.age }}岁 · {{ detail.height }}cm · {{ detail.city }}</view>
      </view>
      <view class="more" @click="moreActions">···</view>
      <view v-if="detail.match_rate > 0" class="match-badge">匹配度 {{ detail.match_rate }}%</view>
    </view>

    <!-- 关系提示 -->
    <view v-if="detail.liked_me && !detail.is_matched" class="hint">
      TA 已经喜欢你了，回应一下试试？
    </view>

    <!-- 核心条件 -->
    <view class="card">
      <view class="card-title">基本条件</view>
      <view class="grid">
        <view class="grid-item">
          <text class="k">学历</text><text class="v">{{ detail.education_text }}</text>
        </view>
        <view class="grid-item">
          <text class="k">月收入</text><text class="v">{{ detail.income_text }}</text>
        </view>
        <view class="grid-item">
          <text class="k">婚姻状况</text><text class="v">{{ detail.marital_text }}</text>
        </view>
        <view class="grid-item">
          <text class="k">子女</text>
          <text class="v">{{ store.label('has_children', detail.has_children) }}</text>
        </view>
        <view class="grid-item">
          <text class="k">购房</text>
          <text class="v">{{ store.label('house_status', detail.house_status) }}</text>
        </view>
        <view class="grid-item">
          <text class="k">购车</text>
          <text class="v">{{ store.label('car_status', detail.car_status) }}</text>
        </view>
        <view class="grid-item">
          <text class="k">职业</text><text class="v">{{ detail.occupation || '—' }}</text>
        </view>
        <view class="grid-item">
          <text class="k">院校</text><text class="v">{{ detail.school || '—' }}</text>
        </view>
        <view class="grid-item">
          <text class="k">现居</text>
          <text class="v">{{ detail.province }} {{ detail.city }}</text>
        </view>
        <view class="grid-item">
          <text class="k">户籍</text>
          <text class="v">{{ detail.home_province }} {{ detail.home_city }}</text>
        </view>
      </view>
    </view>

    <!-- 自我介绍 -->
    <view class="card">
      <view class="card-title">自我介绍</view>
      <view class="intro">{{ detail.self_intro }}</view>
    </view>

    <!-- 相册 -->
    <view v-if="detail.photos?.length" class="card">
      <view class="card-title">相册</view>
      <view class="photos">
        <image
          v-for="(p, i) in detail.photos"
          :key="i"
          :src="p"
          class="photo"
          mode="aspectFill"
          @click="previewPhoto(i)"
        />
      </view>
    </view>

    <view class="safe-tip">
      平台不会以任何形式索要钱财。遇到索要红包、诱导投资等情况请立即举报。
    </view>

    <!-- 底部操作 -->
    <view class="actions">
      <template v-if="detail.is_matched">
        <view class="btn-primary full" @click="goChat">去聊天</view>
      </template>
      <template v-else>
        <view class="btn-plain half" @click="handlePass">跳过</view>
        <view class="btn-primary half" :class="{ disabled: detail.is_liked }" @click="handleLike">
          {{ detail.is_liked ? '已喜欢' : '喜欢 TA' }}
        </view>
      </template>
    </view>
  </view>

  <view v-else-if="!loading" class="empty">资料加载失败</view>
</template>

<style lang="scss" scoped>
.page {
  padding-bottom: 180rpx;
}

.hero {
  position: relative;
  height: 620rpx;
}

.hero-img {
  width: 100%;
  height: 100%;
  background: #ebedf0;
}

.hero-mask {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 220rpx;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.55) 100%);
}

.hero-info {
  position: absolute;
  left: 32rpx;
  bottom: 32rpx;
  color: #fff;
}

.hero-name {
  font-size: 42rpx;
  font-weight: 600;
  margin-bottom: 10rpx;
}

.hero-meta {
  font-size: 26rpx;
  opacity: 0.92;
}

.more {
  position: absolute;
  right: 24rpx;
  top: 24rpx;
  width: 64rpx;
  height: 64rpx;
  line-height: 52rpx;
  text-align: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.3);
  color: #fff;
  font-size: 34rpx;
}

.match-badge {
  position: absolute;
  right: 24rpx;
  bottom: 32rpx;
  padding: 8rpx 20rpx;
  border-radius: 28rpx;
  background: rgba(232, 80, 109, 0.92);
  color: #fff;
  font-size: 24rpx;
}

.hint {
  margin: 20rpx 24rpx 0;
  padding: 20rpx 24rpx;
  background: #fff0f3;
  color: #e8506d;
  border-radius: 12rpx;
  font-size: 26rpx;
}

.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 28rpx 24rpx;
  margin: 20rpx 24rpx 0;
}

.card-title {
  font-size: 30rpx;
  font-weight: 600;
  margin-bottom: 20rpx;
}

.grid {
  display: flex;
  flex-wrap: wrap;
}

.grid-item {
  width: 50%;
  display: flex;
  padding: 14rpx 0;
  font-size: 26rpx;

  .k {
    color: #8a8f99;
    width: 130rpx;
    flex-shrink: 0;
  }

  .v {
    color: #1f2329;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.intro {
  font-size: 28rpx;
  line-height: 1.8;
  color: #4e5969;
  white-space: pre-wrap;
}

.photos {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.photo {
  width: 210rpx;
  height: 210rpx;
  border-radius: 12rpx;
  background: #f5f6f8;
}

.safe-tip {
  margin: 24rpx;
  font-size: 22rpx;
  color: #bbbfc4;
  line-height: 1.6;
  text-align: center;
}

.actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  display: flex;
  gap: 20rpx;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.half {
  flex: 1;
}

.full {
  flex: 1;
}
</style>
