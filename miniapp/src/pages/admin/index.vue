<script setup lang="ts">
import { computed, ref } from 'vue'
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

const roleText = computed(
  () => ['', '普通用户', '审核员', '管理员', '超级管理员'][store.role] || '未知身份'
)
const pendingTotal = computed(() => stat.value.pending_audit + stat.value.pending_report)

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

/** 用户端是 tabBar 页，必须走 switchTab；后台工作台是普通页，切走后再回来靠 reLaunch */
function goUserSide() {
  uni.switchTab({ url: '/pages/index/index' })
}
</script>

<template>
  <view class="page">
    <view class="role-row">
      <text class="role-chip">当前身份 · {{ roleText }}</text>
      <text v-if="pendingTotal" class="role-alert">{{ pendingTotal }} 项待处理</text>
    </view>

    <view class="sec-head">
      <text class="sec-title">待处理</text>
    </view>
    <view class="queue">
      <view
        class="q-item"
        :class="{ hot: stat.pending_audit > 0 }"
        @click="go('/pages/admin/audit')"
      >
        <text class="q-num">{{ stat.pending_audit }}</text>
        <text class="q-lab">资料审核</text>
        <view class="arrow"></view>
      </view>
      <view
        class="q-item"
        :class="{ hot: stat.pending_report > 0 }"
        @click="go('/pages/admin/report')"
      >
        <text class="q-num">{{ stat.pending_report }}</text>
        <text class="q-lab">举报处理</text>
        <view class="arrow"></view>
      </view>
    </view>

    <view class="sec-head">
      <text class="sec-title">数据概览</text>
    </view>
    <view class="panel">
      <view class="cell">
        <text class="num">{{ stat.total_user }}</text>
        <text class="lab">注册用户</text>
      </view>
      <view class="cell">
        <text class="num">{{ stat.today_new }}</text>
        <text class="lab">今日新增</text>
      </view>
      <view class="cell">
        <text class="num">{{ stat.active_week }}</text>
        <text class="lab">7 日活跃</text>
      </view>
      <view class="cell">
        <text class="num">{{ stat.matched }}</text>
        <text class="lab">配对总数</text>
      </view>
    </view>

    <view class="sec-head">
      <text class="sec-title">管理</text>
    </view>
    <view class="menu">
      <view class="item" @click="go('/pages/admin/user')">
        <text class="label">用户管理</text>
        <text v-if="store.role < Role.Admin" class="hint">仅查看</text>
        <view class="arrow"></view>
      </view>
      <!-- 管理员登录后直接落到本页、不在 tabBar 里，给一条回用户端的路 -->
      <view class="item" @click="goUserSide">
        <text class="label">进入用户端</text>
        <text class="hint">以管理员身份浏览</text>
        <view class="arrow"></view>
      </view>
    </view>

    <view class="note">
      角色权限由后端逐接口校验。审核员可审核资料与处理举报，封禁用户需要管理员权限，
      变更角色需要超级管理员权限。
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.role-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 8rpx;
}

.role-chip {
  padding: 8rpx 22rpx;
  border-radius: $radius-round;
  background: $bg-subtle;
  color: $text-sub;
  font-size: $font-caption;
}

.role-alert {
  color: $brand;
  font-size: $font-caption;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.sec-head {
  margin: 40rpx 0 18rpx;
  padding: 0 8rpx;
}

.sec-title {
  font-size: $font-mid;
  font-weight: 650;
  color: $text-main;
  letter-spacing: 0.5rpx;
}

/* 待处理是这一页的主角：数字最大，有积压时整块点亮 */
.queue {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.q-item {
  position: relative;
  padding: 30rpx 26rpx 26rpx;
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  transition: transform 180ms cubic-bezier(0.16, 1, 0.3, 1), background-color 200ms ease,
    border-color 200ms ease;

  &:active {
    transform: scale(0.98);
  }

  &.hot {
    background: $bg-accent;
    border-color: transparent;
  }

  .arrow {
    position: absolute;
    top: 30rpx;
    right: 24rpx;
    margin-left: 0;
  }
}

.q-num {
  display: block;
  font-size: $font-title;
  font-weight: 650;
  letter-spacing: -1rpx;
  color: $text-main;
  font-variant-numeric: tabular-nums;
}

.hot .q-num {
  color: $brand;
}

.q-lab {
  display: block;
  margin-top: 6rpx;
  font-size: $font-caption;
  color: $text-weak;
}

/* 概览数据合并成一块面板，避免四张卡片互相抢注意力 */
.panel {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  overflow: hidden;
}

.cell {
  padding: 28rpx 26rpx;
  border-right: 1rpx solid $border;
  border-bottom: 1rpx solid $border;

  &:nth-child(2n) {
    border-right: none;
  }

  &:nth-child(n + 3) {
    border-bottom: none;
  }
}

.num {
  display: block;
  font-size: $font-section;
  font-weight: 600;
  color: $text-main;
  font-variant-numeric: tabular-nums;
  margin-bottom: 8rpx;
}

.lab {
  font-size: $font-micro;
  color: $text-weak;
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
  transition: background-color 180ms ease;

  &:active {
    background: $bg-subtle;
  }
}

.label {
  flex: 1;
  font-size: $font-body;
}

.hint {
  font-size: $font-micro;
  color: $text-placeholder;
  margin-right: 10rpx;
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

.note {
  font-size: $font-caption;
  color: $text-placeholder;
  line-height: 1.7;
  padding: 36rpx 8rpx 0;
}
</style>
