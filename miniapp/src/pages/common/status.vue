<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { Code } from '@/utils/request'

const code = ref(0)
const msg = ref('')

onLoad((q) => {
  code.value = Number(q?.code || 0)
  msg.value = decodeURIComponent(q?.msg || '')
})

const view = () => {
  switch (code.value) {
    case Code.ProfileAuditing:
      return {
        icon: '审',
        title: '资料审核中',
        desc: '我们正在核实你的资料，通常 24 小时内完成。期间可以先浏览其他人。',
        btn: '去逛逛',
        action: () => uni.reLaunch({ url: '/pages/index/index' }),
      }
    case Code.ProfileRejected:
      return {
        icon: '改',
        title: '资料未通过审核',
        desc: msg.value || '资料不符合规范，请修改后重新提交。',
        btn: '去修改资料',
        action: () => uni.reLaunch({ url: '/pages/profile/setup?mode=edit' }),
      }
    case Code.Banned:
      return {
        icon: '停',
        title: '账号已被封禁',
        desc: msg.value || '你的账号因违反社区规范已被封禁。如有疑问请联系客服。',
        btn: '',
        action: () => {},
      }
    default:
      return {
        icon: 'i',
        title: '提示',
        desc: msg.value || '当前操作无法完成',
        btn: '返回',
        action: () => uni.navigateBack(),
      }
  }
}
</script>

<template>
  <view class="page">
    <view class="icon">{{ view().icon }}</view>
    <view class="title">{{ view().title }}</view>
    <view class="desc">{{ view().desc }}</view>
    <view v-if="view().btn" class="btn-primary" @click="view().action()">{{ view().btn }}</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  min-height: 100dvh;
  padding: 148rpx 48rpx 80rpx;
}

.icon {
  width: 88rpx;
  height: 88rpx;
  line-height: 88rpx;
  margin-bottom: 40rpx;
  border-radius: $radius-lg;
  background: $brand-light;
  color: $brand;
  font-size: 38rpx;
  font-weight: 600;
  text-align: center;
}

.title {
  font-size: 42rpx;
  font-weight: 600;
  color: $text-main;
  letter-spacing: -1rpx;
  margin-bottom: 20rpx;
}

.desc {
  font-size: 27rpx;
  color: $text-weak;
  line-height: 1.7;
  margin-bottom: 64rpx;
  max-width: 600rpx;
}

.btn-primary {
  width: 100%;
}
</style>
