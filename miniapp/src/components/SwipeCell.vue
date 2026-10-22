<script setup lang="ts">
import { computed, ref } from 'vue'

/**
 * 左滑露出操作按钮的单元格。
 *
 * 三个容易被忽略但很影响手感的地方：
 * 1. 方向锁定 —— 首次判定出横向意图前，不抢纵向滚动，否则列表没法正常滚。
 * 2. 阻尼 —— 滑过按钮总宽后位移按 0.32 衰减，给出「到头了」的触感。
 * 3. 吸附 —— 松手时按位移是否过半决定展开还是收回，中间态不残留。
 *
 * 手势用 uni 的 touch 事件实现（小程序无原生 <div> 事件体系），
 * 所有临界参数集中在下面的常量里，便于统一调。
 */
const props = withDefaults(
  defineProps<{
    /** 是否禁用滑动 */
    disabled?: boolean
  }>(),
  { disabled: false }
)

const emit = defineEmits<{ action: [] }>()

/** 触发方向锁定的最小位移，小于它一律视为点击 */
const LOCK_THRESHOLD = 8
/** 展开后按钮露出的宽度，单位 px */
const ACTION_WIDTH = 88
/** 超出 ACTION_WIDTH 后的位移衰减系数 */
const RUBBER = 0.32

const offset = ref(0)
const animating = ref(true)
/** 手势是否已判定为横向，判定前不动，避免和纵向滚动打架 */
const horizontal = ref(false)

let startX = 0
let startY = 0
let startOffset = 0
let tracking = false

const actionStyle = computed(() => ({
  width: `${ACTION_WIDTH}px`,
}))

const bodyStyle = computed(() => ({
  transform: `translateX(${offset.value}px)`,
  transition: animating.value ? 'transform 260ms cubic-bezier(0.22, 0.61, 0.36, 1)' : 'none',
}))

function onTouchStart(e: any) {
  if (props.disabled) return
  const t = e.touches?.[0]
  if (!t) return
  tracking = true
  horizontal.value = false
  startX = t.clientX
  startY = t.clientY
  startOffset = offset.value
  animating.value = false
}

function onTouchMove(e: any) {
  if (!tracking) return
  const t = e.touches?.[0]
  if (!t) return

  const dx = t.clientX - startX
  const dy = t.clientY - startY

  if (!horizontal.value) {
    // 纵向优先：先动得比较多就判定为滚动，本轮不再接管
    if (Math.abs(dy) > Math.abs(dx) && Math.abs(dy) > LOCK_THRESHOLD) {
      tracking = false
      return
    }
    if (Math.abs(dx) < LOCK_THRESHOLD) return
    horizontal.value = true
  }

  let next = startOffset + dx
  if (next > 0) {
    // 右滑没有内容，用强阻尼压住，只留一点点跟随感
    next *= 0.16
  } else if (next < -ACTION_WIDTH) {
    next = -ACTION_WIDTH + (next + ACTION_WIDTH) * RUBBER
  }
  offset.value = next
}

function onTouchEnd() {
  if (!tracking) return
  tracking = false
  if (!horizontal.value) return

  animating.value = true
  // 过按钮宽度的一半就展开，否则收回；中间态不残留
  offset.value = offset.value < -ACTION_WIDTH / 2 ? -ACTION_WIDTH : 0
}

/** 收起。操作完成后外部调用，或用户点空白处 */
function close() {
  animating.value = true
  offset.value = 0
}

function trigger() {
  close()
  emit('action')
}

defineExpose({ close })
</script>

<template>
  <view class="swipe">
    <view class="swipe-action" :style="actionStyle">
      <view class="swipe-btn" @click.stop="trigger">
        <slot name="action">操作</slot>
      </view>
    </view>

    <view
      class="swipe-body"
      :style="bodyStyle"
      @touchstart.passive="onTouchStart"
      @touchmove.passive="onTouchMove"
      @touchend="onTouchEnd"
      @touchcancel="onTouchEnd"
      @click="offset !== 0 && close()"
    >
      <slot></slot>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.swipe {
  position: relative;
  overflow: hidden;
}

.swipe-action {
  position: absolute;
  top: 0;
  right: 0;
  height: 100%;
  display: flex;
  align-items: stretch;
}

.swipe-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $danger;
  color: $text-on-accent;
  font-size: $font-caption;
  font-weight: 600;
}

.swipe-body {
  position: relative;
  background: $bg-raised;
  will-change: transform;
}
</style>
