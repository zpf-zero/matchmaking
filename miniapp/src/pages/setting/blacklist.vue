<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { commonApi } from '@/api'

const list = ref<{ user_id: number; nickname: string; avatar: string }[]>([])
const loaded = ref(false)

onShow(load)

async function load() {
  try {
    list.value = await commonApi.blacklist()
  } catch {
    // 已提示
  } finally {
    loaded.value = true
  }
}

function remove(userId: number, nickname: string) {
  uni.showModal({
    title: '移出黑名单',
    content: `将 ${nickname} 移出黑名单后，你们可以重新看到对方`,
    success: async (r) => {
      if (!r.confirm) return
      try {
        await commonApi.removeBlacklist(userId)
        await load()
        uni.showToast({ title: '已移出', icon: 'none' })
      } catch {
        // 已提示
      }
    },
  })
}
</script>

<template>
  <view class="page">
    <view v-if="list.length" class="list">
      <view v-for="b in list" :key="b.user_id" class="item">
        <image :src="b.avatar" class="avatar" mode="aspectFill" />
        <text class="name">{{ b.nickname }}</text>
        <view class="btn" @click="remove(b.user_id, b.nickname)">移出</view>
      </view>
    </view>
    <view v-else-if="loaded" class="empty">黑名单是空的</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 24rpx;
}

.list {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.item {
  display: flex;
  align-items: center;
  padding: 24rpx;
  border-bottom: 1rpx solid #f2f3f5;

  &:last-child {
    border-bottom: none;
  }
}

.avatar {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  background: #f5f6f8;
}

.name {
  flex: 1;
  margin-left: 24rpx;
  font-size: 29rpx;
}

.btn {
  padding: 10rpx 28rpx;
  border-radius: 28rpx;
  border: 1rpx solid #ebedf0;
  color: #646a73;
  font-size: 25rpx;
}
</style>
