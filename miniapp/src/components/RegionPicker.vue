<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { REGIONS, findCityByCode, findIndexByName } from '@/utils/region'

/**
 * 省市两级选择器。
 *
 * 不用 <picker mode="region">：H5 端 uni-app 没有实现 region 模式，弹层是空的。
 * 统一走本地行政区划数据 + multiSelector，各端表现一致，编码也和微信的 region 选择器对得上。
 */
const props = withDefaults(
  defineProps<{
    /** 已选省名 */
    province?: string
    /** 已选市名 */
    city?: string
    /** 已选城市编码；只存了编码没存名字时（择偶要求、搜索筛选）用它回显 */
    code?: string
    placeholder?: string
  }>(),
  { province: '', city: '', code: '', placeholder: '请选择' },
)

const emit = defineEmits<{
  (e: 'change', v: { province: string; city: string; code: string }): void
}>()

const index = ref<[number, number]>([0, 0])

const range = computed(() => [
  REGIONS.map((p) => p.name),
  (REGIONS[index.value[0]]?.cities || []).map((c) => c.name),
])

/** 省市名优先，只有编码时反查 */
const text = computed(() => {
  if (props.province) return props.city ? `${props.province} ${props.city}` : props.province
  const hit = findCityByCode(props.code)
  return hit ? `${hit.province.name} ${hit.city.name}` : ''
})

// 外部值变了就把滚轮定位到当前选中项，再次打开时不会跳回北京
watch(
  () => [props.province, props.city, props.code],
  () => {
    if (props.province) {
      index.value = findIndexByName(props.province, props.city)
      return
    }
    const hit = findCityByCode(props.code)
    if (hit) {
      index.value = findIndexByName(hit.province.name, hit.city.name)
    }
  },
  { immediate: true },
)

function onColumnChange(e: any) {
  const { column, value } = e.detail
  // 换省份时市列表整体换掉，下标必须归零，否则会越界
  index.value = column === 0 ? [Number(value), 0] : [index.value[0], Number(value)]
}

function onChange(e: any) {
  const [pi, ci] = (e.detail.value as number[]).map(Number)
  const province = REGIONS[pi]
  const city = province?.cities[ci]
  if (!province || !city) return
  index.value = [pi, ci]
  emit('change', { province: province.name, city: city.name, code: city.code })
}
</script>

<template>
  <picker
    mode="multiSelector"
    :range="range"
    :value="index"
    @columnchange="onColumnChange"
    @change="onChange"
  >
    <slot :text="text">
      <text :class="{ ph: !text }">{{ text || placeholder }}</text>
    </slot>
  </picker>
</template>

<style lang="scss" scoped>
.ph {
  color: $text-placeholder;
}
</style>
