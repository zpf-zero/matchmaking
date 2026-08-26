<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { adminApi } from '@/api'
import { Role, useUserStore } from '@/stores/user'

const store = useUserStore()
const list = ref<any[]>([])
const page = ref(1)
const finished = ref(false)
const loading = ref(false)
const loaded = ref(false)
const filterStatus = ref(0) // 0 待处理 / -1 全部

const statusText = ['待处理', '已处理', '已驳回']

onShow(() => load(true))

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
    const res = await adminApi.reportList(filterStatus.value, page.value)
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

function switchFilter(s: number) {
  filterStatus.value = s
  load(true)
}

function handle(item: any) {
  const canBan = store.role >= Role.Admin
  const items = canBan ? ['忽略（无问题）', '封禁被举报用户'] : ['忽略（无问题）']

  uni.showActionSheet({
    itemList: items,
    success: (res) => {
      if (res.tapIndex === 0) doHandle(item, 'ignore', '核实后未发现问题')
      else confirmBan(item)
    },
  })
}

function confirmBan(item: any) {
  uni.showModal({
    title: '封禁原因',
    editable: true,
    placeholderText: '将展示给被封禁用户',
    success: (r) => {
      if (r.confirm) doHandle(item, 'ban', r.content?.trim() || '违反社区规范')
    },
  })
}

async function doHandle(item: any, action: 'ignore' | 'ban', remark: string) {
  try {
    await adminApi.handleReport(item.report.id, action, remark)
    uni.showToast({ title: '已处理', icon: 'none' })
    load(true)
  } catch {
    // 已提示
  }
}

function goUser(id: number) {
  uni.navigateTo({ url: `/pages/admin/user-detail?id=${id}` })
}
</script>

<template>
  <view class="page">
    <view class="filter">
      <view class="fi" :class="{ on: filterStatus === 0 }" @click="switchFilter(0)">待处理</view>
      <view class="fi" :class="{ on: filterStatus === -1 }" @click="switchFilter(-1)">全部</view>
    </view>

    <view v-for="item in list" :key="item.report.id" class="card">
      <view class="row-top">
        <text class="reason">{{ item.reason_text }}</text>
        <text class="status" :class="'s' + item.report.status">
          {{ statusText[item.report.status] }}
        </text>
      </view>

      <view class="line">
        <text class="k">举报人</text>
        <text class="v">{{ item.from_nickname }}（#{{ item.report.from_user_id }}）</text>
      </view>
      <view class="line" @click="goUser(item.report.target_user_id)">
        <text class="k">被举报</text>
        <text class="v link">{{ item.target_nickname }}（#{{ item.report.target_user_id }}）›</text>
      </view>

      <view v-if="item.report.description" class="desc">{{ item.report.description }}</view>
      <view v-if="item.report.handle_remark" class="remark">
        处理备注：{{ item.report.handle_remark }}
      </view>

      <view v-if="item.report.status === 0" class="actions">
        <view class="btn" @click="handle(item)">处理</view>
      </view>
    </view>

    <view v-if="loaded && !list.length" class="empty">没有需要处理的举报</view>
    <view v-else-if="list.length && finished" class="end">已经到底啦</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.filter {
  display: flex;
  gap: 16rpx;
  padding-bottom: 24rpx;
}

.fi {
  padding: 10rpx 32rpx;
  border-radius: $radius-sm;
  border: 1rpx solid $border;
  background: $bg-card;
  color: $text-sub;
  font-size: 26rpx;

  &.on {
    background: $brand;
    border-color: $brand;
    color: $text-on-accent;
  }
}

.card {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 28rpx;
  margin-bottom: 20rpx;
}

.row-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.reason {
  font-size: 30rpx;
  font-weight: 600;
}

.status {
  font-size: 22rpx;
  padding: 4rpx 14rpx;
  border-radius: 6rpx;

  &.s0 {
    background: $warning-light;
    color: $warning;
  }

  &.s1 {
    background: $success-light;
    color: $success;
  }

  &.s2 {
    background: $bg-subtle;
    color: $text-weak;
  }
}

.line {
  display: flex;
  font-size: 26rpx;
  padding: 8rpx 0;

  .k {
    color: $text-weak;
    width: 110rpx;
  }

  .v {
    flex: 1;
  }

  .link {
    color: $brand;
  }
}

.desc {
  margin-top: 14rpx;
  padding: 18rpx;
  background: $bg-subtle;
  border-radius: 10rpx;
  font-size: 25rpx;
  color: $text-sub;
  line-height: 1.6;
}

.remark {
  margin-top: 14rpx;
  font-size: 24rpx;
  color: $text-weak;
}

.actions {
  padding-top: 24rpx;
}

.btn {
  height: 76rpx;
  line-height: 76rpx;
  text-align: center;
  border-radius: 12rpx;
  background: $brand;
  color: $text-on-accent;
  font-size: 28rpx;
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: 24rpx;
  padding: 30rpx 0;
}
</style>
