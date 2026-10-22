<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'

/**
 * 心形喜欢按钮。
 *
 * 心形用「两个圆 + 旋转 45° 的方块」拼出来，不依赖 SVG、图片或字体图标 ——
 * 小程序端没有 <svg>，图片又要多一次请求，纯结构绘制的成本最低、双端表现一致。
 *
 * 点击时三层反馈叠在一起：心形弹性缩放、同色圆环扩散、六向粒子飞散。
 * 「喜欢」是这类产品最高频也最需要正反馈的动作，值得给足手感。
 */
const props = withDefaults(
  defineProps<{
    liked?: boolean
    /** 心形图标占据的方框边长，单位 rpx */
    size?: number
    /** pill 带文字标签 / icon 仅图标 */
    variant?: 'pill' | 'icon'
    label?: string
    activeLabel?: string
    disabled?: boolean
    /** pill 形态下撑满父容器 */
    block?: boolean
  }>(),
  {
    liked: false,
    size: 40,
    variant: 'pill',
    label: '喜欢',
    activeLabel: '已喜欢',
    disabled: false,
    block: false,
  }
)

const emit = defineEmits<{ like: [] }>()

const bursting = ref(false)
let burstTimer: ReturnType<typeof setTimeout> | null = null

/**
 * 心形还原成方块后的边长系数，以及纵向补偿量。
 *
 * 方块边长 S、绕中心旋转 -45° 时，两个圆心的落点是 (±0.354S, -0.354S)、半径 0.5S，
 * 方块的四个角点落到 (0, ±0.707S) 与 (±0.707S, 0)。取并集后：
 *   宽度 = 2 × 0.854S = 1.707S   （由两个圆的外缘决定）
 *   高度 = 0.854S + 0.707S = 1.561S，且包围盒中心比方块中心高 0.0735S
 * 宽度是约束项，所以 S 取 size / 1.707 ≈ 0.586 × size；
 * 纵向再补 0.0735S 把偏上的视觉重心压回方框正中。
 */
const squareSize = computed(() => Math.max(12, Math.round(props.size * 0.586)))

const stageStyle = computed(() => ({
  width: `${props.size}rpx`,
  height: `${props.size}rpx`,
}))

const shapeStyle = computed(() => {
  const s = squareSize.value
  return {
    width: `${s}rpx`,
    height: `${s}rpx`,
    marginLeft: `${-s / 2}rpx`,
    marginTop: `${-s / 2 + s * 0.0735}rpx`,
  }
})

const labelText = computed(() => (props.liked ? props.activeLabel : props.label))

/**
 * 六个方向的飞散粒子。
 * 把 rotate 放在外层包裹元素上、位移放在共用的 keyframes 里，
 * 这样六个粒子只需要一条动画定义 —— 否则要写六份内容相同的 keyframes。
 */
const angles = [0, 60, 120, 180, 240, 300]

function onTap(e?: any) {
  // 这个组件经常嵌在整块可点击的卡片里，必须截断冒泡，
  // 否则点「喜欢」会连带触发卡片本身的跳转
  e?.stopPropagation?.()
  if (props.disabled || props.liked) return
  playBurst()
  vibrate()
  emit('like')
}

function playBurst() {
  clearTimer()
  bursting.value = false
  // 先卸载再挂载，animation 才会重新从头播放
  nextTick(() => {
    bursting.value = true
    burstTimer = setTimeout(() => (bursting.value = false), 780)
  })
}

function clearTimer() {
  if (burstTimer) {
    clearTimeout(burstTimer)
    burstTimer = null
  }
}

/** 真机上的轻震反馈。H5 与部分基础库不支持，失败静默 */
function vibrate() {
  try {
    uni.vibrateShort({ fail: () => {} })
  } catch {
    /* 环境不支持，忽略 */
  }
}
</script>

<template>
  <view
    class="hb"
    :class="[`v-${variant}`, { on: liked, off: disabled, block }]"
    role="button"
    :aria-label="liked ? '已喜欢' : '喜欢'"
    :aria-disabled="disabled || liked"
    tabindex="0"
    @click="onTap"
    @keyup.enter="onTap"
  >
    <view class="stage" :style="stageStyle">
      <view class="shape" :style="shapeStyle">
        <view class="shape-pop" :class="{ pop: bursting }">
          <view class="lobe lobe-a"></view>
          <view class="lobe lobe-b"></view>
          <view class="body"></view>
        </view>
      </view>

      <view v-if="bursting" class="ring"></view>

      <view v-if="bursting" class="sparks">
        <view
          v-for="(a, i) in angles"
          :key="a"
          class="spark-wrap"
          :style="{ transform: `rotate(${a}deg)` }"
        >
          <view class="spark" :style="{ animationDelay: `${i * 26}ms` }"></view>
        </view>
      </view>
    </view>

    <text v-if="variant === 'pill'" class="hb-label">{{ labelText }}</text>
  </view>
</template>

<style lang="scss" scoped>
.hb {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  transition: transform 160ms ease, background-color 200ms ease, color 200ms ease,
    border-color 200ms ease;

  &:active {
    transform: scale(0.95);
  }
}

/* ---------- 形态 ---------- */

.v-pill {
  min-width: 132rpx;
  height: 64rpx;
  padding: 0 22rpx;
  border-radius: $radius;
  border: 1rpx solid $brand;
  background: $brand;
  color: $text-on-accent;

  &.on {
    background: $brand-light;
    color: $brand;
  }

  &.off {
    background: $bg-subtle;
    border-color: $bg-subtle;
    color: $text-placeholder;
  }
}

.v-icon {
  color: $brand;

  &.on {
    color: $brand;
  }

  &.off {
    color: $text-placeholder;
  }
}

.block {
  width: 100%;
  /* 撑满父容器（目前只有详情页底部那排）时，要和同排的 .btn-primary / .btn-plain 对齐：
   *   高度取 88rpx —— .v-pill 的 64rpx 是为卡片里的小按钮定的，并排时比兄弟矮一截；
   *   去掉左右内边距 —— 否则会比兄弟宽出左右各 22rpx，两个按钮一宽一窄。
   * 卡片里（无 block）的小 pill 不受影响，仍是 64rpx + 内边距的紧凑形态。 */
  height: 88rpx;
  padding: 0;
}

/* block 形态的标签跟随正文号，与兄弟按钮的文字大小一致 */
.block .hb-label {
  font-size: $font-body;
}

.hb-label {
  font-size: $font-caption;
  font-weight: 600;
  line-height: 1;
}

/* ---------- 心形本体 ---------- */

.stage {
  position: relative;
  flex-shrink: 0;
}

.shape {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: rotate(-45deg);
}

.shape-pop {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
}

/* 两个圆补出心形的上缘，左边那个绕到方块上方，右边那个挂到右侧 */
.lobe {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: currentColor;
}

.lobe-a {
  left: 0;
  top: -50%;
}

.lobe-b {
  left: 50%;
  top: 0;
}

.body {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  background: currentColor;
}

/* ---------- 点击反馈 ---------- */

.ring {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 12rpx;
  height: 12rpx;
  margin: -6rpx 0 0 -6rpx;
  border-radius: 50%;
  border: 2rpx solid currentColor;
  animation: hb-ring 620ms cubic-bezier(0.22, 0.61, 0.36, 1) forwards;
}

.sparks {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 0;
  height: 0;
}

.spark-wrap {
  position: absolute;
  left: 0;
  top: 0;
  width: 0;
  height: 0;
}

.spark {
  position: absolute;
  left: -3rpx;
  top: -3rpx;
  width: 6rpx;
  height: 6rpx;
  border-radius: 50%;
  background: currentColor;
  opacity: 0;
  animation: hb-spark 620ms cubic-bezier(0.22, 0.61, 0.36, 1) forwards;
}

@keyframes hb-ring {
  0% {
    transform: scale(0.4);
    opacity: 0.7;
  }
  100% {
    transform: scale(4.2);
    opacity: 0;
  }
}

@keyframes hb-spark {
  0% {
    transform: translate(0, 0) scale(1);
    opacity: 1;
  }
  100% {
    transform: translate(0, -46rpx) scale(0.2);
    opacity: 0;
  }
}

@keyframes hb-pop {
  0% {
    transform: scale(1);
  }
  28% {
    transform: scale(1.34);
  }
  56% {
    transform: scale(0.94);
  }
  100% {
    transform: scale(1);
  }
}

.pop {
  animation: hb-pop 520ms cubic-bezier(0.34, 1.56, 0.64, 1);
}
</style>
