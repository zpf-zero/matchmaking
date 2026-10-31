<script setup lang="ts">
import { computed, ref, watch } from 'vue'

/**
 * 配对成功全屏动效。
 *
 * 这是整个产品情绪浓度最高的一刻，原生 showModal 太寡淡了。
 * 编排顺序：遮罩淡入 → 双头像从两侧汇聚 → 心形弹跳放大 → 光晕扩散 + 粒子上升 → 文案与按钮入场。
 * 各段靠 animation-delay 串起来，顺序改一处即可整体前移后移。
 */
const props = withDefaults(
  defineProps<{
    visible: boolean
    myAvatar?: string
    peerAvatar?: string
    peerName?: string
    /** 私聊开关。关闭时不出现「去聊天」，只留一个「知道了」 */
    chatEnabled?: boolean
  }>(),
  { myAvatar: '', peerAvatar: '', peerName: '对方', chatEnabled: true },
)

const emit = defineEmits<{ chat: []; close: [] }>()

/** 收起时保留 DOM 一帧，让退出动画能播完 */
const dying = ref(false)

watch(
  () => props.visible,
  (v) => {
    if (v) {
      dying.value = false
    } else {
      dying.value = true
      setTimeout(() => (dying.value = false), 240)
    }
  }
)

const shown = computed(() => props.visible || dying.value)

const particles = Array.from({ length: 10 }, (_, i) => ({
  left: 8 + i * 8.4,
  delay: 620 + i * 42,
  scale: 0.6 + ((i * 37) % 50) / 100,
}))

function initialOf(name: string) {
  const n = (name || '').trim()
  if (!n) return '缘'
  return /^[a-zA-Z]/.test(n) ? n[0].toUpperCase() : n[0]
}
</script>

<template>
  <view v-if="shown" class="match-overlay" :class="{ out: !visible }" @click.self="emit('close')">
    <view class="halo halo-1"></view>
    <view class="halo halo-2"></view>

    <view class="sparks">
      <view
        v-for="(p, i) in particles"
        :key="i"
        class="spark"
        :style="{
          left: `${p.left}%`,
          animationDelay: `${p.delay}ms`,
          transform: `scale(${p.scale})`,
        }"
      ></view>
    </view>

    <view class="panel">
      <view class="pair">
        <view class="face face-l">
          <image v-if="myAvatar" class="face-pic" :src="myAvatar" mode="aspectFill" />
          <text v-else class="face-mark">我</text>
        </view>
        <view class="face face-r">
          <image v-if="peerAvatar" class="face-pic" :src="peerAvatar" mode="aspectFill" />
          <text v-else class="face-mark">{{ initialOf(peerName) }}</text>
        </view>
        <view class="heart-seat">
          <view class="mini-heart">
            <view class="mh-lobe mh-a"></view>
            <view class="mh-lobe mh-b"></view>
            <view class="mh-body"></view>
          </view>
        </view>
      </view>

      <view class="copy">
        <view class="title">配对成功</view>
        <view v-if="chatEnabled" class="sub">
          你和 {{ peerName }} 互相喜欢，现在可以开始聊天了
        </view>
        <view v-else class="sub">
          你和 {{ peerName }} 互相喜欢，私聊功能开放后就能开始聊天
        </view>
      </view>

      <view class="acts">
        <template v-if="chatEnabled">
          <view class="act-primary" @click="emit('chat')">去聊天</view>
          <view class="act-plain" @click="emit('close')">稍后</view>
        </template>
        <view v-else class="act-primary" @click="emit('close')">知道了</view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.match-overlay {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 56rpx;
  background: rgba(34, 31, 32, 0.72);
  backdrop-filter: blur(10rpx);
  animation: ov-in 240ms ease both;

  &.out {
    animation: ov-out 220ms ease both;
  }
}

/* 两层同色光晕错开扩散，比单层更有呼吸感 */
.halo {
  position: absolute;
  left: 50%;
  top: 42%;
  width: 520rpx;
  height: 520rpx;
  margin: -260rpx 0 0 -260rpx;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(223, 138, 163, 0.34) 0%, rgba(223, 138, 163, 0) 68%);
  animation: halo-out 1700ms cubic-bezier(0.22, 0.61, 0.36, 1) 380ms both;
}

.halo-2 {
  animation-delay: 640ms;
  animation-duration: 2000ms;
}

.sparks {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 62%;
}

.spark {
  position: absolute;
  bottom: 0;
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  background: rgba(255, 233, 240, 0.92);
  opacity: 0;
  animation: spark-up 1900ms cubic-bezier(0.22, 0.61, 0.36, 1) infinite;
}

.panel {
  position: relative;
  width: 100%;
  max-width: 560rpx;
  padding: 56rpx 40rpx 40rpx;
  border-radius: $radius-lg;
  background: $bg-raised;
  box-shadow: 0 24rpx 80rpx rgba(0, 0, 0, 0.28);
  animation: panel-in 520ms cubic-bezier(0.34, 1.4, 0.64, 1) 140ms both;
}

/* ---------- 双头像与心形 ---------- */

.pair {
  position: relative;
  height: 176rpx;
  margin-bottom: 36rpx;
}

/* 两个头像并排对称分居中线两侧，留 28rpx 间隙；心形落在中缝下方把两者连起来 */
.face {
  position: absolute;
  top: 0;
  left: 50%;
  width: 144rpx;
  height: 144rpx;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  border: 4rpx solid $bg-raised;
  background: $brand-light;
  box-shadow: 0 6rpx 20rpx rgba(78, 43, 55, 0.18);
}

/* 左头像：左边缘落在 中线左侧「自身宽度 + 14rpx」处 */
.face-l {
  animation: face-in-l 780ms cubic-bezier(0.34, 1.3, 0.64, 1) 200ms both;
}

/* 右头像：左边缘落在 中线右侧 14rpx 处 */
.face-r {
  animation: face-in-r 780ms cubic-bezier(0.34, 1.3, 0.64, 1) 200ms both;
}

.face-pic {
  width: 100%;
  height: 100%;
}

.face-mark {
  color: $brand;
  font-size: 52rpx;
  font-weight: 650;
  opacity: 0.75;
}

/* 心形坐在两个头像的交叠处，稍晚于头像入场 */
.heart-seat {
  position: absolute;
  left: 50%;
  bottom: -14rpx;
  width: 64rpx;
  height: 64rpx;
  margin-left: -32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: $bg-raised;
  box-shadow: 0 6rpx 18rpx rgba(78, 43, 55, 0.2);
  animation: heart-in 620ms cubic-bezier(0.34, 1.56, 0.64, 1) 560ms both;
}

.mini-heart {
  position: relative;
  width: 38rpx;
  height: 38rpx;
  transform: rotate(-45deg);
}

.mh-lobe,
.mh-body {
  position: absolute;
  background: $brand;
}

.mh-lobe {
  width: 100%;
  height: 100%;
  border-radius: 50%;
}

.mh-a {
  left: 0;
  top: -50%;
}

.mh-b {
  left: 50%;
  top: 0;
}

.mh-body {
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
}

/* ---------- 文案与操作 ---------- */

.copy {
  text-align: center;
  animation: copy-in 460ms ease 780ms both;
}

.title {
  font-size: $font-title;
  font-weight: 650;
  color: $text-main;
  letter-spacing: -0.8rpx;
  margin-bottom: 14rpx;
}

.sub {
  font-size: $font-mid;
  color: $text-weak;
  line-height: 1.7;
}

.acts {
  display: flex;
  gap: 20rpx;
  margin-top: 44rpx;
  animation: copy-in 460ms ease 900ms both;
}

.act-primary {
  flex: 1;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  border-radius: $radius;
  background: $brand;
  color: $text-on-accent;
  font-size: $font-body;
  font-weight: 650;
  transition: transform 160ms ease;

  &:active {
    transform: scale(0.97);
  }
}

.act-plain {
  width: 168rpx;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  border-radius: $radius;
  border: 1rpx solid $border-strong;
  color: $text-sub;
  font-size: $font-body;
  font-weight: 500;
  transition: transform 160ms ease;

  &:active {
    transform: scale(0.97);
  }
}

/* ---------- 动画 ---------- */

@keyframes ov-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes ov-out {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}

@keyframes panel-in {
  from {
    opacity: 0;
    transform: translateY(36rpx) scale(0.92);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 从头像从画面外侧滑入，停在分居中线两侧的位置（间距 28rpx） */
@keyframes face-in-l {
  from {
    transform: translateX(-360rpx) scale(0.6);
    opacity: 0;
  }
  to {
    transform: translateX(-158rpx) scale(1);
    opacity: 1;
  }
}

@keyframes face-in-r {
  from {
    transform: translateX(216rpx) scale(0.6);
    opacity: 0;
  }
  to {
    transform: translateX(14rpx) scale(1);
    opacity: 1;
  }
}

@keyframes heart-in {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  60% {
    transform: scale(1.24);
    opacity: 1;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes copy-in {
  from {
    opacity: 0;
    transform: translateY(14rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes halo-out {
  0% {
    transform: scale(0.3);
    opacity: 0.9;
  }
  100% {
    transform: scale(1.5);
    opacity: 0;
  }
}

@keyframes spark-up {
  0% {
    transform: translateY(0);
    opacity: 0;
  }
  18% {
    opacity: 1;
  }
  100% {
    transform: translateY(-560rpx);
    opacity: 0;
  }
}
</style>
