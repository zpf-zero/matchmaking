<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import UserCardItem from '@/components/UserCardItem.vue'
import { discoverApi, interactionApi } from '@/api'
import type { UserCard } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const list = ref<UserCard[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const firstLoaded = ref(false)

onShow(() => {
  if (!store.canBrowse) return
  if (!firstLoaded.value) load(true)
  store.fetchUnread()
})

onPullDownRefresh(async () => {
  await load(true)
  uni.stopPullDownRefresh()
})

onReachBottom(() => {
  if (!finished.value) load(false)
})

async function load(reset: boolean) {
  if (loading.value) return
  loading.value = true

  if (reset) {
    page.value = 1
    finished.value = false
  }

  try {
    const res = await discoverApi.recommend(page.value)
    total.value = res.total
    list.value = reset ? res.list || [] : [...list.value, ...(res.list || [])]
    finished.value = list.value.length >= res.total
    if (!finished.value) page.value++
    firstLoaded.value = true
  } catch {
    // request 层已处理
  } finally {
    loading.value = false
  }
}

async function handleLike(user: UserCard) {
  if (!store.canInteract) {
    uni.showToast({ title: '资料审核通过后才能表达心意', icon: 'none' })
    return
  }

  try {
    const res = await interactionApi.like(user.id)
    // 先本地移除保证操作即时生效，再拉第一页避免 OFFSET 翻页漏掉一条
    list.value = list.value.filter((item) => item.id !== user.id)
    total.value = Math.max(0, total.value - 1)
    await load(true)

    if (res.is_match) {
      uni.showModal({
        title: '配对成功',
        content: `你和 ${user.nickname} 互相喜欢，现在可以聊天了`,
        confirmText: '去聊天',
        cancelText: '稍后',
        success: (r) => {
          if (r.confirm) {
            uni.navigateTo({
              url: `/pages/chat/index?id=${res.conversation_id}&name=${encodeURIComponent(user.nickname)}&peer=${user.id}`,
            })
          }
        },
      })
    } else {
      uni.showToast({ title: '已表达心意', icon: 'none' })
    }
  } catch {
    // request 层已提示
  }
}

function goSearch() {
  uni.switchTab({ url: '/pages/search/index' })
}

function goEditProfile() {
  uni.navigateTo({ url: '/pages/profile/setup?mode=edit' })
}

function goRequirement() {
  uni.navigateTo({ url: '/pages/user/requirement' })
}
</script>

<template>
  <view class="page">
    <!-- 审核状态提示 -->
    <view v-if="store.status === 1" class="notice">
      资料审核中，通过后即可与他人互动
    </view>
    <view v-else-if="store.status === 3" class="notice reject">
      资料未通过：{{ store.profile?.audit_remark }}
      <text class="notice-btn" @click="goEditProfile">去修改</text>
    </view>

    <view class="filter-bar" @click="goSearch">
      <view class="filter-copy">
        <text class="filter-title">精确筛选</text>
        <text class="filter-text">年龄、学历、收入与城市</text>
      </view>
      <text class="filter-action">调整</text>
    </view>

    <view v-if="loading && !firstLoaded" class="skeleton-list" aria-label="正在加载推荐">
      <view v-for="n in 3" :key="n" class="skeleton-card">
        <view class="skeleton skeleton-avatar"></view>
        <view class="skeleton-lines">
          <view class="skeleton skeleton-line strong"></view>
          <view class="skeleton skeleton-line"></view>
          <view class="skeleton skeleton-line short"></view>
        </view>
      </view>
    </view>

    <view v-if="list.length" class="list">
      <UserCardItem
        v-for="u in list"
        :key="u.id"
        :user="u"
        show-like
        @like="handleLike"
      />
      <view v-if="finished" class="end">暂时没有更多了</view>
      <view v-else class="end">加载中…</view>
    </view>

    <view v-else-if="firstLoaded" class="empty">
      <view class="empty-mark">缘</view>
      <view class="empty-title">暂时没有更多推荐</view>
      <view class="empty-sub">试试放宽择偶要求，或稍后再来看看</view>
      <view class="empty-btn" @click="goRequirement">调整择偶要求</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  min-height: calc(100dvh - 188rpx);
  padding: 24rpx 24rpx 48rpx;
}

.notice {
  background: $warning-light;
  color: $warning;
  border-left: 4rpx solid $warning;
  font-size: 24rpx;
  padding: 18rpx 20rpx;
  border-radius: $radius-sm;
  margin-bottom: 24rpx;
  line-height: 1.6;

  &.reject {
    background: $danger-light;
    color: $danger;
    border-left-color: $danger;
  }
}

.notice-btn {
  margin-left: 12rpx;
  text-decoration: underline;
}

.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 112rpx;
  background: $bg-accent;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 22rpx 26rpx;
  margin-bottom: 28rpx;
  box-shadow: $shadow-card;
  transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1),
    border-color 200ms ease, box-shadow 200ms ease;

  &:active {
    transform: scale(0.985);
    border-color: $border-strong;
    box-shadow: none;
  }
}

.filter-copy {
  display: flex;
  flex-direction: column;
  gap: 3rpx;
}

.filter-title {
  color: $text-main;
  font-size: 29rpx;
  font-weight: 650;
}

.filter-text {
  font-size: 23rpx;
  color: $text-weak;
}

.filter-action {
  padding: 10rpx 16rpx;
  border-radius: $radius-sm;
  background: $bg-raised;
  color: $brand;
  border: 1rpx solid $border;
  font-size: 23rpx;
  font-weight: 600;
}

.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.skeleton-card {
  display: grid;
  grid-template-columns: 202rpx 1fr;
  gap: 26rpx;
  padding: 18rpx;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  background: $bg-raised;
  box-shadow: $shadow-card;
}

.skeleton-avatar {
  width: 202rpx;
  height: 264rpx;
  border-radius: $radius;
}

.skeleton-lines {
  padding-top: 14rpx;
}

.skeleton-line {
  height: 22rpx;
  border-radius: $radius-sm;
  margin-bottom: 24rpx;

  &.strong {
    width: 76%;
    height: 30rpx;
  }

  &.short {
    width: 54%;
  }
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: 24rpx;
  padding: 36rpx 0 20rpx;
}

.empty {
  padding: 128rpx 40rpx;
  text-align: center;
}

.empty-mark {
  width: 72rpx;
  height: 72rpx;
  line-height: 72rpx;
  margin: 0 auto 28rpx;
  border-radius: $radius;
  background: $brand-light;
  color: $brand;
  font-size: 30rpx;
  font-weight: 600;
}

.empty-title {
  font-size: 31rpx;
  font-weight: 600;
  color: $text-main;
  margin-bottom: 12rpx;
}

.empty-sub {
  font-size: 26rpx;
  color: $text-weak;
  line-height: 1.7;
  margin-bottom: 40rpx;
}

.empty-btn {
  display: inline-block;
  padding: 0 36rpx;
  height: 72rpx;
  line-height: 72rpx;
  border-radius: $radius-sm;
  background: $brand-light;
  color: $brand;
  font-size: 28rpx;
  font-weight: 500;

  &:active {
    transform: scale(0.98);
  }
}
</style>
