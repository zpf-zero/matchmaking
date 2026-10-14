<script setup lang="ts">
import { computed, ref, watch } from 'vue'

/**
 * 头像。
 * 图片缺失或加载失败时不留灰块，改用品牌色的字形占位，
 * 保证列表在任何数据状态下都不出现“破图”。
 */
const props = withDefaults(
  defineProps<{
    src?: string
    name?: string
    /** 边长，单位 rpx */
    size?: number
    /** 铺满父容器（用于资料页头图这类大幅位） */
    fill?: boolean
    /** 圆角档位，跟全站形状锁一致 */
    radius?: 'none' | 'md' | 'lg' | 'round'
  }>(),
  { src: '', name: '', size: 96, fill: false, radius: 'md' }
)

const failed = ref(false)
watch(
  () => props.src,
  () => {
    failed.value = false
  }
)

const showImage = computed(() => !!props.src && !failed.value)

/** 中文取姓氏一个字，英文取首字母 */
const initial = computed(() => {
  const n = (props.name || '').trim()
  if (!n) return '缘'
  return /^[a-zA-Z]/.test(n) ? n[0].toUpperCase() : n[0]
})

const boxStyle = computed(() => ({
  ...(props.fill
    ? { width: '100%', height: '100%' }
    : { width: `${props.size}rpx`, height: `${props.size}rpx` }),
  fontSize: props.fill
    ? '110rpx'
    : `${Math.max(22, Math.round(props.size * 0.4))}rpx`,
}))
</script>

<template>
  <view class="avatar" :class="`r-${radius}`" :style="boxStyle">
    <image
      v-if="showImage"
      class="pic"
      :src="src"
      mode="aspectFill"
      lazy-load
      @error="failed = true"
    />
    <text v-else class="mark">{{ initial }}</text>
  </view>
</template>

<style lang="scss" scoped>
.avatar {
  position: relative;
  flex-shrink: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $brand-light;
  border: 1rpx solid $border;
}

.r-md {
  border-radius: $radius;
}

.r-none {
  border-radius: 0;
  border: 0;
}

.r-lg {
  border-radius: $radius-lg;
}

.r-round {
  border-radius: $radius-round;
}

.pic {
  width: 100%;
  height: 100%;
}

.mark {
  /* 占位字形跟随品牌色，与整站单一强调色保持一致 */
  color: $brand;
  font-weight: 600;
  line-height: 1;
  opacity: 0.72;
}
</style>
