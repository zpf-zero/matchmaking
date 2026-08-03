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
        icon: '⏳',
        title: '资料审核中',
        desc: '我们正在核实你的资料，通常 24 小时内完成。期间可以先浏览其他人。',
        btn: '去逛逛',
        action: () => uni.reLaunch({ url: '/pages/index/index' }),
      }
    case Code.ProfileRejected:
      return {
        icon: '⚠️',
        title: '资料未通过审核',
        desc: msg.value || '资料不符合规范，请修改后重新提交。',
        btn: '去修改资料',
        action: () => uni.reLaunch({ url: '/pages/profile/setup?mode=edit' }),
      }
    case Code.Banned:
      return {
        icon: '🚫',
        title: '账号已被封禁',
        desc: msg.value || '你的账号因违反社区规范已被封禁。如有疑问请联系客服。',
        btn: '',
        action: () => {},
      }
    default:
      return {
        icon: 'ℹ️',
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
  padding: 180rpx 60rpx;
  text-align: center;
}

.icon {
  font-size: 100rpx;
  margin-bottom: 40rpx;
}

.title {
  font-size: 36rpx;
  font-weight: 600;
  margin-bottom: 24rpx;
}

.desc {
  font-size: 27rpx;
  color: #8a8f99;
  line-height: 1.7;
  margin-bottom: 80rpx;
}
</style>
