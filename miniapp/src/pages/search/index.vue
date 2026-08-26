<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import UserCardItem from '@/components/UserCardItem.vue'
import RegionPicker from '@/components/RegionPicker.vue'
import { discoverApi } from '@/api'
import type { UserCard } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const list = ref<UserCard[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const finished = ref(false)
const searched = ref(false)
const panelOpen = ref(true)

const heightRange = [0, ...Array.from({ length: 81 }, (_, i) => 140 + i)]
const ageRange = [0, ...Array.from({ length: 63 }, (_, i) => 18 + i)]

const filters = reactive({
  age_min: 0,
  age_max: 0,
  height_min: 0,
  height_max: 0,
  education_min: 0,
  income_min: 0,
  marital_status: '',
  has_children: '',
  city_code: '',
  house_status: 0,
  car_status: 0,
})

// 记住上次的筛选条件：每次进来都重设一遍太烦
const FILTER_KEY = 'search_filters'
const saved = uni.getStorageSync(FILTER_KEY)
if (saved && typeof saved === 'object') {
  // 只认当前还在用的字段，免得历史版本存下的键被原样发给接口
  Object.keys(filters).forEach((k) => {
    if (saved[k] !== undefined) (filters as any)[k] = saved[k]
  })
}

onShow(() => {
  store.fetchDict()
})

function pickNumber(range: number[], field: 'age_min' | 'age_max' | 'height_min' | 'height_max') {
  uni.showActionSheet({
    itemList: range.map((v) => (v === 0 ? '不限' : String(v))),
    success: (res) => (filters[field] = range[res.tapIndex]),
  })
}

function pickDict(type: string, field: 'education_min' | 'income_min' | 'house_status' | 'car_status') {
  const list = [{ value: 0, label: '不限' }, ...(store.dict[type] || [])]
  uni.showActionSheet({
    itemList: list.map((i) => i.label),
    success: (res) => (filters[field] = list[res.tapIndex].value),
  })
}

function pickMarital() {
  const items = [{ value: '', label: '不限' }, ...(store.dict['marital_status'] || []).map((i) => ({ value: String(i.value), label: i.label }))]
  uni.showActionSheet({
    itemList: items.map((i) => i.label),
    success: (res) => (filters.marital_status = items[res.tapIndex].value),
  })
}

function pickChildren() {
  const items = [{ value: '', label: '不限' }, ...(store.dict['has_children'] || []).map((i) => ({ value: String(i.value), label: i.label }))]
  uni.showActionSheet({
    itemList: items.map((i) => i.label),
    success: (res) => (filters.has_children = items[res.tapIndex].value),
  })
}

function onRegion(v: { code: string }) {
  filters.city_code = v.code
}

function reset() {
  Object.assign(filters, {
    age_min: 0, age_max: 0, height_min: 0, height_max: 0,
    education_min: 0, income_min: 0, marital_status: '', has_children: '',
    city_code: '', house_status: 0, car_status: 0,
  })
  uni.removeStorageSync(FILTER_KEY)
}

async function doSearch(resetPage = true) {
  if (loading.value) return
  loading.value = true

  if (resetPage) {
    page.value = 1
    finished.value = false
    searched.value = false
    list.value = []
    panelOpen.value = false
    uni.setStorageSync(FILTER_KEY, { ...filters })
  }

  const params: Record<string, any> = { page: page.value, page_size: 20 }
  Object.entries(filters).forEach(([k, v]) => {
    if (v !== 0 && v !== '') params[k] = v
  })

  try {
    const res = await discoverApi.search(params)
    total.value = res.total
    list.value = resetPage ? res.list || [] : [...list.value, ...(res.list || [])]
    finished.value = list.value.length >= res.total
    if (!finished.value) page.value++
    searched.value = true
  } catch {
    // 已提示
  } finally {
    loading.value = false
  }
}

onReachBottom(() => {
  if (searched.value && !finished.value) doSearch(false)
})

function labelOr(type: string, value: number, fallback = '不限') {
  return value ? store.label(type, value) : fallback
}
</script>

<template>
  <view class="page">
    <view class="panel-head" @click="panelOpen = !panelOpen">
      <view class="panel-copy">
        <text class="panel-title">筛选条件</text>
        <text class="panel-note">条件会自动保存</text>
      </view>
      <view class="panel-toggle">
        <text>{{ panelOpen ? '收起' : '展开' }}</text>
        <view class="chevron" :class="{ open: panelOpen }"></view>
      </view>
    </view>

    <view v-show="panelOpen" class="card">
      <view class="filter-group-title">基本条件</view>
      <view class="row-item">
        <text class="k">年龄</text>
        <view class="v">
          <text class="pick" @click="pickNumber(ageRange, 'age_min')">
            {{ filters.age_min || '不限' }}
          </text>
          <text class="dash">-</text>
          <text class="pick" @click="pickNumber(ageRange, 'age_max')">
            {{ filters.age_max || '不限' }}
          </text>
        </view>
      </view>

      <view class="row-item">
        <text class="k">身高</text>
        <view class="v">
          <text class="pick" @click="pickNumber(heightRange, 'height_min')">
            {{ filters.height_min || '不限' }}
          </text>
          <text class="dash">-</text>
          <text class="pick" @click="pickNumber(heightRange, 'height_max')">
            {{ filters.height_max || '不限' }}
          </text>
        </view>
      </view>

      <view class="row-item" @click="pickDict('education', 'education_min')">
        <text class="k">学历</text>
        <text class="v ta-r">{{ labelOr('education', filters.education_min) }}</text>
      </view>

      <view class="row-item" @click="pickDict('income_level', 'income_min')">
        <text class="k">月收入</text>
        <text class="v ta-r">{{ labelOr('income_level', filters.income_min) }}</text>
      </view>

      <view class="filter-group-title second">生活情况</view>
      <view class="row-item" @click="pickMarital">
        <text class="k">婚姻状况</text>
        <text class="v ta-r">
          {{ filters.marital_status ? store.label('marital_status', +filters.marital_status) : '不限' }}
        </text>
      </view>

      <view class="row-item" @click="pickChildren">
        <text class="k">子女情况</text>
        <text class="v ta-r">
          {{ filters.has_children !== '' ? store.label('has_children', +filters.has_children) : '不限' }}
        </text>
      </view>

      <view class="row-item">
        <text class="k">所在城市</text>
        <RegionPicker v-slot="{ text }" class="v ta-r" :code="filters.city_code" @change="onRegion">
          <text>{{ text || '不限' }}</text>
        </RegionPicker>
      </view>

      <view class="row-item" @click="pickDict('house_status', 'house_status')">
        <text class="k">购房情况</text>
        <text class="v ta-r">{{ labelOr('house_status', filters.house_status) }}</text>
      </view>

      <view class="row-item" @click="pickDict('car_status', 'car_status')">
        <text class="k">购车情况</text>
        <text class="v ta-r">{{ labelOr('car_status', filters.car_status) }}</text>
      </view>

      <view class="panel-actions">
        <view class="btn-plain half" @click="reset">重置</view>
        <view class="btn-primary half" @click="doSearch(true)">查找</view>
      </view>
    </view>

    <view v-if="searched" class="result-count">共找到 {{ total }} 人</view>

    <view v-if="loading && !searched" class="result-skeleton" aria-label="正在查找">
      <view v-for="n in 2" :key="n" class="skeleton-result-card">
        <view class="skeleton skeleton-result-avatar"></view>
        <view class="skeleton-result-copy">
          <view class="skeleton skeleton-result-line strong"></view>
          <view class="skeleton skeleton-result-line"></view>
          <view class="skeleton skeleton-result-line short"></view>
        </view>
      </view>
    </view>

    <UserCardItem v-for="u in list" :key="u.id" :user="u" />

    <view v-if="searched && !list.length" class="empty">
      没有符合条件的人，试试放宽一些条件
    </view>
    <view v-else-if="list.length && finished" class="end">已经到底啦</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  min-height: calc(100dvh - 188rpx);
  padding: 24rpx 24rpx 48rpx;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 4rpx 22rpx;
}

.panel-copy {
  display: flex;
  flex-direction: column;
  gap: 5rpx;
}

.panel-title {
  font-size: 35rpx;
  font-weight: 650;
  letter-spacing: -0.5rpx;
}

.panel-note {
  font-size: 22rpx;
  color: $text-placeholder;
}

.panel-toggle {
  display: flex;
  align-items: center;
  gap: 10rpx;
  font-size: 24rpx;
  color: $brand;
  font-weight: 600;
}

.chevron {
  width: 12rpx;
  height: 12rpx;
  border-right: 2rpx solid currentColor;
  border-bottom: 2rpx solid currentColor;
  transform: rotate(45deg) translateY(-2rpx);
  transition: transform 180ms ease;

  &.open {
    transform: rotate(225deg) translate(-2rpx, -2rpx);
  }
}

.card {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 0 28rpx 28rpx;
  margin-bottom: 24rpx;
  box-shadow: $shadow-card;
}

.filter-group-title {
  padding: 26rpx 0 8rpx;
  font-size: 23rpx;
  color: $text-weak;
  font-weight: 600;

  &.second {
    padding-top: 30rpx;
  }
}

.row-item {
  display: flex;
  align-items: center;
  min-height: 94rpx;
  border-bottom: 1rpx solid $border;
  font-size: 28rpx;

  .k {
    width: 160rpx;
    color: $text-main;
    font-weight: 500;
    flex-shrink: 0;
  }

  .v {
    flex: 1;
    color: $text-sub;
  }

  .ta-r {
    text-align: right;
  }
}

.pick {
  display: inline-block;
  min-width: 96rpx;
  text-align: center;
  color: $brand;
  background: $brand-light;
  border-radius: $radius-sm;
  padding: 6rpx 10rpx;
  font-variant-numeric: tabular-nums;
}

.dash {
  color: $text-placeholder;
  margin: 0 10rpx;
}

.panel-actions {
  display: flex;
  gap: 20rpx;
  padding-top: 30rpx;
}

.half {
  flex: 1;
}

.result-count {
  font-size: 26rpx;
  color: $text-weak;
  padding: 10rpx 4rpx 20rpx;
  font-variant-numeric: tabular-nums;
}

.result-skeleton {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
  margin-bottom: 24rpx;
}

.skeleton-result-card {
  display: grid;
  grid-template-columns: 202rpx 1fr;
  gap: 26rpx;
  padding: 18rpx;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  background: $bg-raised;
  box-shadow: $shadow-card;
}

.skeleton-result-avatar {
  width: 202rpx;
  height: 264rpx;
  border-radius: $radius;
}

.skeleton-result-copy {
  padding-top: 14rpx;
}

.skeleton-result-line {
  height: 22rpx;
  border-radius: $radius-sm;
  margin-bottom: 24rpx;

  &.strong {
    width: 76%;
    height: 30rpx;
  }

  &.short {
    width: 52%;
  }
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: 24rpx;
  padding: 36rpx 0;
}
</style>
