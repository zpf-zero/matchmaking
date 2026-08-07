<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import FormRow from '@/components/FormRow.vue'
import { userApi } from '@/api'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const saving = ref(false)

const heightRange = Array.from({ length: 81 }, (_, i) => 140 + i)
const ageRange = Array.from({ length: 63 }, (_, i) => 18 + i)

const req = reactive({
  age_min: 22,
  age_max: 40,
  height_min: 150,
  height_max: 195,
  education_min: 0,
  income_min: 0,
  marital_status: '1',
  accept_children: 1,
  city_codes: '',
  other: '',
})

const acceptMarital = reactive<Record<number, boolean>>({ 1: true, 2: false, 3: false })
const cityName = ref('')
const CITY_NAME_KEY = 'requirement_city_name'

onLoad(async () => {
  await store.fetchDict()
  if (!store.requirement) await store.fetchMe()

  const r = store.requirement
  if (r) {
    Object.assign(req, r)
    Object.keys(acceptMarital).forEach((k) => (acceptMarital[+k] = false))
    ;(r.marital_status || '').split(',').forEach((v) => {
      if (v) acceptMarital[+v] = true
    })
    const savedCity = uni.getStorageSync(CITY_NAME_KEY)
    if (savedCity?.code === r.city_codes) cityName.value = savedCity.name || ''
  }
})

function pickNum(range: number[], field: 'age_min' | 'age_max' | 'height_min' | 'height_max') {
  uni.showActionSheet({
    itemList: range.map(String),
    success: (res) => (req[field] = range[res.tapIndex]),
  })
}

function pickDict(type: string, field: 'education_min' | 'income_min') {
  const list = [{ value: 0, label: '不限' }, ...(store.dict[type] || [])]
  uni.showActionSheet({
    itemList: list.map((i) => i.label),
    success: (res) => (req[field] = list[res.tapIndex].value),
  })
}

function onRegion(e: any) {
  const [prov, city] = e.detail.value
  cityName.value = `${prov} ${city}`
  req.city_codes = (e.detail.code || [])[1] || ''
}

function clearCity() {
  cityName.value = ''
  req.city_codes = ''
}

async function save() {
  if (req.age_min > req.age_max) {
    uni.showToast({ title: '年龄范围填反了', icon: 'none' })
    return
  }
  if (req.height_min > req.height_max) {
    uni.showToast({ title: '身高范围填反了', icon: 'none' })
    return
  }

  saving.value = true
  req.marital_status = Object.keys(acceptMarital)
    .filter((k) => acceptMarital[+k])
    .join(',')

  try {
    await userApi.saveRequirement(req)
    if (req.city_codes && cityName.value) {
      uni.setStorageSync(CITY_NAME_KEY, { code: req.city_codes, name: cityName.value })
    } else {
      uni.removeStorageSync(CITY_NAME_KEY)
    }
    await store.fetchMe()
    uni.showToast({ title: '已保存' })
    setTimeout(() => uni.navigateBack(), 700)
  } catch {
    // 已提示
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <view class="page">
    <view class="tip">择偶要求会用来筛选推荐给你的人，也会影响你和对方的匹配度</view>

    <view class="card">
      <FormRow label="年龄范围" align-right>
        <text class="pick" @click="pickNum(ageRange, 'age_min')">{{ req.age_min }}</text>
        <text class="dash">-</text>
        <text class="pick" @click="pickNum(ageRange, 'age_max')">{{ req.age_max }}</text>
        <text class="unit">岁</text>
      </FormRow>

      <FormRow label="身高范围" align-right>
        <text class="pick" @click="pickNum(heightRange, 'height_min')">{{ req.height_min }}</text>
        <text class="dash">-</text>
        <text class="pick" @click="pickNum(heightRange, 'height_max')">{{ req.height_max }}</text>
        <text class="unit">cm</text>
      </FormRow>

      <FormRow label="学历要求" align-right>
        <text @click="pickDict('education', 'education_min')">
          {{ req.education_min ? store.label('education', req.education_min) + '及以上' : '不限' }}
        </text>
      </FormRow>

      <FormRow label="收入要求" align-right>
        <text @click="pickDict('income_level', 'income_min')">
          {{ req.income_min ? store.label('income_level', req.income_min) + '及以上' : '不限' }}
        </text>
      </FormRow>

      <FormRow label="可接受婚史" align-right>
        <view class="multi">
          <view
            v-for="m in store.dict['marital_status'] || []"
            :key="m.value"
            class="multi-item"
            :class="{ on: acceptMarital[m.value] }"
            @click="acceptMarital[m.value] = !acceptMarital[m.value]"
          >
            {{ m.label }}
          </view>
        </view>
      </FormRow>

      <FormRow label="接受有子女" align-right>
        <switch
          :checked="req.accept_children === 1"
          color="#e8506d"
          @change="req.accept_children = $event.detail.value ? 1 : 0"
        />
      </FormRow>

      <FormRow label="期望城市" align-right>
        <picker mode="region" @change="onRegion">
          <text>{{ cityName || (req.city_codes ? '已设置' : '不限') }}</text>
        </picker>
        <text v-if="req.city_codes" class="clear" @click="clearCity">清除</text>
      </FormRow>
    </view>

    <view class="card">
      <view class="label">其他要求</view>
      <textarea
        v-model="req.other"
        class="other"
        maxlength="200"
        placeholder="还有什么特别看重的？例如 希望对方也喜欢运动"
        placeholder-class="ph"
      />
    </view>

    <view class="actions">
      <view class="btn-primary" :class="{ disabled: saving }" @click="save">保存</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 24rpx 24rpx 180rpx;
}

.tip {
  font-size: 24rpx;
  color: #8a8f99;
  line-height: 1.6;
  padding: 0 8rpx 20rpx;
}

.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 8rpx 24rpx;
  margin-bottom: 20rpx;
}

.label {
  font-size: 28rpx;
  padding: 24rpx 0 12rpx;
}

.other {
  width: 100%;
  height: 160rpx;
  font-size: 27rpx;
  line-height: 1.6;
  padding-bottom: 20rpx;
}

.ph {
  color: #bbbfc4;
}

.pick {
  display: inline-block;
  min-width: 80rpx;
  text-align: center;
  color: #e8506d;
}

.dash,
.unit {
  color: #8a8f99;
  margin: 0 6rpx;
}

.clear {
  margin-left: 16rpx;
  font-size: 24rpx;
  color: #bbbfc4;
}

.multi {
  display: flex;
  justify-content: flex-end;
  gap: 12rpx;
}

.multi-item {
  padding: 8rpx 20rpx;
  border-radius: 24rpx;
  background: #f5f6f8;
  color: #646a73;
  font-size: 24rpx;

  &.on {
    background: #fff0f3;
    color: #e8506d;
  }
}

.actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
}
</style>
