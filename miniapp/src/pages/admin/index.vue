<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { adminApi } from '@/api'
import { Role, useUserStore } from '@/stores/user'

const store = useUserStore()
const stat = ref({
  total_user: 0,
  today_new: 0,
  pending_audit: 0,
  pending_report: 0,
  matched: 0,
  active_week: 0,
})

onShow(async () => {
  try {
    stat.value = await adminApi.overview()
  } catch {
    // 无权限时后端返回 403，request 层已提示
  }
})

function go(url: string) {
  uni.navigateTo({ url })
}
</script>

<template>
  <view class="page">
    <view class="role-bar">
      当前身份：{{ ['', '普通用户', '审核员', '管理员', '超级管理员'][store.role] }}
    </view>

    <view class="stat-grid">
      <view class="stat-card">
        <text class="num">{{ stat.total_user }}</text>
        <text class="lab">注册用户</text>
      </view>
      <view class="stat-card">
        <text class="num">{{ stat.today_new }}</text>
        <text class="lab">今日新增</text>
      </view>
      <view class="stat-card">
        <text class="num">{{ stat.active_week }}</text>
        <text class="lab">7日活跃</text>
      </view>
      <view class="stat-card">
        <text class="num">{{ stat.matched }}</text>
        <text class="lab">配对总数</text>
      </view>
    </view>

    <view class="menu">
      <view class="item" @click="go('/pages/admin/audit')">
        <text class="label">资料审核</text>
        <view v-if="stat.pending_audit" class="badge">{{ stat.pending_audit }}</view>
        <text class="arrow">›</text>
      </view>

      <view class="item" @click="go('/pages/admin/report')">
        <text class="label">举报处理</text>
        <view v-if="stat.pending_report" class="badge">{{ stat.pending_report }}</view>
        <text class="arrow">›</text>
      </view>

      <view class="item" @click="go('/pages/admin/user')">
        <text class="label">用户管理</text>
        <text v-if="store.role < Role.Admin" class="hint">仅查看</text>
        <text class="arrow">›</text>
      </view>
    </view>

    <view class="note">
      角色权限由后端逐接口校验。审核员可审核资料与处理举报，封禁用户需要管理员权限。
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.role-bar {
  font-size: 24rpx;
  color: $text-weak;
  padding: 0 8rpx 20rpx;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
  margin-bottom: 24rpx;
}

.stat-card {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 32rpx 24rpx;

  .num {
    display: block;
    font-size: 44rpx;
    font-weight: 600;
    color: $text-main;
    font-variant-numeric: tabular-nums;
    margin-bottom: 10rpx;
  }

  .lab {
    font-size: 24rpx;
    color: $text-weak;
  }
}

.menu {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  overflow: hidden;
}

.item {
  display: flex;
  align-items: center;
  min-height: 104rpx;
  padding: 0 24rpx;
  border-bottom: 1rpx solid $border;
  transition: background-color 180ms ease;

  &:active {
    background: $bg-subtle;
  }

  &:last-child {
    border-bottom: none;
  }
}

.label {
  flex: 1;
  font-size: 29rpx;
}

.hint {
  font-size: 22rpx;
  color: $text-placeholder;
  margin-right: 10rpx;
}

.badge {
  min-width: 36rpx;
  height: 36rpx;
  line-height: 36rpx;
  padding: 0 10rpx;
  border-radius: $radius-sm;
  background: $danger;
  color: $text-on-accent;
  font-size: 22rpx;
  text-align: center;
}

.arrow {
  color: $text-placeholder;
  font-size: 36rpx;
  margin-left: 10rpx;
}

.note {
  font-size: 22rpx;
  color: $text-placeholder;
  line-height: 1.6;
  padding: 32rpx 8rpx;
}
</style>
