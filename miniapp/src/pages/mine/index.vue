<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { adminApi } from '@/api'
import { Status, useUserStore } from '@/stores/user'

const store = useUserStore()
const todo = ref(0)

onShow(async () => {
  if (!store.isLogin) return
  try {
    await store.fetchMe()
    store.fetchUnread()
    // 管理员额外拉一下待办数，直接显示在入口上
    if (store.isAdminRole) {
      const ov = await adminApi.overview()
      todo.value = ov.pending_audit + ov.pending_report
    }
  } catch {
    // 已处理
  }
})

const statusText = computed(() => {
  switch (store.status) {
    case Status.Incomplete:
      return { text: '资料未完善', cls: 'warn' }
    case Status.Auditing:
      return { text: '审核中', cls: 'warn' }
    case Status.Normal:
      return { text: '已通过审核', cls: 'ok' }
    case Status.Rejected:
      return { text: '审核未通过', cls: 'danger' }
    case Status.Banned:
      return { text: '账号已封禁', cls: 'danger' }
    default:
      return { text: '', cls: '' }
  }
})

const roleText = computed(() => {
  return ['', '', '审核员', '管理员', '超级管理员'][store.role] || ''
})

function go(url: string) {
  uni.navigateTo({ url })
}
</script>

<template>
  <view class="page">
    <!-- 资料卡 -->
    <view class="profile-card" @click="go('/pages/profile/setup?mode=edit')">
      <image :src="store.user?.avatar" class="avatar" mode="aspectFill" />
      <view class="info">
        <view class="name-row">
          <text class="name">{{ store.user?.nickname || '未填写' }}</text>
          <text v-if="roleText" class="role-tag">{{ roleText }}</text>
        </view>
        <view class="status" :class="statusText.cls">{{ statusText.text }}</view>
        <view v-if="store.profile" class="complete">
          资料完整度 {{ store.profile.complete_rate }}%
        </view>
      </view>
      <text class="arrow">›</text>
    </view>

    <view v-if="store.status === Status.Rejected" class="reject-tip">
      未通过原因：{{ store.profile?.audit_remark || '资料不符合规范' }}
    </view>

    <!-- 数据 -->
    <view class="stat-row">
      <view class="stat" @click="go('/pages/interaction/list?tab=0')">
        <text class="num">{{ store.unread.liked_me }}</text>
        <text class="lab">新的喜欢</text>
      </view>
      <view class="stat" @click="go('/pages/interaction/list?tab=1')">
        <text class="num">—</text>
        <text class="lab">我喜欢的</text>
      </view>
      <view class="stat" @click="go('/pages/interaction/list?tab=2')">
        <text class="num">{{ store.unread.viewed_me }}</text>
        <text class="lab">谁看过我</text>
      </view>
    </view>

    <!-- 菜单 -->
    <view class="menu">
      <view class="menu-item" @click="go('/pages/profile/setup?mode=edit')">
        <text class="mi-label">我的资料</text><text class="arrow">›</text>
      </view>
      <view class="menu-item" @click="go('/pages/user/requirement')">
        <text class="mi-label">择偶要求</text><text class="arrow">›</text>
      </view>
      <view class="menu-item" @click="go('/pages/user/photos')">
        <text class="mi-label">我的相册</text>
        <text class="mi-extra">{{ store.photos.length }}/9</text>
        <text class="arrow">›</text>
      </view>
      <view class="menu-item" @click="go('/pages/setting/index')">
        <text class="mi-label">设置</text><text class="arrow">›</text>
      </view>
    </view>

    <!-- 管理入口：role > 1 才渲染。真正的拦截在后端 RequireRole -->
    <view v-if="store.isAdminRole" class="menu admin-menu">
      <view class="menu-title">管理功能</view>
      <view class="menu-item" @click="go('/pages/admin/index')">
        <text class="mi-label">⚙️ 管理中心</text>
        <view v-if="todo" class="todo-badge">{{ todo }}</view>
        <text class="arrow">›</text>
      </view>
    </view>

    <view class="version">缘来婚恋 v1.0.0</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 24rpx 24rpx 60rpx;
}

.profile-card {
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx 24rpx;
}

.avatar {
  width: 128rpx;
  height: 128rpx;
  border-radius: 50%;
  background: #f5f6f8;
  flex-shrink: 0;
}

.info {
  flex: 1;
  margin-left: 24rpx;
  min-width: 0;
}

.name-row {
  display: flex;
  align-items: center;
  margin-bottom: 12rpx;
}

.name {
  font-size: 34rpx;
  font-weight: 600;
}

.role-tag {
  margin-left: 14rpx;
  font-size: 20rpx;
  padding: 3rpx 12rpx;
  border-radius: 6rpx;
  background: #eef5ff;
  color: #3f7ef5;
}

.status {
  font-size: 24rpx;
  margin-bottom: 8rpx;

  &.ok {
    color: #00b578;
  }

  &.warn {
    color: #ff8f1f;
  }

  &.danger {
    color: #f53f3f;
  }
}

.complete {
  font-size: 22rpx;
  color: #bbbfc4;
}

.arrow {
  color: #c9cdd4;
  font-size: 36rpx;
  margin-left: 10rpx;
}

.reject-tip {
  margin-top: 16rpx;
  padding: 20rpx 24rpx;
  background: #ffece8;
  color: #f53f3f;
  border-radius: 12rpx;
  font-size: 24rpx;
  line-height: 1.5;
}

.stat-row {
  display: flex;
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx 0;
  margin-top: 20rpx;
}

.stat {
  flex: 1;
  text-align: center;

  .num {
    display: block;
    font-size: 36rpx;
    font-weight: 600;
    color: #1f2329;
    margin-bottom: 8rpx;
  }

  .lab {
    font-size: 24rpx;
    color: #8a8f99;
  }
}

.menu {
  background: #fff;
  border-radius: 16rpx;
  margin-top: 20rpx;
  overflow: hidden;
}

.menu-title {
  font-size: 24rpx;
  color: #8a8f99;
  padding: 24rpx 24rpx 8rpx;
}

.menu-item {
  display: flex;
  align-items: center;
  min-height: 100rpx;
  padding: 0 24rpx;
  border-bottom: 1rpx solid #f2f3f5;

  &:last-child {
    border-bottom: none;
  }
}

.mi-label {
  flex: 1;
  font-size: 29rpx;
}

.mi-extra {
  font-size: 26rpx;
  color: #bbbfc4;
}

.todo-badge {
  min-width: 36rpx;
  height: 36rpx;
  line-height: 36rpx;
  padding: 0 10rpx;
  border-radius: 18rpx;
  background: #f53f3f;
  color: #fff;
  font-size: 22rpx;
  text-align: center;
}

.version {
  text-align: center;
  font-size: 22rpx;
  color: #c9cdd4;
  padding: 60rpx 0 20rpx;
}
</style>
