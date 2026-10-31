<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { authApi } from '@/api'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const agreed = ref(false)
const loading = ref(false)

onLoad(() => {
  store.fetchConfig()
})

// #ifdef H5
/* H5 没有微信授权，手机号 + 验证码就是完整的登录链路 */
const phone = ref('')
const smsCode = ref('')
const sending = ref(false)
const countdown = ref(0)
/** 开发环境 mock 短信时由接口回传，点一下即可填入；生产不会返回 */
const devCode = ref('')
const errorMsg = ref('')
let timer: ReturnType<typeof setInterval> | null = null

const phoneOk = computed(() => /^1[3-9]\d{9}$/.test(phone.value))
const canSend = computed(() => phoneOk.value && !sending.value && countdown.value <= 0)
const canSubmit = computed(() => phoneOk.value && /^\d{6}$/.test(smsCode.value) && !loading.value)

function clearTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

onUnmounted(clearTimer)

function startCountdown(seconds: number) {
  countdown.value = seconds
  clearTimer()
  timer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) clearTimer()
  }, 1000)
}

async function sendCode() {
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并同意协议', icon: 'none' })
    return
  }
  if (!phoneOk.value) {
    errorMsg.value = '请输入正确的 11 位手机号'
    return
  }
  if (!canSend.value) return

  sending.value = true
  errorMsg.value = ''
  try {
    const res = await authApi.sendSmsCode(phone.value)
    devCode.value = res.dev_code || ''
    startCountdown(res.resend_in || 60)
    uni.showToast({ title: '验证码已发送', icon: 'none' })
  } catch (e: any) {
    errorMsg.value = e?.message || '发送失败，请稍后重试'
  } finally {
    sending.value = false
  }
}

async function handlePhoneLogin() {
  if (!phoneOk.value) {
    errorMsg.value = '请输入正确的 11 位手机号'
    return
  }
  if (!/^\d{6}$/.test(smsCode.value)) {
    errorMsg.value = '请输入 6 位验证码'
    return
  }

  errorMsg.value = ''
  try {
    await store.phoneLogin(phone.value, smsCode.value)
    await store.fetchMe()
    uni.reLaunch({ url: store.needsProfile ? '/pages/profile/setup' : store.homePath })
  } catch (e: any) {
    errorMsg.value = e?.message || '登录失败，请稍后重试'
  }
}
// #endif

// #ifdef MP-WEIXIN
/** mock 模式下用它当 openid：同一个值就是同一个账号，方便反复测试 */
const mockCode = ref('test_user_1')

/**
 * 预设账号。
 *
 * mock 模式下 openid 就是这里输入的字符串 —— 敲错了不会报错，而是当场注册一个
 * 新用户，然后被「完善资料」向导拦住，很难意识到是账号名打错。
 * 这里把常用账号直接列出来点选，避免手打错。
 */
const mockPresets = [
  { code: 'admin', label: '管理员', note: '超管' },
  { code: 'test_user_1', label: '测试小美', note: '正常' },
  { code: 'test_user_2', label: '测试小明', note: '正常' },
]

async function handleWxLogin() {
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

    await store.login(code)
    await store.fetchMe()

    // 运营角色（审核员及以上）登录只为进后台，直接落到工作台；
    // 普通用户才按资料完善度分流。判据统一走 store.needsProfile。
    if (store.needsProfile) {
      uni.reLaunch({ url: '/pages/profile/setup' })
    } else {
      uni.reLaunch({ url: store.homePath })
    }
  } catch {
    // request 层已提示
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
// #endif

async function handleLogin() {
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并同意协议', icon: 'none' })
    return
  }
  if (loading.value) return
  loading.value = true

  // #ifdef H5
  await handlePhoneLogin()
  // #endif
  // #ifdef MP-WEIXIN
  await handleWxLogin()
  // #endif

  loading.value = false
}

function openAgreement(type: 'user' | 'privacy') {
  uni.navigateTo({ url: `/pages/common/agreement?type=${type}` })
}
</script>

<template>
  <view class="login">
    <view class="header">
      <view class="brand-row enter" style="animation-delay: 40ms">
        <view class="logo">缘</view>
        <view class="title">良缘天合</view>
      </view>
      <view class="statement">
        <!-- 品牌主张两行错峰升起，是全页唯一的「展演」时刻 -->
        <text class="line enter" style="animation-delay: 180ms">认真认识，</text>
        <text class="line enter" style="animation-delay: 300ms">再决定心意</text>
      </view>
      <view class="slogan enter" style="animation-delay: 420ms">
        真实资料人工审核，让每一次认识更有依据。
      </view>
    </view>

    <view class="form enter" style="animation-delay: 520ms">
      <!-- #ifdef H5 -->
      <view class="auth-head">
        <text class="auth-title">手机号登录</text>
        <text class="auth-note">无需密码 · 验证码直达</text>
      </view>

      <view class="auth-card">
        <view class="auth-field">
          <text class="field-label">手机号</text>
          <text class="dial-code">+86</text>
          <input
            v-model="phone"
            class="auth-input"
            type="number"
            maxlength="11"
            placeholder="请输入手机号"
            placeholder-class="ph"
            @input="errorMsg = ''"
          />
        </view>
        <view class="auth-field">
          <text class="field-label">验证码</text>
          <input
            v-model="smsCode"
            class="auth-input code-input"
            type="number"
            maxlength="6"
            placeholder="6 位验证码"
            placeholder-class="ph"
            @input="errorMsg = ''"
          />
          <view class="send-code" :class="{ disabled: !canSend }" @click="sendCode">
            {{ sending ? '发送中' : countdown > 0 ? `${countdown}s` : '获取验证码' }}
          </view>
        </view>
      </view>

      <!-- mock 短信的验证码回显。生产环境接口不返回 dev_code，这一块不会出现 -->
      <view v-if="devCode" class="dev-code">
        <text>开发模式验证码</text>
        <text class="dev-code-value" @click="smsCode = devCode">{{ devCode }}</text>
        <text class="dev-code-tip">点击填入</text>
      </view>

      <view v-if="errorMsg" class="form-error">{{ errorMsg }}</view>

      <view class="btn-primary auth-submit" :class="{ disabled: !canSubmit }" @click="handleLogin">
        {{ loading ? '登录中…' : '登录 / 注册' }}
      </view>
      <view class="auth-footnote">未注册的手机号将自动创建账号</view>
      <!-- #endif -->

      <!-- #ifdef MP-WEIXIN -->
      <view v-if="store.mockMode" class="mock-box">
        <view class="mock-label">开发账号</view>
        <view class="mock-tip">
          输错标识会当场注册成普通用户，请从下面点选或照抄
        </view>
        <view class="mock-presets">
          <view
            v-for="p in mockPresets"
            :key="p.code"
            class="preset"
            :class="{ on: mockCode === p.code }"
            @click="mockCode = p.code"
          >
            <text class="preset-name">{{ p.label }}</text>
            <text class="preset-note">{{ p.note }}</text>
          </view>
        </view>
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
      <!-- #endif -->

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

    <view class="footer enter" style="animation-delay: 640ms">
      <text>仅向 18 周岁以上人士提供服务</text>
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
  padding: 72rpx 48rpx calc(40rpx + env(safe-area-inset-bottom));
  display: flex;
  flex-direction: column;
  background: $bg-page;
}

.header {
  /* 与 .form 的 auto 下边距配合，把品牌区+表单整体垂直居中 */
  margin-top: auto;
  margin-bottom: 48rpx;
}

.brand-row {
  display: flex;
  align-items: center;
  gap: 18rpx;
  margin-bottom: 56rpx;
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
  font-size: $font-lead;
  font-weight: 600;
  text-align: center;
  box-shadow: $shadow-card;
}

.title {
  font-size: $font-body;
  font-weight: 650;
  color: $text-main;
  letter-spacing: 2rpx;
}

/*
 * 品牌主张是整站唯一使用衬线的一处：
 * 宋体在中文里自带「郑重、书面」的语气，与严肃婚恋的定位一致。
 * 其余界面文字一律无衬线，衬线不作为通用默认。
 */
.statement {
  display: flex;
  flex-direction: column;
  color: $text-main;
  font-family: 'Songti SC', 'STSong', 'Noto Serif CJK SC', serif;
  font-size: $font-display;
  font-weight: 600;
  letter-spacing: -2.4rpx;
  line-height: 1.2;
  text-wrap: balance;
  margin-bottom: 24rpx;

  .line {
    display: block;
  }
}

/* 入场：元素从下方轻微浮起并淡入。
   位移刻意做得克制（18rpx），登录页要的是「稳」，不是活泼 */
.enter {
  animation: enter-up 560ms cubic-bezier(0.22, 0.8, 0.3, 1) both;
}

@keyframes enter-up {
  from {
    transform: translateY(18rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* 用户开了系统「减弱动态效果」时直接落到终态，不做位移 */
@media (prefers-reduced-motion: reduce) {
  .enter {
    animation: none;
  }
}

.slogan {
  max-width: 520rpx;
  font-size: $font-mid;
  color: $text-weak;
  line-height: 1.75;
}

.form {
  margin-bottom: auto;
}

/* ---------- H5 手机号登录 ---------- */

.auth-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 0 4rpx;
  margin-bottom: 18rpx;
}

.auth-title {
  font-size: $font-body;
  font-weight: 600;
  color: $text-main;
  letter-spacing: 1rpx;
}

.auth-note {
  font-size: $font-micro;
  color: $text-placeholder;
  letter-spacing: 0.5rpx;
}

.auth-card {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 0 28rpx;
  box-shadow: $shadow-card;
}

.auth-field {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 108rpx;
  border-bottom: 1rpx solid $border;

  &:last-child {
    border-bottom: none;
  }

  /*
   * 聚焦时标签变色 + 底线自左向右铺开。
   * 比整框描边克制，也给键盘操作一个明确的落点（仅 H5 使用，小程序端会忽略）
   */
  &::after {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    bottom: -1rpx;
    height: 2rpx;
    background: $brand;
    transform: scaleX(0);
    transform-origin: left;
    transition: transform 240ms cubic-bezier(0.16, 1, 0.3, 1);
  }

  &:focus-within {
    .field-label {
      color: $brand;
    }

    &::after {
      transform: scaleX(1);
    }
  }
}

.field-label {
  width: 110rpx;
  flex-shrink: 0;
  font-size: $font-caption;
  color: $text-weak;
  letter-spacing: 1rpx;
  transition: color 180ms ease;
}

.dial-code {
  margin-right: 18rpx;
  font-size: $font-mid;
  color: $text-weak;
  font-variant-numeric: tabular-nums;
}

.auth-input {
  flex: 1;
  min-width: 0;
  height: 72rpx;
  font-size: $font-lead;
  font-weight: 550;
  color: $text-main;
}

.code-input {
  letter-spacing: 6rpx;
  font-variant-numeric: tabular-nums;
}

.send-code {
  flex-shrink: 0;
  padding: 12rpx 0 12rpx 24rpx;
  border-left: 1rpx solid $border;
  font-size: $font-caption;
  font-weight: 550;
  color: $brand;
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
  transition: color 180ms ease;

  &.disabled {
    color: $text-placeholder;
  }

  &:active {
    color: $brand-dark;
  }
}

.dev-code {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 18rpx;
  padding: 16rpx 20rpx;
  border-radius: $radius-sm;
  background: $warning-light;
  font-size: $font-micro;
  color: $warning;
  line-height: 1.5;
}

.dev-code-value {
  padding: 2rpx 10rpx;
  border-radius: $radius-sm;
  background: $bg-raised;
  font-weight: 650;
  letter-spacing: 3rpx;
  font-variant-numeric: tabular-nums;
}

.dev-code-tip {
  margin-left: auto;
  color: $text-placeholder;
}

.form-error {
  margin-top: 18rpx;
  padding: 16rpx 20rpx;
  border-left: 4rpx solid $danger;
  background: $danger-light;
  color: $danger;
  font-size: $font-caption;
  line-height: 1.6;
}

.auth-submit {
  margin-top: 28rpx;
}

.auth-footnote {
  margin-top: 18rpx;
  text-align: center;
  font-size: $font-micro;
  color: $text-placeholder;
  letter-spacing: 0.5rpx;
}

/* ---------- 小程序开发账号 ---------- */

.mock-box {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 30rpx;
  margin-bottom: 28rpx;
  box-shadow: $shadow-card;
}

.mock-label {
  font-size: $font-body;
  color: $text-main;
  font-weight: 600;
  margin-bottom: 6rpx;
}

.mock-tip {
  font-size: $font-caption;
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
  font-size: $font-body;
  transition: border-color 180ms ease, background-color 180ms ease;

  &:focus {
    border-color: $brand;
    background: $bg-card;
  }
}

/* 预设账号：点一下就填好，省得手打出「差一个字母」的新账号 */
.mock-presets {
  display: flex;
  gap: 14rpx;
  margin-bottom: 20rpx;
}

.preset {
  flex: 1;
  min-width: 0;
  padding: 16rpx 12rpx;
  border: 1rpx solid $border;
  border-radius: $radius;
  background: $bg-card;
  text-align: center;
  transition: border-color 180ms ease, background-color 180ms ease;

  &:active {
    background: $bg-subtle;
  }

  &.on {
    border-color: $brand;
    background: $bg-accent;
  }
}

.preset-name {
  display: block;
  font-size: $font-caption;
  font-weight: 600;
  color: $text-main;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preset-note {
  display: block;
  margin-top: 4rpx;
  font-size: $font-micro;
  color: $text-weak;
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
  font-size: $font-caption;
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
  transition: background-color 160ms ease, border-color 160ms ease;

  &.checked {
    background: $brand;
    border-color: $brand;

    .tick {
      animation: tick-in 260ms cubic-bezier(0.34, 1.56, 0.64, 1);
    }
  }

  .tick {
    color: $text-on-accent;
    font-size: $font-micro;
  }
}

@keyframes tick-in {
  from {
    transform: scale(0.3);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

.agreement .text {
  flex: 1;
  font-size: $font-caption;
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
  font-size: $font-micro;
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

  .header {
    margin-top: 0;
  }

  .brand-row,
  .header {
    margin-bottom: 40rpx;
  }
}
</style>
