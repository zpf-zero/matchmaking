<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import UserAvatar from '@/components/UserAvatar.vue'
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
        <UserAvatar :src="b.avatar" :name="b.nickname" :size="88" />
        <text class="name">{{ b.nickname }}</text>
        <view class="btn" @click="remove(b.user_id, b.nickname)">移出</view>
      </view>
    </view>
    <view v-else-if="loaded" class="empty">黑名单是空的</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.list {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  overflow: hidden;
}

.item {
  display: flex;
  align-items: center;
  padding: 24rpx;
  border-bottom: 1rpx solid $border;

  &:last-child {
    border-bottom: none;
  }
}

.name {
  flex: 1;
  margin-left: 24rpx;
  font-size: $font-body;
}

.btn {
  padding: 10rpx 28rpx;
  border-radius: $radius;
  border: 1rpx solid $border;
  color: $text-sub;
  font-size: $font-mid;
  transition: transform 160ms ease, background-color 160ms ease;

  &:active {
    transform: scale(0.97);
    background: $bg-subtle;
  }
}
</style>
