<script setup lang="ts">
import { reactive, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import FormRow from '@/components/FormRow.vue'
import RegionPicker from '@/components/RegionPicker.vue'
import RangePicker from '@/components/RangePicker.vue'
import { userApi } from '@/api'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const saving = ref(false)

const heightRange = Array.from({ length: 81 }, (_, i) => 140 + i)
const ageRange = Array.from({ length: 63 }, (_, i) => 18 + i)

/** 字典选项补一个「不限」，用于学历/收入要求 */
const dictOr = (type: string) => [
  { value: 0, label: '不限' },
  ...(store.dict[type] || []),
]

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
        <RangePicker v-model="req.age_min" :options="ageRange" :numeric-suffix="''">
          <template #default="{ text, placeholder }">
            <text class="pick" :class="{ ph: !text }">{{ text || placeholder }}</text>
          </template>
        </RangePicker>
        <text class="dash">-</text>
        <RangePicker v-model="req.age_max" :options="ageRange" :numeric-suffix="''">
          <template #default="{ text, placeholder }">
            <text class="pick" :class="{ ph: !text }">{{ text || placeholder }}</text>
          </template>
        </RangePicker>
        <text class="unit">岁</text>
      </FormRow>

      <FormRow label="身高范围" align-right>
        <RangePicker v-model="req.height_min" :options="heightRange" :numeric-suffix="''">
          <template #default="{ text, placeholder }">
            <text class="pick" :class="{ ph: !text }">{{ text || placeholder }}</text>
          </template>
        </RangePicker>
        <text class="dash">-</text>
        <RangePicker v-model="req.height_max" :options="heightRange" :numeric-suffix="''">
          <template #default="{ text, placeholder }">
            <text class="pick" :class="{ ph: !text }">{{ text || placeholder }}</text>
          </template>
        </RangePicker>
        <text class="unit">cm</text>
      </FormRow>

      <FormRow label="学历要求" align-right>
        <RangePicker v-model="req.education_min" :options="dictOr('education')">
          <template #default="{ text, placeholder }">
            <text :class="{ ph: !req.education_min }">
              {{ req.education_min ? text + '及以上' : placeholder }}
            </text>
          </template>
        </RangePicker>
      </FormRow>

      <FormRow label="收入要求" align-right>
        <RangePicker v-model="req.income_min" :options="dictOr('income_level')">
          <template #default="{ text, placeholder }">
            <text :class="{ ph: !req.income_min }">
              {{ req.income_min ? text + '及以上' : placeholder }}
            </text>
          </template>
        </RangePicker>
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
  font-size: $font-caption;
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
  font-size: $font-body;
  padding: 24rpx 0 12rpx;
}

.other {
  width: 100%;
  height: 160rpx;
  font-size: $font-mid;
  line-height: 1.6;
  padding-bottom: 20rpx;
}

.ph {
  color: $text-placeholder;
}

/* .pick.ph 用复合选择器，保证占位色能盖住 .pick 的品牌色
   （两者同为单类时靠源码顺序决胜，很容易被后续改动破坏） */
.pick {
  display: inline-block;
  min-width: 72rpx;
  text-align: center;
  color: $brand;
  font-variant-numeric: tabular-nums;

  &.ph {
    color: $text-placeholder;
  }
}

/* 年龄/身高那行要塞 4 个 picker + 分隔符 + 单位，比别处更挤 */
.dash,
.unit {
  color: $text-weak;
  margin: 0 4rpx;
}

.clear {
  margin-left: 16rpx;
  font-size: $font-caption;
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
  font-size: $font-caption;

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
