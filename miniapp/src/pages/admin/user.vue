<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { adminApi } from '@/api'
import type { AdminUserItem } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const list = ref<AdminUserItem[]>([])
const keyword = ref('')
const status = ref(-1)
const page = ref(1)
const total = ref(0)
const finished = ref(false)
const loading = ref(false)
const loaded = ref(false)

const statusOptions = [
  { v: -1, t: '全部' },
  { v: 1, t: '待审核' },
  { v: 2, t: '正常' },
  { v: 3, t: '未通过' },
  { v: 4, t: '已封禁' },
]

const statusText = ['待完善', '待审核', '正常', '未通过', '已封禁']

onShow(async () => {
  await store.fetchDict()
  load(true)
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
    const res = await adminApi.userList({
      keyword: keyword.value.trim(),
      status: status.value,
      page: page.value,
    })
    total.value = res.total
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

function setStatus(v: number) {
  status.value = v
  load(true)
}

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/admin/user-detail?id=${id}` })
}
</script>

<template>
  <view class="page">
    <view class="search">
      <input
        v-model="keyword"
        class="search-input"
        placeholder="搜索昵称 / 手机号 / ID"
        placeholder-class="ph"
        confirm-type="search"
        @confirm="load(true)"
      />
      <view class="search-btn" @click="load(true)">搜索</view>
    </view>

    <scroll-view class="filter" scroll-x>
      <view
        v-for="o in statusOptions"
        :key="o.v"
        class="fi"
        :class="{ on: status === o.v }"
        @click="setStatus(o.v)"
      >
        {{ o.t }}
      </view>
    </scroll-view>

    <view class="count">共 {{ total }} 人</view>

    <view v-for="u in list" :key="u.id" class="item" @click="goDetail(u.id)">
      <image :src="u.avatar" class="avatar" mode="aspectFill" />
      <view class="info">
        <view class="name-row">
          <text class="name">{{ u.nickname || '未填写' }}</text>
          <text class="uid">#{{ u.id }}</text>
          <text v-if="u.role > 1" class="role">
            {{ ['', '', '审核员', '管理员', '超管'][u.role] }}
          </text>
        </view>
        <view class="meta">
          {{ u.age ? u.age + '岁 / ' : '' }}{{ store.label('gender', u.gender) }}
          {{ u.city ? ' / ' + u.city : '' }}
        </view>
        <view class="meta">{{ u.phone || '未绑定手机号' }}</view>
      </view>
      <view class="status" :class="'s' + u.status">{{ statusText[u.status] }}</view>
    </view>

    <view v-if="loaded && !list.length" class="empty">没有符合条件的用户</view>
    <view v-else-if="list.length && finished" class="end">已经到底啦</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.search {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.search-input {
  flex: 1;
  height: 76rpx;
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-sm;
  padding: 0 22rpx;
  font-size: 27rpx;
}

.ph {
  color: $text-placeholder;
}

.search-btn {
  padding: 0 32rpx;
  height: 76rpx;
  line-height: 76rpx;
  border-radius: $radius-sm;
  background: $brand;
  color: $text-on-accent;
  font-size: 27rpx;
}

.filter {
  white-space: nowrap;
  margin-bottom: 16rpx;
}

.fi {
  display: inline-block;
  padding: 8rpx 28rpx;
  margin-right: 14rpx;
  border-radius: $radius-sm;
  border: 1rpx solid $border;
  background: $bg-card;
  color: $text-sub;
  font-size: 25rpx;

  &.on {
    background: $brand;
    border-color: $brand;
    color: $text-on-accent;
  }
}

.count {
  font-size: 24rpx;
  color: $text-weak;
  padding: 4rpx 8rpx 16rpx;
}

.item {
  display: flex;
  align-items: center;
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 24rpx;
  margin-bottom: 16rpx;
}

.avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: $radius;
  background: $bg-subtle;
  flex-shrink: 0;
}

.info {
  flex: 1;
  margin-left: 20rpx;
  min-width: 0;
}

.name-row {
  display: flex;
  align-items: center;
  margin-bottom: 8rpx;
}

.name {
  font-size: 29rpx;
  font-weight: 500;
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.uid {
  font-size: 22rpx;
  color: $text-placeholder;
  margin-left: 12rpx;
}

.role {
  font-size: 20rpx;
  color: $brand;
  background: $brand-light;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin-left: 10rpx;
}

.meta {
  font-size: 23rpx;
  color: $text-weak;
  margin-top: 4rpx;
}

.status {
  font-size: 22rpx;
  padding: 6rpx 16rpx;
  border-radius: 6rpx;
  flex-shrink: 0;
  background: $bg-subtle;
  color: $text-weak;

  &.s1 {
    background: $warning-light;
    color: $warning;
  }

  &.s2 {
    background: $success-light;
    color: $success;
  }

  &.s3,
  &.s4 {
    background: $danger-light;
    color: $danger;
  }
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: 24rpx;
  padding: 30rpx 0;
}
</style>
