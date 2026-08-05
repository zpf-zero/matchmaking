<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import UserCardItem from '@/components/UserCardItem.vue'
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
  city_name: '',
  house_status: 0,
  car_status: 0,
})

// 记住上次的筛选条件：每次进来都重设一遍太烦
const FILTER_KEY = 'search_filters'
const saved = uni.getStorageSync(FILTER_KEY)
if (saved && typeof saved === 'object') {
  Object.assign(filters, saved)
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

function onRegion(e: any) {
  const [prov, city] = e.detail.value
  filters.city_name = `${prov} ${city}`
  filters.city_code = (e.detail.code || [])[1] || ''
}

function reset() {
  Object.assign(filters, {
    age_min: 0, age_max: 0, height_min: 0, height_max: 0,
    education_min: 0, income_min: 0, marital_status: '', has_children: '',
    city_code: '', city_name: '', house_status: 0, car_status: 0,
  })
  uni.removeStorageSync(FILTER_KEY)
}

async function doSearch(resetPage = true) {
  if (loading.value) return
  loading.value = true

  if (resetPage) {
    page.value = 1
    finished.value = false
    panelOpen.value = false
    uni.setStorageSync(FILTER_KEY, { ...filters })
  }

  const params: Record<string, any> = { page: page.value, page_size: 20 }
  Object.entries(filters).forEach(([k, v]) => {
    if (k === 'city_name') return
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
      <text class="panel-title">筛选条件</text>
      <text class="panel-toggle">{{ panelOpen ? '收起 ∧' : '展开 ∨' }}</text>
    </view>

    <view v-show="panelOpen" class="card">
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
        <picker mode="region" class="v ta-r" @change="onRegion">
          <text>{{ filters.city_name || '不限' }}</text>
        </picker>
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

    <UserCardItem v-for="u in list" :key="u.id" :user="u" />

    <view v-if="searched && !list.length" class="empty">
      没有符合条件的人，试试放宽一些条件
    </view>
    <view v-else-if="list.length && finished" class="end">已经到底啦</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 20rpx 24rpx 40rpx;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 8rpx 20rpx;
}

.panel-title {
  font-size: 30rpx;
  font-weight: 600;
}

.panel-toggle {
  font-size: 24rpx;
  color: #8a8f99;
}

.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 8rpx 24rpx 24rpx;
  margin-bottom: 24rpx;
}

.row-item {
  display: flex;
  align-items: center;
  min-height: 92rpx;
  border-bottom: 1rpx solid #f2f3f5;
  font-size: 28rpx;

  .k {
    width: 160rpx;
    color: #1f2329;
    flex-shrink: 0;
  }

  .v {
    flex: 1;
    color: #4e5969;
  }

  .ta-r {
    text-align: right;
  }
}

.pick {
  display: inline-block;
  min-width: 90rpx;
  text-align: center;
  color: #e8506d;
}

.dash {
  color: #bbbfc4;
  margin: 0 10rpx;
}

.panel-actions {
  display: flex;
  gap: 20rpx;
  padding-top: 28rpx;
}

.half {
  flex: 1;
}

.result-count {
  font-size: 26rpx;
  color: #8a8f99;
  padding: 8rpx 8rpx 20rpx;
}

.end {
  text-align: center;
  color: #bbbfc4;
  font-size: 24rpx;
  padding: 30rpx 0;
}
</style>
