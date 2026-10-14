<script setup lang="ts">
import { ref } from 'vue'
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
  }

  const fn = [interactionApi.likedMe, interactionApi.myLikes, interactionApi.viewedMe][active.value]

  try {
    const res = await fn(page.value)
    list.value = reset ? res.list || [] : [...list.value, ...(res.list || [])]
    finished.value = list.value.length >= res.total
    if (!finished.value) page.value++
    loaded.value = true
  } catch {
    loaded.value = true
  } finally {
    loading.value = false
  }
}

const emptyText = () =>
  ['还没有人喜欢你，完善资料能提高曝光', '你还没有喜欢过任何人', '还没有人看过你的资料'][active.value]
</script>

<template>
  <view class="page">
    <view class="tabs">
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

    <view class="list">
      <UserCardItem v-for="u in list" :key="u.id" :user="u" />
      <view v-if="loaded && !list.length" class="empty">{{ emptyText() }}</view>
      <view v-else-if="list.length && finished" class="end">已经到底啦</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 0 24rpx 48rpx;
}

.tabs {
  display: flex;
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  margin: 24rpx 0;
  padding: 8rpx;
}

.tab {
  flex: 1;
  height: 72rpx;
  line-height: 72rpx;
  text-align: center;
  font-size: $font-mid;
  color: $text-sub;
  border-radius: $radius;
  transition: background-color 180ms ease, color 180ms ease;

  &.on {
    background: $brand-light;
    color: $brand;
    font-weight: 500;
  }
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: $font-caption;
  padding: 36rpx 0;
}
</style>
