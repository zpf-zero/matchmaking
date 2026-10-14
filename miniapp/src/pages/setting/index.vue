<script setup lang="ts">
import { authApi } from '@/api'
import { useUserStore } from '@/stores/user'

const store = useUserStore()

function goBlacklist() {
  uni.navigateTo({ url: '/pages/setting/blacklist' })
}

function maskPhone(p: string) {
  if (!p || p.length < 11) return p || '未绑定'
  return p.slice(0, 3) + '****' + p.slice(7)
}

function showAgreement(type: 'user' | 'privacy') {
  uni.navigateTo({ url: `/pages/common/agreement?type=${type}` })
}

function logout() {
  uni.showModal({
    title: '退出登录',
    content: '确定退出当前账号吗？',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await authApi.logout()
      } catch {
        // 退出接口失败不影响本地清理
      }
      store.logout()
    },
  })
}

function deleteAccount() {
  uni.showModal({
    title: '注销账号',
    content: '注销后资料、配对和聊天记录将无法恢复，确定继续吗？',
    confirmText: '确认注销',
    confirmColor: '#f53f3f',
    success: async (r) => {
      if (!r.confirm) return
      // 二次确认，避免误触这种不可逆操作
      uni.showModal({
        title: '再次确认',
        content: '这是不可逆操作，请再次确认',
        confirmText: '仍要注销',
        confirmColor: '#f53f3f',
        success: async (r2) => {
          if (!r2.confirm) return
          try {
            await authApi.deleteAccount()
            uni.showToast({ title: '账号已注销', icon: 'none' })
            setTimeout(() => store.logout(), 1000)
          } catch {
            // 已提示
          }
        },
      })
    },
  })
}
</script>

<template>
  <view class="page">
    <view class="group">
      <view class="item">
        <text class="label">手机号</text>
        <text class="value">{{ maskPhone(store.phone) }}</text>
      </view>
      <view class="item" @click="goBlacklist">
        <text class="label">黑名单</text>
        <view class="arrow"></view>
      </view>
    </view>

    <view class="group">
      <view class="item" @click="showAgreement('user')">
        <text class="label">用户协议</text>
        <view class="arrow"></view>
      </view>
      <view class="item" @click="showAgreement('privacy')">
        <text class="label">隐私政策</text>
        <view class="arrow"></view>
      </view>
    </view>

    <view class="group">
      <view class="item center" @click="logout">
        <text class="label-center">退出登录</text>
      </view>
    </view>

    <view class="group">
      <view class="item center" @click="deleteAccount">
        <text class="label-center danger">注销账号</text>
      </view>
    </view>

    <view class="tip">注销后账号信息将被清除，且无法恢复</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.group {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  margin-bottom: 24rpx;
  overflow: hidden;
}

.item {
  display: flex;
  align-items: center;
  min-height: 100rpx;
  padding: 0 24rpx;
  border-bottom: 1rpx solid $border;
  transition: background-color 180ms ease;

  &:active {
    background: $bg-subtle;
  }

  &:last-child {
    border-bottom: none;
  }

  &.center {
    justify-content: center;
  }
}

.label {
  flex: 1;
  font-size: $font-body;
}

.label-center {
  font-size: $font-body;
  color: $text-sub;

  &.danger {
    color: $danger;
  }
}

.value {
  font-size: $font-mid;
  color: $text-weak;
}

.arrow {
  width: 12rpx;
  height: 12rpx;
  flex-shrink: 0;
  margin-left: 16rpx;
  border-right: 2rpx solid $text-placeholder;
  border-bottom: 2rpx solid $text-placeholder;
  transform: rotate(-45deg);
  opacity: 0.6;
}

.tip {
  text-align: center;
  font-size: $font-micro;
  color: $text-placeholder;
  padding: 12rpx 40rpx;
  line-height: 1.6;
}
</style>
