<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const agreed = ref(false)
const loading = ref(false)
/** mock 模式下用它当 openid：同一个值就是同一个账号，方便反复测试 */
const mockCode = ref('test_user_1')

onLoad(() => {
  store.fetchConfig()
})

async function handleLogin() {
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并同意协议', icon: 'none' })
    return
  }
  if (loading.value) return
  loading.value = true

  try {
    if (!store.configLoaded || store.configError) await store.fetchConfig()
    if (store.configError) {
      uni.showToast({ title: '配置加载失败，请稍后重试', icon: 'none' })
      return
    }
    const code = store.mockMode ? mockCode.value.trim() : await getWxCode()
    if (!code) {
      uni.showToast({ title: '请填写测试账号', icon: 'none' })
      return
    }

    const res = await store.login(code)
    await store.fetchMe()

    if (res.need_profile || store.status === 0) {
      uni.reLaunch({ url: '/pages/profile/setup' })
    } else {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  } catch {
    // request 层已提示
  } finally {
    loading.value = false
  }
}

function getWxCode(): Promise<string> {
  return new Promise((resolve, reject) => {
    uni.login({
      provider: 'weixin',
      success: (res) => resolve(res.code),
      fail: reject,
    })
  })
}

function openAgreement(type: 'user' | 'privacy') {
  uni.navigateTo({ url: `/pages/common/agreement?type=${type}` })
}
</script>

<template>
  <view class="login">
    <view class="header">
      <view class="brand-row">
        <view class="logo">缘</view>
        <view class="title">缘来婚恋</view>
      </view>
      <view class="statement">
        <text>认真认识，</text>
        <text>再决定心意</text>
      </view>
      <view class="slogan">真实资料人工审核，让每一次认识更有依据。</view>
    </view>

    <view class="form">
      <view v-if="store.mockMode" class="mock-box">
        <view class="mock-label">开发账号</view>
        <view class="mock-tip">输入不同标识可切换测试用户</view>
        <input
          v-model="mockCode"
          class="mock-input"
          placeholder="如 test_user_1"
          placeholder-class="ph"
        />
      </view>

      <view v-if="store.configError" class="config-state">
        无法连接服务，请检查网络后重试。
      </view>

      <view
        class="btn-primary"
        :class="{ disabled: loading || !store.configLoaded }"
        @click="handleLogin"
      >
        {{
          !store.configLoaded
            ? '正在加载…'
            : store.configError
              ? '重新加载配置'
              : loading
              ? '登录中…'
              : store.mockMode
                ? '模拟登录'
                : '微信一键登录'
        }}
      </view>

      <view class="agreement" @click="agreed = !agreed">
        <view class="checkbox" :class="{ checked: agreed }">
          <text v-if="agreed" class="tick">✓</text>
        </view>
        <view class="text">
          我已阅读并同意
          <text class="link" @click.stop="openAgreement('user')">《用户协议》</text>
          和
          <text class="link" @click.stop="openAgreement('privacy')">《隐私政策》</text>
        </view>
      </view>
    </view>

    <view class="footer">
      <text>仅向 18 周岁以上单身人士提供服务</text>
      <text class="footer-sep"></text>
      <text>资料人工审核</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.login {
  min-height: 100dvh;
  max-width: 960rpx;
  margin: 0 auto;
  padding: 84rpx 48rpx calc(40rpx + env(safe-area-inset-bottom));
  display: flex;
  flex-direction: column;
  background: $bg-page;
}

.header {
  margin-bottom: 64rpx;
}

.brand-row {
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin-bottom: 80rpx;
}

.logo {
  width: 68rpx;
  height: 68rpx;
  line-height: 68rpx;
  border-radius: $radius;
  background: $bg-raised;
  border: 1rpx solid $brand;
  color: $brand;
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 32rpx;
  font-weight: 600;
  text-align: center;
  box-shadow: $shadow-card;
}

.title {
  font-size: 29rpx;
  font-weight: 650;
  color: $text-main;
  letter-spacing: 2rpx;
}

.statement {
  display: flex;
  flex-direction: column;
  color: $text-main;
  font-family: 'Songti SC', 'STSong', 'Noto Serif CJK SC', serif;
  font-size: 64rpx;
  font-weight: 600;
  letter-spacing: -2.4rpx;
  line-height: 1.2;
  text-wrap: balance;
  margin-bottom: 24rpx;
}

.slogan {
  max-width: 520rpx;
  font-size: 26rpx;
  color: $text-weak;
  line-height: 1.75;
}

.form {
  flex: 1;
}

.mock-box {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 30rpx;
  margin-bottom: 28rpx;
  box-shadow: $shadow-card;
}

.mock-label {
  font-size: 28rpx;
  color: $text-main;
  font-weight: 600;
  margin-bottom: 6rpx;
}

.mock-tip {
  font-size: 23rpx;
  color: $text-weak;
  line-height: 1.6;
  margin-bottom: 20rpx;
}

.mock-input {
  height: 88rpx;
  background: $bg-subtle;
  border: 1rpx solid transparent;
  border-radius: $radius;
  padding: 0 22rpx;
  font-size: 28rpx;
  transition: border-color 180ms ease, background-color 180ms ease;

  &:focus {
    border-color: $brand;
    background: $bg-card;
  }
}

.ph {
  color: $text-placeholder;
}

.config-state {
  margin-bottom: 20rpx;
  padding: 18rpx 20rpx;
  border-left: 4rpx solid $danger;
  background: $danger-light;
  color: $danger;
  font-size: 24rpx;
  line-height: 1.6;
}

.agreement {
  display: flex;
  align-items: flex-start;
  margin-top: 26rpx;
  padding: 10rpx 0;
}

.checkbox {
  width: 34rpx;
  height: 34rpx;
  border: 1rpx solid $border-strong;
  border-radius: $radius-sm;
  margin-right: 14rpx;
  flex-shrink: 0;
  margin-top: 4rpx;
  display: flex;
  align-items: center;
  justify-content: center;

  &.checked {
    background: $brand;
    border-color: $brand;
  }

  .tick {
    color: $text-on-accent;
    font-size: 22rpx;
  }
}

.agreement .text {
  flex: 1;
  font-size: 24rpx;
  color: $text-weak;
  line-height: 1.6;
}

.link {
  color: $brand;
}

.footer {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 16rpx;
  font-size: 22rpx;
  color: $text-placeholder;
  padding-top: 48rpx;
}

.form .btn-primary {
  height: 94rpx;
  line-height: 94rpx;
  border-radius: $radius;
  box-shadow: 0 12rpx 30rpx rgba(118, 43, 67, 0.045);
}

.footer-sep {
  width: 1rpx;
  height: 18rpx;
  background: $border-strong;
}

@media (max-height: 700px) {
  .login {
    padding-top: 48rpx;
  }

  .brand-row,
  .header {
    margin-bottom: 40rpx;
  }
}
</style>
