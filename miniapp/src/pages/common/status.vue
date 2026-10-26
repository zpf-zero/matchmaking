<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { Code } from '@/utils/request'

const code = ref(0)
const msg = ref('')

onLoad((q) => {
  code.value = Number(q?.code || 0)
  // 后端文案里可能带 %（如「已完成 100%」），直接 decode 会抛 URIError
  // 把 onLoad 打断成白屏。解不开就按原文显示
  const raw = q?.msg || ''
  try {
    msg.value = decodeURIComponent(raw)
  } catch {
    msg.value = raw
  }
})

/** 状态 -> 展示配置。用 computed 而非普通函数：
 *  原来模板里 view() 被调了 4 次，每次渲染都要重跑一遍 switch */
const view = computed(() => {
  switch (code.value) {
    case Code.ProfileAuditing:
      return {
        icon: '审',
        tone: 'pending',
        title: '资料审核中',
        desc: '我们正在核实你的资料，通常 24 小时内完成。期间可以先浏览其他人。',
        btn: '去逛逛',
        action: () => uni.reLaunch({ url: '/pages/index/index' }),
      }
    case Code.ProfileRejected:
      return {
        icon: '改',
        tone: 'warn',
        title: '资料未通过审核',
        desc: msg.value || '资料不符合规范，请修改后重新提交。',
        btn: '去修改资料',
        action: () => uni.reLaunch({ url: '/pages/profile/setup?mode=edit' }),
      }
    case Code.Banned:
      return {
        icon: '停',
        tone: 'danger',
        title: '账号已被封禁',
        desc: msg.value || '你的账号因违反社区规范已被封禁。如有疑问请联系客服。',
        btn: '',
        action: () => {},
      }
    default:
      return {
        icon: 'i',
        tone: 'pending',
        title: '提示',
        desc: msg.value || '当前操作无法完成',
        btn: '返回',
        action: () => uni.navigateBack(),
      }
  }
})
</script>

<template>
  <view class="page">
    <view class="icon" :class="view.tone">{{ view.icon }}</view>
    <view class="title">{{ view.title }}</view>
    <view class="desc">{{ view.desc }}</view>
    <view v-if="view.btn" class="btn-primary" @click="view.action()">{{ view.btn }}</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 80rpx 56rpx calc(80rpx + env(safe-area-inset-bottom));
}

.icon {
  width: 132rpx;
  height: 132rpx;
  line-height: 132rpx;
  margin-bottom: 40rpx;
  border-radius: 50%;
  background: $brand-light;
  color: $brand;
  font-size: 52rpx;
  font-weight: 600;
  text-align: center;
  position: relative;
  animation: icon-in 520ms cubic-bezier(0.34, 1.5, 0.5, 1) both;

  /* 图标外扩一圈涟漪，给状态一个「已确认」的落点 */
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 2rpx solid currentColor;
    opacity: 0.4;
    animation: icon-ripple 1100ms ease-out 260ms both;
  }

  /* 不同状态用不同色调：审核中是中性等待，被拒偏警示，封禁才是危险色。
     原先三种状态全用品牌色，用户看不出轻重差别 */
  &.warn {
    background: $warning-light;
    color: $warning;
  }

  &.danger {
    background: $danger-light;
    color: $danger;
  }
}

.title {
  font-size: $font-title;
  font-weight: 600;
  color: $text-main;
  letter-spacing: -1rpx;
  margin-bottom: 20rpx;
  animation: up-in 420ms ease 120ms both;
}

.desc {
  font-size: $font-mid;
  color: $text-weak;
  line-height: 1.7;
  margin-bottom: 56rpx;
  max-width: 560rpx;
  animation: up-in 420ms ease 200ms both;
}

.btn-primary {
  width: 420rpx;
  max-width: 100%;
  animation: up-in 420ms ease 300ms both;
}

@keyframes icon-in {
  from {
    transform: scale(0.5);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes icon-ripple {
  from {
    transform: scale(1);
    opacity: 0.4;
  }
  to {
    transform: scale(1.7);
    opacity: 0;
  }
}

@keyframes up-in {
  from {
    transform: translateY(14rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .icon,
  .icon::after,
  .title,
  .desc,
  .btn-primary {
    animation: none;
  }
}
</style>
