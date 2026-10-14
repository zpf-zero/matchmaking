<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import { adminApi } from '@/api'
import { parseTime } from '@/utils/time'
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

/** 举报时间，处理队列里“多久之前”比原始时间戳更有用 */
function fmtTime(t?: string) {
  const ts = parseTime(t)
  if (Number.isNaN(ts)) return ''
  const diff = Date.now() - ts
  if (diff < 3600_000) return `${Math.max(1, Math.floor(diff / 60000))} 分钟前`
  if (diff < 86400_000) return `${Math.floor(diff / 3600_000)} 小时前`
  if (diff < 86400_000 * 7) return `${Math.floor(diff / 86400_000)} 天前`
  const d = new Date(ts)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

/** 举报附带的证据图，后端存成逗号分隔的字符串 */
function evidence(item: any): string[] {
  return String(item?.report?.images || '')
    .split(',')
    .map((s: string) => s.trim())
    .filter(Boolean)
}

function previewEvidence(item: any, index: number) {
  const urls = evidence(item)
  if (!urls.length) return
  uni.previewImage({ urls, current: index })
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
      <view class="line tap" @click="goUser(item.report.target_user_id)">
        <text class="k">被举报</text>
        <text class="v link">{{ item.target_nickname }}（#{{ item.report.target_user_id }}）</text>
        <view class="arrow"></view>
      </view>
      <view v-if="item.report.created_at" class="line">
        <text class="k">举报于</text>
        <text class="v weak">{{ fmtTime(item.report.created_at) }}</text>
      </view>

      <view v-if="item.report.description" class="desc">{{ item.report.description }}</view>

      <view v-if="evidence(item).length" class="shots">
        <image
          v-for="(url, i) in evidence(item)"
          :key="url"
          :src="url"
          class="shot"
          mode="aspectFill"
          @click="previewEvidence(item, i)"
        />
      </view>

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
  font-size: $font-mid;
  transition: background-color 180ms ease, color 180ms ease, border-color 180ms ease;

  &.on {
    background: $brand;
    border-color: $brand;
    color: $text-on-accent;
    font-weight: 600;
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
  margin-bottom: 16rpx;
}

.reason {
  font-size: $font-lead;
  font-weight: 600;
  color: $text-main;
}

.status {
  font-size: $font-micro;
  padding: 4rpx 14rpx;
  border-radius: $radius-sm;

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
  align-items: center;
  font-size: $font-mid;
  padding: 8rpx 0;

  &.tap {
    margin: 0 -12rpx;
    padding: 8rpx 12rpx;
    border-radius: $radius-sm;
    transition: background-color 180ms ease;

    &:active {
      background: $bg-subtle;
    }
  }

  .k {
    color: $text-weak;
    width: 110rpx;
    flex-shrink: 0;
  }

  .v {
    flex: 1;
    color: $text-main;

    &.weak {
      color: $text-weak;
    }
  }

  .link {
    color: $brand;
  }
}

.arrow {
  width: 11rpx;
  height: 11rpx;
  flex-shrink: 0;
  margin-left: 12rpx;
  border-right: 2rpx solid $brand;
  border-bottom: 2rpx solid $brand;
  transform: rotate(-45deg);
  opacity: 0.7;
}

.desc {
  margin-top: 14rpx;
  padding: 20rpx;
  background: $bg-subtle;
  border-radius: $radius;
  font-size: $font-mid;
  color: $text-sub;
  line-height: 1.7;
}

.shots {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 16rpx;
}

.shot {
  width: 150rpx;
  height: 150rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
}

.remark {
  margin-top: 16rpx;
  font-size: $font-caption;
  color: $text-weak;
  line-height: 1.6;
}

.actions {
  padding-top: 24rpx;
}

.btn {
  height: 80rpx;
  line-height: 80rpx;
  text-align: center;
  border-radius: $radius;
  background: $brand;
  color: $text-on-accent;
  font-size: $font-body;
  font-weight: 600;
  transition: transform 180ms cubic-bezier(0.16, 1, 0.3, 1), background-color 180ms ease;

  &:active {
    transform: scale(0.98);
    background: $brand-dark;
  }
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: $font-caption;
  padding: 30rpx 0;
}
</style>
