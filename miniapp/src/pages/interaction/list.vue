<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad, onReachBottom } from '@dcloudio/uni-app'
import UserCardItem from '@/components/UserCardItem.vue'
import { interactionApi } from '@/api'
import type { UserCard } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const tabs = ['谁喜欢我', '我喜欢的', '谁看过我']
const active = ref(0)
const list = ref<UserCard[]>([])
const page = ref(1)
const finished = ref(false)
const loading = ref(false)
const loaded = ref(false)

/** 切换 Tab 时给结果区一个轻量淡入，避免内容瞬间换脸 */
const contentKey = ref(0)

const indicatorStyle = computed(() => ({
  transform: `translateX(${active.value * 100}%)`,
}))

/** 「谁看过我」被权益锁住的人数。0 表示未锁定 */
const locked = ref(0)

onLoad(async (q) => {
  active.value = Number(q?.tab || 0)
  await store.fetchDict()
  switchTab(active.value)
})

onReachBottom(() => {
  if (!finished.value) load(false)
})

function switchTab(i: number) {
  active.value = i
  list.value = []
  loaded.value = false
  contentKey.value++
  load(true)

  // 进入即清红点
  if (i === 0) interactionApi.markRead(2).then(() => store.fetchUnread())
  if (i === 2) interactionApi.markRead(1).then(() => store.fetchUnread())
}

async function load(reset: boolean) {
  if (loading.value) return
  loading.value = true
  if (reset) {
    page.value = 1
    finished.value = false
    locked.value = 0
  }

  try {
    // 「谁看过我」返回结构与另两个 Tab 不同：非会员只给前 N 条，
    // 其余以 locked 计数返回，需要会员权益才能解锁
    if (active.value === 2) {
      const res = await interactionApi.viewedMe(page.value)
      list.value = reset ? res.list || [] : [...list.value, ...(res.list || [])]
      locked.value = res.locked || 0
      // 有锁定内容时不再继续翻页（后端也只给第一页）
      finished.value = res.locked > 0 || list.value.length >= res.total
    } else {
      const fn = [interactionApi.likedMe, interactionApi.myLikes][active.value]
      const res = await fn(page.value)
      list.value = reset ? res.list || [] : [...list.value, ...(res.list || [])]
      finished.value = list.value.length >= res.total
    }
    if (!finished.value) page.value++
    loaded.value = true
  } catch {
    loaded.value = true
  } finally {
    loading.value = false
  }
}

function goInvite() {
  uni.navigateTo({ url: '/pages/invite/index' })
}

const emptyText = () =>
  ['还没有人喜欢你，完善资料能提高曝光', '你还没有喜欢过任何人', '还没有人看过你的资料'][active.value]
</script>

<template>
  <view class="page">
    <view class="tabs">
      <view class="indicator" :style="indicatorStyle"></view>
      <view
        v-for="(t, i) in tabs"
        :key="i"
        class="tab"
        :class="{ on: active === i }"
        @click="switchTab(i)"
      >
        {{ t }}
      </view>
    </view>

    <view :key="contentKey" class="list">
      <UserCardItem
        v-for="(u, i) in list"
        :key="u.id"
        :user="u"
        :index="i"
      />

      <!-- 权益解锁提示。刻意只锁「看更多」，不锁任何功能入口 -->
      <view v-if="active === 2 && locked > 0" class="locked-card" @click="goInvite">
        <view class="lock-icon">🔒</view>
        <view class="lock-title">还有 {{ locked }} 位看过你</view>
        <view class="lock-desc">
          免费用户可查看最近 3 位。邀请 1 位好友通过资料审核，双方永久解锁完整列表，无需再次邀请
        </view>
        <view class="lock-btn">邀请好友解锁</view>
      </view>

      <view v-if="loaded && !list.length" class="empty">{{ emptyText() }}</view>
      <view v-else-if="list.length && finished && !locked" class="end">已经到底啦</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 0 24rpx 48rpx;
}

.tabs {
  position: relative;
  display: flex;
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  margin: 24rpx 0;
  padding: 8rpx;
}

/* 指示器宽度按 1/3 走，靠 translateX 平滑滑动，替掉原来直接换背景的生硬感 */
.indicator {
  position: absolute;
  top: 8rpx;
  left: 8rpx;
  width: calc((100% - 16rpx) / 3);
  height: calc(100% - 16rpx);
  border-radius: $radius;
  background: $brand-light;
  transition: transform 300ms cubic-bezier(0.22, 0.61, 0.36, 1);
}

.tab {
  position: relative;
  flex: 1;
  height: 72rpx;
  line-height: 72rpx;
  text-align: center;
  font-size: $font-mid;
  color: $text-sub;
  border-radius: $radius;
  transition: color 220ms ease;

  &.on {
    color: $brand;
    font-weight: 500;
  }
}

.list {
  animation: list-in 320ms ease both;
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: $font-caption;
  padding: 36rpx 0;
}

/* 解锁卡片：把「信息可见性」作为权益点呈现，不遮挡、不弹窗、不做红点 */
.locked-card {
  margin-top: 20rpx;
  padding: 40rpx 32rpx;
  border-radius: $radius-lg;
  background: $bg-accent;
  border: 1rpx solid $brand-light;
  text-align: center;

  &:active {
    opacity: 0.9;
  }
}

.lock-icon {
  font-size: 44rpx;
  line-height: 1;
  margin-bottom: 18rpx;
}

.lock-title {
  font-size: $font-lead;
  font-weight: 600;
  color: $text-main;
  margin-bottom: 12rpx;
}

.lock-desc {
  font-size: $font-caption;
  color: $text-weak;
  line-height: 1.7;
  margin-bottom: 24rpx;
}

.lock-btn {
  display: inline-block;
  padding: 0 44rpx;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: $radius;
  background: $brand;
  color: $text-on-accent;
  font-size: $font-mid;
  font-weight: 600;
}

@keyframes list-in {
  from {
    opacity: 0;
    transform: translateY(12rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
