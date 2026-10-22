<script setup lang="ts">
import { computed } from 'vue'

/**
 * 单列滚轮选择器。
 *
 * 为什么不用 uni.showActionSheet：
 * 它是「弹出列表」，不是「滚轮」。微信小程序端 itemList 上限 6 项，
 * 年龄（63 项）、身高（81 项）这类长范围会被截断甚至直接失败；
 * 就算在 H5 端能滚动，选个年龄也要划半天。
 * 统一走原生 <picker mode="selector">，两端都是系统级滚轮，手感一致。
 *
 * 用法：
 *   <RangePicker :options="opts" :model-value="v" @update:model-value="v = $event">
 *     <template #default="{ text }">{{ text || '请选择' }}</template>
 *   </RangePicker>
 */
export interface Option {
  value: number
  label: string
}

const props = withDefaults(
  defineProps<{
    /** 选项列表，要么传 {value,label}，要么直接传数字数组 */
    options: Option[] | number[]
    modelValue: number
    /** 显示后缀，如「岁」「cm」，只影响回显文案 */
    unit?: string
    /** 空值时的回显，传 0 时会用它 */
    placeholder?: string
    /** 数字数组模式下的显示后缀（老调用方传的是纯数字） */
    numericSuffix?: string
  }>(),
  { unit: '', placeholder: '请选择', numericSuffix: '' }
)

const emit = defineEmits<{
  (e: 'update:modelValue', v: number): void
}>()

/** 归一成 {value,label}，兼容直接传 number[] 的调用方 */
const normalized = computed<Option[]>(() =>
  props.options.map((o) =>
    typeof o === 'number' ? { value: o, label: `${o}${props.numericSuffix}` } : o
  )
)

/** picker 的 range 只认字符串或对象数组的 label 字段，这里显式给 string[] */
const labels = computed(() => normalized.value.map((o) => o.label))

/** 挡位下标：找不到时退回 0，避免 picker 收到越界的 value 后白屏 */
const index = computed(() => {
  const i = normalized.value.findIndex((o) => o.value === props.modelValue)
  return i >= 0 ? i : 0
})

const text = computed(() => {
  // 注意不能用 !props.modelValue 判空：0 是合法值（「不限」这类选项就存在 0），
  // 按真假判断会把选中「不限」的用户显示成空
  const hit = normalized.value.find((o) => o.value === props.modelValue)
  if (!hit) return ''
  return props.unit ? `${hit.label}${props.unit}` : hit.label
})

function onChange(e: any) {
  const i = Number(e.detail.value)
  const hit = normalized.value[i]
  if (!hit) return
  emit('update:modelValue', hit.value)
}
</script>

<template>
  <!-- 外面套一层行内块：H5 端 uni-picker 渲染成块级元素，
       直接放在非 flex 容器里（如 FormRow 的右对齐区）会各占一行、竖着堆叠。
       用 inline-block 包一层后，同行多个 picker 才能并排。 -->
  <view class="rp">
    <picker mode="selector" :range="labels" :value="index" @change="onChange">
      <slot :text="text" :placeholder="placeholder">
        <text class="rp-text" :class="{ ph: !text }">{{ text || placeholder }}</text>
      </slot>
    </picker>
  </view>
</template>

<style lang="scss" scoped>
.rp {
  display: inline-block;
  vertical-align: middle;
}

/* 默认外观：未选中走占位色，选中走品牌色，与全站 picker 类控件一致。
   调用方要自定义时用作用域插槽覆盖即可。 */
.rp-text {
  color: $brand;
  font-variant-numeric: tabular-nums;
}

.ph {
  color: $text-placeholder;
}
</style>
