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
  uni.showModal({
    title: type === 'user' ? '用户协议' : '隐私政策',
    content: '协议文本待补充。提审前必须提供正式文本。',
    showCancel: false,
  })
}
</script>

<template>
  <view class="login">
    <view class="header">
      <view class="logo">缘</view>
      <view class="title">缘来婚恋</view>
      <view class="slogan">认真的人，遇见认真的人</view>
    </view>

    <view class="form">
      <!-- mock 模式：手填账号标识，换个值就是换个账号 -->
      <view v-if="store.mockMode" class="mock-box">
        <view class="mock-tip">
          开发模式：下方内容将作为账号标识（openid），
          换一个值即可模拟新用户
        </view>
        <input
          v-model="mockCode"
          class="mock-input"
          placeholder="如 test_user_1"
          placeholder-class="ph"
        />
      </view>

      <view class="btn-primary" :class="{ disabled: loading }" @click="handleLogin">
        {{ loading ? '登录中...' : store.mockMode ? '模拟登录' : '微信一键登录' }}
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

    <view class="footer">本平台仅向 18 周岁以上单身人士提供服务</view>
  </view>
</template>

<style lang="scss" scoped>
.login {
  min-height: 100vh;
  padding: 160rpx 60rpx 60rpx;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #fff0f3 0%, #f5f6f8 45%);
}

.header {
  text-align: center;
  margin-bottom: 100rpx;
}

.logo {
  width: 140rpx;
  height: 140rpx;
  line-height: 140rpx;
  margin: 0 auto 32rpx;
  border-radius: 40rpx;
  background: #e8506d;
  color: #fff;
  font-size: 64rpx;
  font-weight: 600;
}

.title {
  font-size: 44rpx;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 16rpx;
}

.slogan {
  font-size: 26rpx;
  color: #8a8f99;
}

.form {
  flex: 1;
}

.mock-box {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 40rpx;
}

.mock-tip {
  font-size: 22rpx;
  color: #ff8f1f;
  line-height: 1.6;
  margin-bottom: 16rpx;
}

.mock-input {
  height: 80rpx;
  background: #f5f6f8;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
}

.ph {
  color: #bbbfc4;
}

.agreement {
  display: flex;
  align-items: flex-start;
  margin-top: 40rpx;
  padding: 0 10rpx;
}

.checkbox {
  width: 34rpx;
  height: 34rpx;
  border: 1rpx solid #c9cdd4;
  border-radius: 50%;
  margin-right: 14rpx;
  flex-shrink: 0;
  margin-top: 4rpx;
  display: flex;
  align-items: center;
  justify-content: center;

  &.checked {
    background: #e8506d;
    border-color: #e8506d;
  }

  .tick {
    color: #fff;
    font-size: 22rpx;
  }
}

.agreement .text {
  flex: 1;
  font-size: 24rpx;
  color: #8a8f99;
  line-height: 1.6;
}

.link {
  color: #e8506d;
}

.footer {
  text-align: center;
  font-size: 22rpx;
  color: #bbbfc4;
  padding-bottom: 40rpx;
}
</style>
