<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import UserCardItem from '@/components/UserCardItem.vue'
import AppTabBar from '@/components/AppTabBar.vue'
import RegionPicker from '@/components/RegionPicker.vue'
import RangePicker from '@/components/RangePicker.vue'
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

// 首项是 0 表示「不限」。做成 {value,label} 让滚轮里直接显示「不限」两个字，
// 比原来滚出一串 0 更清楚
const heightOptions = [
  { value: 0, label: '不限' },
  ...Array.from({ length: 81 }, (_, i) => {
    const v = 140 + i
    return { value: v, label: `${v} cm` }
  }),
]
const ageOptions = [
  { value: 0, label: '不限' },
  ...Array.from({ length: 63 }, (_, i) => {
    const v = 18 + i
    return { value: v, label: `${v} 岁` }
  }),
]

/**
 * 字典选项补一个「不限」。
 * 用 computed 而不是普通函数：字典是异步拉取的，函数调用拿不到更新，
 * 会出现「字典回来了但滚轮里还是只有不限」。
 */
const educationOptions = computed(() => [
  { value: 0, label: '不限' },
  ...(store.dict['education'] || []),
])

const incomeOptions = computed(() => [
  { value: 0, label: '不限' },
  ...(store.dict['income_level'] || []),
])

const houseOptions = computed(() => [
  { value: 0, label: '不限' },
  ...(store.dict['house_status'] || []),
])

const carOptions = computed(() => [
  { value: 0, label: '不限' },
  ...(store.dict['car_status'] || []),
])

// 婚史和子女在筛选里是字符串空值表示「不限」，而别处用的是数字。
// 这里用负号当哨兵值：0 会被 doSearch 那层当「未填」丢掉，负号不会，
// 也不会和字典里真实存在的 0/1/2/3 撞车。
const maritalOptions = computed(() => [
  { value: -1, label: '不限' },
  ...(store.dict['marital_status'] || []),
])

const childrenOptions = computed(() => [
  { value: -1, label: '不限' },
  ...(store.dict['has_children'] || []),
])

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
  // 底部导航改为自绘（「消息」项受私聊开关控制），原生那条全程藏着
  uni.hideTabBar({ animation: false, fail: () => {} })
  store.fetchDict()
})

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
      <view class="group-title">基本条件</view>
      <view class="row-item">
        <text class="k">年龄</text>
        <view class="v">
          <RangePicker v-model="filters.age_min" :options="ageOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.age_min }">{{ text }}</text>
            </template>
          </RangePicker>
          <text class="dash">-</text>
          <RangePicker v-model="filters.age_max" :options="ageOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.age_max }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
      </view>

      <view class="row-item">
        <text class="k">身高</text>
        <view class="v">
          <RangePicker v-model="filters.height_min" :options="heightOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.height_min }">{{ text }}</text>
            </template>
          </RangePicker>
          <text class="dash">-</text>
          <RangePicker v-model="filters.height_max" :options="heightOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.height_max }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
      </view>

      <view class="row-item">
        <text class="k">学历</text>
        <view class="v">
          <RangePicker v-model="filters.education_min" :options="educationOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.education_min }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
        <view class="arrow"></view>
      </view>

      <view class="row-item">
        <text class="k">月收入</text>
        <view class="v">
          <RangePicker v-model="filters.income_min" :options="incomeOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.income_min }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
        <view class="arrow"></view>
      </view>

      <view class="group-divider"></view>
      <view class="group-title">生活情况</view>

      <view class="row-item">
        <text class="k">婚姻状况</text>
        <view class="v">
          <RangePicker
            :model-value="filters.marital_status === '' ? -1 : Number(filters.marital_status)"
            :options="maritalOptions"
            @update:model-value="filters.marital_status = $event < 0 ? '' : String($event)"
          >
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.marital_status !== '' }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
        <view class="arrow"></view>
      </view>

      <view class="row-item">
        <text class="k">子女情况</text>
        <view class="v">
          <RangePicker
            :model-value="filters.has_children === '' ? -1 : Number(filters.has_children)"
            :options="childrenOptions"
            @update:model-value="filters.has_children = $event < 0 ? '' : String($event)"
          >
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.has_children !== '' }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
        <view class="arrow"></view>
      </view>

      <view class="row-item">
        <text class="k">所在城市</text>
        <view class="v">
          <RegionPicker v-slot="{ text }" :code="filters.city_code" @change="onRegion">
            <text class="val" :class="{ on: !!filters.city_code }">{{ text || '不限' }}</text>
          </RegionPicker>
        </view>
        <view class="arrow"></view>
      </view>

      <view class="row-item">
        <text class="k">购房情况</text>
        <view class="v">
          <RangePicker v-model="filters.house_status" :options="houseOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.house_status }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
        <view class="arrow"></view>
      </view>

      <view class="row-item">
        <text class="k">购车情况</text>
        <view class="v">
          <RangePicker v-model="filters.car_status" :options="carOptions">
            <template #default="{ text }">
              <text class="val" :class="{ on: filters.car_status }">{{ text }}</text>
            </template>
          </RangePicker>
        </view>
        <view class="arrow"></view>
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

  <!-- 放在 .page 之外：.page 有入场动画（终态保留 transform），
       会成为 fixed 的定位祖先，把它沉到内容末尾 -->
  <AppTabBar class="tabbar-host" active="search" />
</template>

<style lang="scss" scoped>
.page {
  min-height: calc(100dvh - 188rpx);
  /* 底部 100rpx 自绘 tabBar + 安全区，避免最后一张卡片被挡住 */
  padding: 24rpx 24rpx calc(48rpx + 100rpx + env(safe-area-inset-bottom));
}

/* 宿主节点钉在视口底部（必须放在 .page 外面，否则被入场动画的 transform 劫持定位） */
.tabbar-host {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 200;
  /* 宽屏 H5 上跟随 .page 的 540px 单列画布，不拉成通栏 */
  max-width: 540px;
  margin-left: auto;
  margin-right: auto;
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
  font-size: $font-section;
  font-weight: 650;
  letter-spacing: -0.5rpx;
}

.panel-note {
  font-size: $font-caption;
  color: $text-placeholder;
}

.panel-toggle {
  display: flex;
  align-items: center;
  gap: 10rpx;
  font-size: $font-caption;
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

.group-title {
  padding: 28rpx 0 6rpx;
  font-size: $font-caption;
  letter-spacing: 1rpx;
  color: $text-weak;
  font-weight: 600;
}

/* 组间只用一条分隔线，组内靠留白分节 */
.group-divider {
  height: 1rpx;
  background: $border;
  margin-top: 18rpx;
}

.row-item {
  display: flex;
  align-items: center;
  min-height: 96rpx;
  font-size: $font-body;

  .k {
    width: 168rpx;
    flex-shrink: 0;
    color: $text-main;
    font-weight: 500;
  }
}

.v {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8rpx;
}

/* 未选是中性灰，已选落到品牌色，一眼看出哪几项在生效 */
.val {
  font-size: $font-body;
  color: $text-placeholder;
  font-variant-numeric: tabular-nums;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  &.on {
    color: $brand;
    font-weight: 600;
  }
}

.arrow {
  width: 12rpx;
  height: 12rpx;
  flex-shrink: 0;
  margin-left: 16rpx;
  border-right: 2rpx solid $text-placeholder;
  border-bottom: 2rpx solid $text-placeholder;
  transform: rotate(-45deg);
  opacity: 0.6;
}

.dash {
  color: $text-placeholder;
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
  font-size: $font-mid;
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
  grid-template-columns: 224rpx 1fr;
  gap: 24rpx;
  padding: 18rpx;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  background: $bg-raised;
  box-shadow: $shadow-card;
}

.skeleton-result-avatar {
  width: 224rpx;
  height: 288rpx;
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
  font-size: $font-caption;
  padding: 36rpx 0;
}
</style>
