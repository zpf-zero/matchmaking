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
    // 表过态的人不再出现在推荐流，本地同步移除，避免刷新才消失
    list.value = list.value.filter((u) => u.id !== user.id)
    total.value = Math.max(0, total.value - 1)

    if (res.is_match) {
      uni.showModal({
        title: '配对成功',
        content: `你和 ${user.nickname} 互相喜欢，现在可以聊天了`,
        confirmText: '去聊天',
        cancelText: '稍后',
        success: (r) => {
          if (r.confirm) {
            uni.navigateTo({
              url: `/pages/chat/index?id=${res.conversation_id}&name=${encodeURIComponent(user.nickname)}`,
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

    <!-- 快捷筛选 -->
    <view class="filter-bar" @click="goSearch">
      <text class="filter-icon">🔍</text>
      <text class="filter-text">按年龄、学历、收入、城市筛选</text>
    </view>

    <view v-if="list.length" class="list">
      <UserCardItem
        v-for="u in list"
        :key="u.id"
        :user="u"
        show-like
        @like="handleLike"
      />
      <view v-if="finished" class="end">已经到底啦</view>
      <view v-else class="end">加载中...</view>
    </view>

    <view v-else-if="firstLoaded" class="empty">
      <view class="empty-title">暂时没有更多推荐</view>
      <view class="empty-sub">试试放宽择偶要求，或稍后再来看看</view>
      <view class="empty-btn" @click="goRequirement">调整择偶要求</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 20rpx 24rpx 40rpx;
}

.notice {
  background: #fff7e8;
  color: #ff8f1f;
  font-size: 24rpx;
  padding: 18rpx 24rpx;
  border-radius: 12rpx;
  margin-bottom: 20rpx;

  &.reject {
    background: #ffece8;
    color: #f53f3f;
  }
}

.notice-btn {
  margin-left: 12rpx;
  text-decoration: underline;
}

.filter-bar {
  display: flex;
  align-items: center;
  height: 76rpx;
  background: #fff;
  border-radius: 38rpx;
  padding: 0 28rpx;
  margin-bottom: 24rpx;
}

.filter-icon {
  font-size: 26rpx;
  margin-right: 12rpx;
}

.filter-text {
  font-size: 26rpx;
  color: #bbbfc4;
}

.end {
  text-align: center;
  color: #bbbfc4;
  font-size: 24rpx;
  padding: 30rpx 0;
}

.empty {
  padding: 160rpx 40rpx;
  text-align: center;
}

.empty-title {
  font-size: 30rpx;
  color: #646a73;
  margin-bottom: 16rpx;
}

.empty-sub {
  font-size: 26rpx;
  color: #bbbfc4;
  margin-bottom: 48rpx;
}

.empty-btn {
  display: inline-block;
  padding: 0 48rpx;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: 38rpx;
  background: #fff0f3;
  color: #e8506d;
  font-size: 28rpx;
}
</style>
