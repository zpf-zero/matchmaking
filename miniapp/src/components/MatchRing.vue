<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

/**
 * 匹配度环形进度。
 *
 * 小程序端没有 <svg>，所以用「左右两个矩形窗口 + 各自内部一块可旋转的半圆环」拼出圆环。
 * 半圆环用 border 的相邻两条边上色得到 —— 圆的四条边各占 90° 弧，交点落在四条对角线上：
 *   上 = [315°,45°]  右 = [45°,135°]  下 = [135°,225°]  左 = [225°,315°]
 * （角度按钟面记：0° 为 12 点，顺时针为正。）
 * 于是「右 + 下」是一段 180° 弧 [45°,225°]，「上 + 左」是另一段 180° 弧 [225°,45°]。
 * 把弧旋转到目标位置、再用只露一半的窗口裁切，旋转角就等价于进度角，交给 transition 走动画。
 *
 * 进度 Θ = value × 3.6，从 12 点起顺时针填充，即填充区为 [0°, Θ]。
 * 右窗口可见 [0°,180°]，左窗口可见 [180°,360°]。
 *
 * 右弧旋转 δ 后为 [45°+δ, 225°+δ]。Θ ≤ 180° 时要让它与右窗口的交集正好是 [0°,Θ]：
 *   下边界 45°+δ ≤ 0°、上边界 225°+δ = Θ ⟹ δ = Θ-225°（Θ>180° 时锁死在 -45°，右半环填满）。
 *   故 δ_右 = min(-45°, Θ-225°)。
 *
 * 左弧旋转 δ 后为 [225°+δ, 45°+δ]。Θ ≤ 180° 时左窗应全空，
 *   而两段 180° 弧互斥当且仅当左弧恰为右半圆 [0°,180°] ⟹ δ = 135°；
 *   Θ ≥ 180° 时需上边界 45°+δ = Θ ⟹ δ = Θ-45°。
 *   故 δ_左 = max(135°, Θ-45°)。
 *
 * 代入关键点自检：Θ=0 两窗皆空；Θ=90 仅右上象限；Θ=180 恰为右半环；
 * Θ=270 为右半环 + 左下象限；Θ=360 满环。
 */
const props = withDefaults(
  defineProps<{
    /** 匹配度，0-100 */
    value: number
    /** 直径，单位 rpx */
    size?: number
    /** 环宽，单位 rpx；不传则按直径的 9% 取，下限 5rpx */
    thickness?: number
    /** 是否展示中心的百分比数字 */
    showValue?: boolean
  }>(),
  {
    size: 88,
    thickness: 0,
    showValue: true,
  }
)

const ringWidth = computed(() =>
  props.thickness > 0 ? props.thickness : Math.max(5, Math.round(props.size * 0.09))
)

const safeValue = computed(() => {
  const v = Number(props.value)
  if (!Number.isFinite(v)) return 0
  return Math.min(100, Math.max(0, Math.round(v)))
})

/** 入场时从 0 涨到目标值，让环「走出去」而不是直接闪出来 */
const shown = ref(0)

onMounted(() => {
  // 留一帧先渲染出 Θ=0 的初始态，否则同一帧内改值 transition 不会触发
  setTimeout(() => (shown.value = safeValue.value), 60)
})

watch(safeValue, (v) => (shown.value = v))

const theta = computed(() => shown.value * 3.6)

const rightDeg = computed(() => Math.min(-45, theta.value - 225))
const leftDeg = computed(() => Math.max(135, theta.value - 45))

const boxStyle = computed(() => ({
  width: `${props.size}rpx`,
  height: `${props.size}rpx`,
}))

const semiStyle = computed(() => ({
  borderWidth: `${ringWidth.value}rpx`,
}))

const numStyle = computed(() => ({
  fontSize: `${Math.max(20, Math.round(props.size * 0.32))}rpx`,
}))
</script>

<template>
  <view class="ring" :style="boxStyle" :aria-label="`匹配度 ${safeValue}%`" role="img">
    <view class="track" :style="semiStyle"></view>

    <view class="half h-right">
      <view class="semi" :style="[semiStyle, { transform: `rotate(${rightDeg}deg)` }]"></view>
    </view>
    <view class="half h-left">
      <view class="semi alt" :style="[semiStyle, { transform: `rotate(${leftDeg}deg)` }]"></view>
    </view>

    <view v-if="showValue" class="center">
      <text class="num" :style="numStyle">{{ safeValue }}</text>
      <text class="pct">%</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.ring {
  position: relative;
  flex-shrink: 0;
}

.track {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  box-sizing: border-box;
  border-style: solid;
  border-color: $bg-subtle;
  border-radius: 50%;
}

/* 左右两个矩形窗口，把内部完整圆裁成各自那一半 */
.half {
  position: absolute;
  top: 0;
  width: 50%;
  height: 100%;
  overflow: hidden;
}

.h-right {
  right: 0;
}

.h-left {
  left: 0;
}

/*
 * 半圆环：宽取窗口的 200%（即与容器等宽）、高与容器等高，
 * 圆心因此正好落在容器中心。left 的百分比基于父窗口宽度，
 * 所以 -100% 恰好把圆心从窗口左侧拉回容器中心。
 * 只给相邻两条边上色，这两条边的圆弧合起来就是 180°。
 */
/* right + bottom 上色，基准弧 [45°,225°]，初始转到 Θ=0 对应的 -225° */
.semi {
  position: absolute;
  top: 0;
  left: -100%;
  width: 200%;
  height: 100%;
  box-sizing: border-box;
  border-style: solid;
  border-color: transparent;
  border-radius: 50%;
  border-right-color: $brand;
  border-bottom-color: $brand;
  transform: rotate(-225deg);
  transition: transform 900ms cubic-bezier(0.22, 0.61, 0.36, 1);
}

/* top + left 上色，基准弧 [225°,45°]，初始转到 Θ=0 对应的 135°（此时完全落在右半平面） */
.semi.alt {
  left: 0;
  border-right-color: transparent;
  border-bottom-color: transparent;
  border-top-color: $brand;
  border-left-color: $brand;
  transform: rotate(135deg);
}

.center {
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  transform: translateY(-50%);
  display: flex;
  align-items: baseline;
  justify-content: center;
  font-variant-numeric: tabular-nums;
}

.num {
  color: $brand;
  font-weight: 650;
  line-height: 1;
  letter-spacing: -0.6rpx;
}

.pct {
  color: $brand;
  font-size: $font-micro;
  font-weight: 600;
  margin-left: 1rpx;
}
</style>
