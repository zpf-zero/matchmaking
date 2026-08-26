<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import FormRow from '@/components/FormRow.vue'
import RegionPicker from '@/components/RegionPicker.vue'
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

function onRegion(v: { code: string }) {
  req.city_codes = v.code
}

function clearCity() {
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
          color="#963E58"
          @change="req.accept_children = $event.detail.value ? 1 : 0"
        />
      </FormRow>

      <FormRow label="期望城市" align-right>
        <RegionPicker :code="req.city_codes" placeholder="不限" @change="onRegion" />
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
  padding: 28rpx 24rpx 196rpx;
}

.tip {
  font-size: 24rpx;
  color: $text-weak;
  line-height: 1.6;
  padding: 0 8rpx 20rpx;
}

.card {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 8rpx 28rpx;
  margin-bottom: 24rpx;
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
  color: $text-placeholder;
}

.pick {
  display: inline-block;
  min-width: 80rpx;
  text-align: center;
  color: $brand;
}

.dash,
.unit {
  color: $text-weak;
  margin: 0 6rpx;
}

.clear {
  margin-left: 16rpx;
  font-size: 24rpx;
  color: $text-placeholder;
}

.multi {
  display: flex;
  justify-content: flex-end;
  gap: 12rpx;
}

.multi-item {
  padding: 8rpx 20rpx;
  border-radius: $radius-sm;
  border: 1rpx solid $border;
  background: $bg-subtle;
  color: $text-sub;
  font-size: 24rpx;

  &.on {
    background: $brand-light;
    border-color: $brand-light;
    color: $brand;
  }
}

.actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: $bg-card;
  border-top: 1rpx solid $border;
}
</style>
