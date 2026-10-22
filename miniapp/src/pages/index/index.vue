<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onShow, onPullDownRefresh, onReachBottom } from '@dcloudio/uni-app'
import UserCardItem from '@/components/UserCardItem.vue'
import MatchSuccessOverlay from '@/components/MatchSuccessOverlay.vue'
import { discoverApi, interactionApi } from '@/api'
import type { UserCard } from '@/api/types'
import { Status, useUserStore } from '@/stores/user'

const store = useUserStore()
const list = ref<UserCard[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const firstLoaded = ref(false)
/** 首屏拉取失败：页面要给出重试入口，不能停在白屏 */
const loadFailed = ref(false)

/** 正在离场的卡片 id：先播收起动画，动画结束再从数据里摘掉 */
const leavingIds = ref<number[]>([])

/** 配对成功弹层的状态 */
const matchVisible = ref(false)
const matchPeer = ref<UserCard | null>(null)
const matchConvId = ref(0)

/**
 * 首屏推荐流的加载入口。
 *
 * 冷启动时序坑：App.vue 的 onLaunch 里是 `await store.fetchMe()`（异步），
 * 而本页 onShow 会先一步触发 —— 那一刻 store.user 还是 null、canBrowse 为 false。
 * 旧写法在这里 `if (!store.canBrowse) return`，等于把首次加载整个吞掉，
 * 页面就停在「只剩顶部筛选条、下方全空」，切到别的 tab 再回来才恢复
 * （那时 store 已就绪，onShow 重跑）。
 *
 * 所以这里不再「放弃」，只做条件判断；等 store 就绪由下面的 watch 兜底补上。
 */
function ensureFirstLoad() {
  if (firstLoaded.value || loading.value) return
  if (!store.canBrowse) return
  load(true)
}

onShow(() => {
  ensureFirstLoad()
  store.fetchUnread()
})

// 兜底：store 就绪（canBrowse 由 false 变 true）后把首屏加载和红点补上
watch(
  () => store.canBrowse,
  (ok) => {
    if (!ok) return
    ensureFirstLoad()
    store.fetchUnread()
  },
)

/** 首屏还在等 store 就绪：这段窗口 `loading` 尚未置位，靠 meLoaded 判断 */
const bootstrapping = computed(() => !store.meLoaded && !store.meError && !firstLoaded.value)

/** 真正在拉推荐流（含等待 store 的那段）都显示骨架屏，避免先闪一下空白 */
const showSkeleton = computed(() => bootstrapping.value || (loading.value && !firstLoaded.value))

/** 加载失败后的重试：store 没拿到就先补 store，再拉推荐流 */
async function retryBootstrap() {
  if (!store.meLoaded) {
    try {
      await store.fetchMe()
    } catch {
      return // 仍失败则保持错误态，用户可再点一次
    }
  }
  await load(true)
}

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
    loadFailed.value = false
  } catch {
    // request 层已提示。这里记下失败态：否则首屏拉取失败时
    // firstLoaded 不置位、loading 已归位，模板各分支全不满足 → 又是白屏
    loadFailed.value = true
  } finally {
    loading.value = false
  }
}

/** 卡片收起动画时长，与 UserCardItem 里 card-leave 的 380ms 对齐 */
const LEAVE_MS = 380

async function handleLike(user: UserCard) {
  if (!store.canInteract) {
    uni.showToast({ title: '资料审核通过后才能表达心意', icon: 'none' })
    return
  }

  try {
    const res = await interactionApi.like(user.id)

    // 先让卡片自己收起来，别直接抽走 —— 视觉上有「处理完成」的交代
    leavingIds.value = [...leavingIds.value, user.id]

    // 必须等收起动画播完再动数据。
    // 原来的写法是挂上 leaving 后立刻 load(true)，那张卡的 DOM 会被直接卸载，
    // card-leave 断在半路，用户看到的仍然是「瞬间消失」，动画等于白做。
    // 这里等动画结束后再重载，其余卡片靠 key 复用 DOM，不会重播入场动画。
    await new Promise((resolve) => setTimeout(resolve, LEAVE_MS))
    leavingIds.value = leavingIds.value.filter((id) => id !== user.id)

    // 重载第一页：推荐流是 OFFSET 分页，结果集缩小后不重载会让下一页漏读一条
    await load(true)

    if (res.is_match) {
      matchPeer.value = user
      matchConvId.value = res.conversation_id
      matchVisible.value = true
    } else {
      uni.showToast({ title: '已表达心意', icon: 'none' })
    }
  } catch {
    // request 层已提示。失败时上面的 leaving 标记要撤掉，否则卡片卡在收起态
    leavingIds.value = leavingIds.value.filter((id) => id !== user.id)
  }
}

function goMatchedChat() {
  const peer = matchPeer.value
  matchVisible.value = false
  if (!peer || !matchConvId.value) return
  uni.navigateTo({
    url: `/pages/chat/index?id=${matchConvId.value}&name=${encodeURIComponent(peer.nickname)}&peer=${peer.id}`,
  })
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
    <view v-if="store.status === Status.Auditing" class="notice">
      资料审核中，通过后即可与他人互动
    </view>
    <view v-else-if="store.status === Status.Rejected" class="notice reject">
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

    <view v-if="showSkeleton" class="skeleton-list" aria-label="正在加载推荐">
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
        v-for="(u, i) in list"
        :key="u.id"
        :user="u"
        :index="i"
        :leaving="leavingIds.includes(u.id)"
        show-like
        @like="handleLike"
      />
      <view v-if="finished" class="end">暂时没有更多了</view>
      <view v-else class="end">加载中…</view>
    </view>

    <view v-else-if="loadFailed || (store.meError && !firstLoaded)" class="empty">
      <view class="empty-mark">缘</view>
      <view class="empty-title">内容没能加载出来</view>
      <view class="empty-sub">可能是网络不太稳定，点下面重试一下</view>
      <view class="empty-btn" @click="retryBootstrap">重新加载</view>
    </view>

    <view v-else-if="firstLoaded" class="empty">
      <view class="empty-mark">缘</view>
      <view class="empty-title">暂时没有更多推荐</view>
      <view class="empty-sub">试试放宽择偶要求，或稍后再来看看</view>
      <view class="empty-btn" @click="goRequirement">调整择偶要求</view>
    </view>

    <MatchSuccessOverlay
      :visible="matchVisible"
      :my-avatar="store.user?.avatar"
      :peer-avatar="matchPeer?.avatar"
      :peer-name="matchPeer?.nickname || '对方'"
      @chat="goMatchedChat"
      @close="matchVisible = false"
    />
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
  font-size: $font-caption;
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
  font-size: $font-body;
  font-weight: 650;
}

.filter-text {
  font-size: $font-caption;
  color: $text-weak;
}

.filter-action {
  padding: 10rpx 16rpx;
  border-radius: $radius;
  background: $bg-raised;
  color: $brand;
  border: 1rpx solid $border;
  font-size: $font-caption;
  font-weight: 600;
}

.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.skeleton-card {
  display: grid;
  grid-template-columns: 224rpx 1fr;
  gap: 24rpx;
  padding: 18rpx;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  background: $bg-raised;
  box-shadow: $shadow-card;
}

.skeleton-avatar {
  width: 224rpx;
  height: 288rpx;
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
  font-size: $font-caption;
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
  font-size: $font-lead;
  font-weight: 600;
}

.empty-title {
  font-size: $font-section;
  font-weight: 600;
  color: $text-main;
  margin-bottom: 12rpx;
}

.empty-sub {
  font-size: $font-mid;
  color: $text-weak;
  line-height: 1.7;
  margin-bottom: 40rpx;
}

.empty-btn {
  display: inline-block;
  padding: 0 36rpx;
  height: 72rpx;
  line-height: 72rpx;
  border-radius: $radius;
  background: $brand-light;
  color: $brand;
  font-size: $font-body;
  font-weight: 500;

  &:active {
    transform: scale(0.98);
  }
}
</style>
