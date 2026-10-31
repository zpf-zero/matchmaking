<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import UserAvatar from '@/components/UserAvatar.vue'
import AppTabBar from '@/components/AppTabBar.vue'
import { adminApi, interactionApi } from '@/api'
import { Status, useUserStore } from '@/stores/user'

const store = useUserStore()
const todo = ref(0)
const myLikes = ref(0)

onShow(async () => {
  // 底部导航改为自绘（「消息」项受私聊开关控制），原生那条全程藏着
  uni.hideTabBar({ animation: false, fail: () => {} })
  if (!store.isLogin) return
  try {
    await store.fetchMe()
    store.fetchUnread()

    const tasks: Promise<unknown>[] = []
    // 三格数据必须同口径，否则「我喜欢的」会是个空位
    if (store.canBrowse) {
      tasks.push(
        interactionApi
          .myLikes(1)
          .then((r) => {
            myLikes.value = r.total || 0
          })
          .catch(() => {})
      )
    }
    // 管理员额外拉一下待办数，直接显示在入口上
    if (store.isAdminRole) {
      tasks.push(
        adminApi
          .overview()
          .then((ov) => {
            todo.value = ov.pending_audit + ov.pending_report
          })
          .catch(() => {})
      )
    }
    await Promise.all(tasks)
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
      <UserAvatar
        :src="store.user?.avatar"
        :name="store.user?.nickname"
        :size="136"
        radius="lg"
      />
      <view class="info">
        <view class="name-row">
          <text class="name">{{ store.user?.nickname || '未填写' }}</text>
          <text v-if="roleText" class="role-tag">{{ roleText }}</text>
        </view>
        <view class="profile-meta">
          <view class="status" :class="statusText.cls">{{ statusText.text }}</view>
          <view v-if="store.profile" class="complete">
            完整度 {{ store.profile.complete_rate }}%
          </view>
        </view>
        <view v-if="store.profile" class="complete-track" aria-label="资料完整度">
          <view
            class="complete-value"
            :style="{ width: `${store.profile.complete_rate}%` }"
          ></view>
        </view>
      </view>
      <text class="profile-action">编辑</text>
    </view>

    <view v-if="store.status === Status.Rejected" class="reject-tip">
      未通过原因：{{ store.profile?.audit_remark || '资料不符合规范' }}
    </view>

    <!-- 数据 -->
    <view class="stat-grid">
      <view class="stat stat-featured" @click="go('/pages/interaction/list?tab=0')">
        <text class="num">{{ store.unread.liked_me }}</text>
        <text class="lab">新的喜欢</text>
      </view>
      <view class="stat" @click="go('/pages/interaction/list?tab=1')">
        <text class="num">{{ myLikes }}</text>
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
        <view class="mi-copy">
          <text class="mi-label">我的资料</text>
          <text class="mi-note">基础信息与自我介绍</text>
        </view>
        <text class="mi-action">编辑</text>
      </view>
      <view class="menu-item" @click="go('/pages/user/requirement')">
        <view class="mi-copy">
          <text class="mi-label">择偶要求</text>
          <text class="mi-note">年龄、城市与生活状态</text>
        </view>
        <text class="mi-action">调整</text>
      </view>
      <view class="menu-item" @click="go('/pages/user/photos')">
        <view class="mi-copy">
          <text class="mi-label">我的相册</text>
          <text class="mi-note">展示真实的生活状态</text>
        </view>
        <text class="mi-extra">{{ store.photos.length }}/{{ store.maxPhotoNum }}</text>
      </view>
      <!-- 三观问答：答题免费，是提高资料质量的核心抓手。
           放在「我的」而不是藏进设置里 —— 它是内容入口，不是配置项 -->
      <view class="menu-item" @click="go('/pages/user/three-view')">
        <view class="mi-copy">
          <text class="mi-label">三观问答</text>
          <text class="mi-note">谈谈想法，更容易遇到观念相合的人</text>
        </view>
        <text class="mi-action">去回答</text>
      </view>
      <view class="menu-item" @click="go('/pages/invite/index')">
        <view class="mi-copy">
          <text class="mi-label">邀请有礼</text>
          <text class="mi-note">
            {{ store.isVip ? '已解锁完整资料' : '邀请好友，双方永久解锁' }}
          </text>
        </view>
        <!-- 好友过审即结算，但邀请人多半不在场：红点把人带回邀请页看回执 -->
        <view v-if="store.inviteUnseen > 0" class="todo-badge">{{ store.inviteUnseen }}</view>
        <text class="mi-action">进入</text>
      </view>
      <view class="menu-item" @click="go('/pages/setting/index')">
        <view class="mi-copy">
          <text class="mi-label">设置</text>
          <text class="mi-note">隐私、黑名单与账号安全</text>
        </view>
        <text class="mi-action">进入</text>
      </view>
    </view>

    <!-- 管理入口：role > 1 才渲染。真正的拦截在后端 RequireRole -->
    <view v-if="store.isAdminRole" class="menu admin-menu">
      <view class="menu-title">管理功能</view>
      <view class="menu-item" @click="go('/pages/admin/index')">
        <view class="mi-copy">
          <text class="mi-label">管理中心</text>
          <text class="mi-note">审核资料、举报和用户</text>
        </view>
        <view v-if="todo" class="todo-badge">{{ todo }}</view>
        <text class="mi-action">进入</text>
      </view>
    </view>

  </view>

  <!-- 放在 .page 之外：.page 有入场动画（终态保留 transform），
       会成为 fixed 的定位祖先，把它沉到内容末尾 -->
  <AppTabBar class="tabbar-host" active="mine" />
</template>

<style lang="scss" scoped>
.page {
  min-height: calc(100dvh - 188rpx);
  /* 底部 100rpx 自绘 tabBar + 安全区，避免最后一项被挡住 */
  padding: 28rpx 24rpx calc(64rpx + 100rpx + env(safe-area-inset-bottom));
}

/* 宿主节点钉在视口底部（必须放在 .page 外面，否则被入场动画的 transform 劫持定位） */
.tabbar-host {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 200;
  /* 宽屏 H5 上跟随 .page 的 540px 单列画布，不拉成通栏 */
  max-width: 540px;
  margin-left: auto;
  margin-right: auto;
}

.profile-card {
  display: flex;
  align-items: center;
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 34rpx 30rpx;
  box-shadow: $shadow-card;
  transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1),
    border-color 200ms ease, box-shadow 200ms ease;

  &:active {
    transform: scale(0.985);
    border-color: $border-strong;
    box-shadow: none;
  }
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
  font-size: $font-title;
  font-weight: 650;
  color: $text-main;
  letter-spacing: -0.5rpx;
}

.role-tag {
  margin-left: 14rpx;
  font-size: $font-micro;
  padding: 3rpx 12rpx;
  border-radius: $radius-sm;
  background: $brand-light;
  color: $brand;
}

.profile-meta {
  display: flex;
  align-items: center;
  gap: 14rpx;
  margin-bottom: 13rpx;
}

.status {
  font-size: $font-caption;
  font-weight: 500;

  &.ok {
    color: $success;
  }

  &.warn {
    color: $warning;
  }

  &.danger {
    color: $danger;
  }
}

.complete {
  font-size: $font-micro;
  color: $text-placeholder;
}

.complete-track {
  height: 4rpx;
  background: $bg-subtle;
  border-radius: $radius-sm;
  overflow: hidden;
}

.complete-value {
  height: 100%;
  background: $brand;
  border-radius: inherit;
}

.profile-action,
.mi-action {
  margin-left: 18rpx;
  color: $brand;
  font-size: $font-caption;
  font-weight: 600;
  white-space: nowrap;
}

.reject-tip {
  margin-top: 16rpx;
  padding: 20rpx 24rpx;
  background: $danger-light;
  color: $danger;
  border-left: 4rpx solid $danger;
  border-radius: $radius-sm;
  font-size: $font-caption;
  line-height: 1.5;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14rpx;
  margin-top: 24rpx;
}

.stat {
  min-height: 138rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius;
  transition: transform 180ms ease, border-color 180ms ease;

  &:active {
    transform: scale(0.97);
    border-color: $border-strong;
  }

  .num {
    display: block;
    font-size: $font-title;
    font-weight: 650;
    color: $text-main;
    margin-bottom: 8rpx;
  }

  .lab {
    font-size: $font-micro;
    color: $text-weak;
  }
}

.stat-featured {
  background: $bg-accent;
  border-color: $brand-light;

  .num {
    color: $brand;
  }
}

.menu {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  margin-top: 24rpx;
  overflow: hidden;
  box-shadow: $shadow-card;
}

.admin-menu {
  border-color: $brand-light;
}

.menu-title {
  font-size: $font-caption;
  color: $text-weak;
  padding: 24rpx 24rpx 6rpx;
}

.menu-item {
  display: flex;
  align-items: center;
  min-height: 120rpx;
  padding: 18rpx 24rpx;
  border-bottom: 1rpx solid $border;
  transition: background-color 180ms ease;

  &:active {
    background: $bg-subtle;
  }

  &:last-child {
    border-bottom: none;
  }
}

.mi-label {
  font-size: $font-body;
  font-weight: 520;
  color: $text-main;
}

.mi-copy {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5rpx;
}

.mi-note {
  font-size: $font-micro;
  color: $text-placeholder;
}

.mi-extra {
  font-size: $font-mid;
  color: $text-placeholder;
}

.todo-badge {
  min-width: 36rpx;
  height: 36rpx;
  line-height: 36rpx;
  padding: 0 10rpx;
  border-radius: $radius-sm;
  background: $danger;
  color: $text-on-accent;
  font-size: $font-micro;
  text-align: center;
}
</style>
