<script setup lang="ts">
import { computed, ref } from 'vue'
import { onShow, onShareAppMessage } from '@dcloudio/uni-app'
import { inviteApi } from '@/api'
import type { MyInviteInfo } from '@/api/types'
import { Status, useUserStore } from '@/stores/user'

const store = useUserStore()
const info = ref<MyInviteInfo | null>(null)
const loading = ref(false)
const copied = ref(false)

/**
 * 分享给好友。
 *
 * 合规红线（《微信小程序平台运营规范》5.1）：
 *   - 文案不写「分享得奖励」「不分享就…」—— 5.1.1 强制分享 / 5.1.4 无深度互动获利
 *   - 不给分享按钮加红点、不做动态形变 —— 5.1.2 暗示分享
 *   - 不诱导分享到朋友圈 —— 5.1.3
 *   - 规则写清、口径真实 —— 5.1.5 误导用户
 * 分享本身不产生任何奖励。权益只有在「好友注册 + 完善资料 + 通过人工审核」后
 * 才由服务端结算，这正是规范定义的「深度互动」。
 *
 * 注意：必须用 uni-app 导出的 onShareAppMessage（而不是普通 function 声明）。
 * 组合式 API 下它靠 injectHook 把回调注册进当前页面实例，普通函数不会被注册，
 * 编译产物里连函数体都会被 tree-shake 掉 —— 表现为「按钮点了没反应」。
 * pages.json 里该页还需 enableShareAppMessage: true。
 */
onShareAppMessage(() => {
  const code = info.value?.code || ''
  return {
    title: '缘来婚恋 · 真实资料人工审核，认真认识一个人',
    path: code ? `/pages/login/index?invite=${code}` : '/pages/login/index',
  }
})

onShow(async () => {
  if (!store.isLogin) return
  loading.value = true
  try {
    info.value = await inviteApi.myCode()
    await store.fetchMe() // 让「我的」页的权益状态同步刷新
    notifySettlement()
  } catch {
    // request 层已提示
  } finally {
    loading.value = false
  }
})

/**
 * 结算回执。结算发生在好友过审那一刻，邀请人多半不在场：
 * 下次进邀请页时一次性告知「权益已到账」，并清除「我的」页红点。
 * 标记已读不等弹层关闭 —— 内容已看到，回调失败下次还会再提示
 */
function notifySettlement() {
  if (!info.value || info.value.unseen_settled <= 0) return
  uni.showModal({
    title: '邀请生效',
    content: '你邀请的好友已通过资料审核，你和 TA 的完整资料查看权益已永久解锁。',
    showCancel: false,
    confirmText: '知道了',
  })
  inviteApi
    .seen()
    .then(() => {
      store.inviteUnseen = 0
    })
    .catch(() => {})
}

/**
 * 邀请入口是否可用。后端只认「已通过人工审核」的邀请人：
 * 未过审账号发出的码，朋友填了也会被拒。这里提前收住复制与分享，
 * 避免把一个无效码发出去、朋友填码时反而吃一个报错
 */
const canInvite = computed(() => store.status === Status.Normal)

const vipText = computed(() => (store.isVip ? '已解锁完整资料 · 永久有效' : '尚未解锁完整资料'))

/**
 * 是否已达邀请上限。
 *
 * 上限是防刷用的风控阈值，不是「福利目标」，正常用户几乎不会碰到 ——
 * 体制内用户的社交圈层本来有限。所以只在真正触顶时才提示，
 * 不把它做成常驻的数字指标（那会暗示「邀请越多越好」，与事实不符）。
 */
const reachedCap = computed(() => {
  const max = info.value?.max_settle
  if (!max) return false
  return (info.value?.settled_count ?? 0) >= max
})

/** 复制邀请码。用户可发到社群/私聊，不必走小程序分享，绕开「分享悖论」 */
function copyCode() {
  const code = info.value?.code
  if (!code) return
  uni.setClipboardData({
    data: code,
    success: () => {
      copied.value = true
      uni.showToast({ title: '邀请码已复制', icon: 'none' })
      setTimeout(() => (copied.value = false), 2000)
    },
  })
}

// #ifdef H5
/**
 * H5 没有 open-type="share"，分享按钮是死的 —— 降级为复制落地链接。
 * 链接用 hash 路由拼（manifest.json router.mode=hash），
 * location.href 去掉 # 之后的部分就是部署基路径，直接复用
 */
function copyLink() {
  const code = info.value?.code
  if (!code) return
  const url = location.href.split('#')[0] + '#/pages/login/index?invite=' + code
  uni.setClipboardData({
    data: url,
    success: () => {
      uni.showToast({ title: '链接已复制，粘贴给好友即可', icon: 'none' })
    },
  })
}
// #endif

function openRules() {
  uni.navigateTo({ url: '/pages/invite/rules' })
}

/** 去填三观问答 —— 权益内容里唯一「免费能产出」的部分，主动引导用户先答 */
function goThreeView() {
  uni.navigateTo({ url: '/pages/user/three-view' })
}

const statusMap: Record<number, string> = {
  0: '待好友通过审核',
  1: '已发放权益',
  2: '未达成',
}
</script>

<template>
  <view class="page">
    <!-- 权益状态 -->
    <view class="hero">
      <view class="hero-label" :class="{ on: store.isVip }">{{ vipText }}</view>
      <view class="hero-title">邀请好友，解锁完整资料</view>
      <view class="hero-desc">
        好友通过资料审核后，你和 TA 各解锁完整资料查看权限。
        权益一次解锁即永久拥有，<text class="em">邀请一位好友即可获得</text>，邀请更多好友不会带来额外权益。
      </view>
    </view>

    <!-- 邀请码 -->
    <view class="card code-card">
      <view class="code-label">我的邀请码</view>
      <view class="code-row">
        <text class="code-value">{{ info?.code || '——' }}</text>
        <view v-if="canInvite" class="code-copy" :class="{ on: copied }" @click="copyCode">
          {{ copied ? '已复制' : '复制' }}
        </view>
      </view>
      <template v-if="canInvite">
        <!-- #ifdef MP-WEIXIN -->
        <view class="code-hint">
          把邀请码发给朋友，TA 在填资料时填写即可。也可以在下方直接分享小程序。
        </view>
        <button class="share-btn" open-type="share">分享给好友</button>
        <!-- #endif -->
        <!-- #ifdef H5 -->
        <view class="code-hint">
          把邀请码或下方链接发给朋友，TA 注册后在填资料时填入即可。
        </view>
        <view class="share-btn" @click="copyLink">复制邀请链接</view>
        <!-- #endif -->
      </template>
      <!-- 后端只认已过审的邀请人：未过审就分发邀请码，朋友填码只会吃一个报错 -->
      <view v-else class="code-hint pending">
        你的资料通过人工审核后即可邀请好友，届时邀请码立即生效
      </view>
    </view>

    <!-- 进度。刻意不展示「可邀请上限」这类数字：
         权益是一次性永久解锁，邀请第 1 位与第 20 位拿到的权益完全相同。
         把上限当标题数字展示，会让新用户联想「邀请越多越好」，与实际不符（5.1.5），
         而实际上第 2 位起对本人没有额外收益。仅在真正触顶时才提示，见下方邀请记录。 -->
    <view class="stat-row">
      <view class="stat">
        <text class="num">{{ info?.settled_count ?? 0 }}</text>
        <text class="lab">已成功邀请</text>
      </view>
      <view class="stat">
        <text class="num">{{ info?.pending_count ?? 0 }}</text>
        <text class="lab">待好友通过审核</text>
      </view>
    </view>

    <!-- 权益说明。刻意与后端实现一一对应：
         「谁看过我」「更多照片」「三观答案」「原生家庭」「曝光加权」，
         改实现时必须同步改这里 —— 规则与实现不符构成 5.1.5 误导用户 -->
    <view class="card">
      <view class="section-title">解锁后你能看到</view>
      <view class="benefit">
        <view class="benefit-dot"></view>
        <view class="benefit-copy">
          <text class="benefit-name">更多照片</text>
          <text class="benefit-note">免费用户可看前 3 张，解锁后可见对方全部照片</text>
        </view>
      </view>
      <view class="benefit">
        <view class="benefit-dot"></view>
        <view class="benefit-copy">
          <text class="benefit-name">三观问答答案</text>
          <text class="benefit-note">答题本身免费，解锁后可查看彼此的选择与一致率</text>
        </view>
      </view>
      <view class="benefit">
        <view class="benefit-dot"></view>
        <view class="benefit-copy">
          <text class="benefit-name">对方的原生家庭</text>
          <text class="benefit-note">仅展示对方主动授权公开的部分</text>
        </view>
      </view>
      <view class="benefit">
        <view class="benefit-dot"></view>
        <view class="benefit-copy">
          <text class="benefit-name">完整「谁看过我」</text>
          <text class="benefit-note">免费版只看最近 3 位，解锁后可看全部</text>
        </view>
      </view>
      <view class="benefit">
        <view class="benefit-dot"></view>
        <view class="benefit-copy">
          <text class="benefit-name">推荐页曝光加权</text>
          <text class="benefit-note">资料在推荐流中优先展示，被更多人看到</text>
        </view>
      </view>
      <view class="benefit-note-all">
        权益为平台赠送且永久有效，不影响任何基础功能 —— 未解锁也能正常浏览、喜欢和聊天。
      </view>
    </view>

    <!-- 三观问答入口。它是权益内容里唯一「免费就能产出」的部分，
         主动引导用户先答，才有东西可在解锁后展示 -->
    <view class="tv-entry" @click="goThreeView">
      <view class="tv-entry-copy">
        <text class="tv-entry-title">先回答三观问答</text>
        <text class="tv-entry-note">答题免费，答完才能看到彼此的选择</text>
      </view>
      <text class="arrow">›</text>
    </view>

    <!-- 邀请记录 -->
    <view class="card">
      <view class="section-title">邀请记录</view>
      <view v-if="!info?.records?.length" class="empty-tip">
        还没有邀请记录。把邀请码发给身边单身的朋友吧。
      </view>
      <view v-for="(r, i) in info?.records || []" :key="i" class="record">
        <view class="rec-copy">
          <text class="rec-name">{{ r.nickname }}</text>
          <text class="rec-time">{{ (r.created_at || '').slice(0, 10) }} 加入</text>
        </view>
        <text class="rec-status" :class="r.status === 1 ? 'ok' : 'wait'">
          {{ statusMap[r.status] || '' }}
        </text>
      </view>
      <!-- 只在真正触顶时出现，不是常驻提示 -->
      <view v-if="reachedCap" class="cap-tip">
        已达邀请结算上限（{{ info?.max_settle }} 位）。你的权益已在首位好友通过审核时获得，长期有效，无需继续邀请。
      </view>
    </view>

    <!-- 规则入口：5.1.5 要求活动规则在页面显著位置可查 -->
    <view class="rules-entry" @click="openRules">
      查看完整活动规则
      <text class="arrow">›</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 64rpx;
}

.hero {
  padding: 8rpx 8rpx 32rpx;
}

.hero-label {
  display: inline-block;
  font-size: $font-micro;
  padding: 5rpx 14rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
  color: $text-weak;
  margin-bottom: 20rpx;

  /* 已解锁时用品牌色，给一个明确的「你已经是特权用户」的确认 */
  &.on {
    background: $brand-light;
    color: $brand;
  }
}

.hero-title {
  font-size: $font-display;
  font-weight: 650;
  color: $text-main;
  letter-spacing: -1rpx;
  line-height: 1.25;
}

.hero-desc {
  margin-top: 14rpx;
  font-size: $font-mid;
  color: $text-weak;
  line-height: 1.7;

  .em {
    color: $text-main;
    font-weight: 600;
  }
}

.hero-expire {
  margin-top: 10rpx;
  font-size: $font-caption;
  color: $brand;
}

/* 三观问答引导条 */
.tv-entry {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 26rpx 28rpx;
  margin-bottom: 24rpx;
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;

  &:active {
    background: $bg-subtle;
  }

  .arrow {
    color: $text-placeholder;
    font-size: $font-title;
  }
}

.tv-entry-copy {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.tv-entry-title {
  font-size: $font-body;
  font-weight: 550;
  color: $text-main;
}

.tv-entry-note {
  font-size: $font-micro;
  color: $text-placeholder;
}

.card {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 28rpx;
  margin-bottom: 24rpx;
  box-shadow: $shadow-card;
}

.code-card {
  text-align: center;
}

.code-label {
  font-size: $font-caption;
  color: $text-weak;
  margin-bottom: 16rpx;
}

.code-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 20rpx;
}

.code-value {
  font-size: 60rpx;
  font-weight: 650;
  letter-spacing: 8rpx;
  color: $brand;
  font-variant-numeric: tabular-nums;
}

.code-copy {
  padding: 10rpx 22rpx;
  border-radius: $radius-sm;
  border: 1rpx solid $brand;
  color: $brand;
  font-size: $font-caption;
  transition: background-color 180ms ease, color 180ms ease;

  &.on {
    background: $brand;
    color: $text-on-accent;
  }
}

.code-hint {
  margin-top: 20rpx;
  font-size: $font-caption;
  color: $text-placeholder;
  line-height: 1.6;
}

/* 未过审时替代分享按钮出现的说明，用警示色和普通提示区分 */
.code-hint.pending {
  color: $warning;
}

.share-btn {
  margin-top: 26rpx;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: $radius;
  background: $brand;
  color: $text-on-accent;
  font-size: $font-body;
  font-weight: 650;

  &:active {
    opacity: 0.88;
  }
}

.stat-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14rpx;
  margin-bottom: 24rpx;
}

.stat {
  min-height: 138rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius;

  .num {
    font-size: $font-title;
    font-weight: 650;
    color: $text-main;
    margin-bottom: 8rpx;
    font-variant-numeric: tabular-nums;
  }

  .lab {
    font-size: $font-micro;
    color: $text-weak;
    text-align: center;
    padding: 0 8rpx;
    line-height: 1.4;
  }
}

.section-title {
  font-size: $font-lead;
  font-weight: 600;
  color: $text-main;
  margin-bottom: 20rpx;
}

.benefit {
  display: flex;
  align-items: flex-start;
  padding: 16rpx 0;
  border-bottom: 1rpx solid $border;

  &:last-of-type {
    border-bottom: none;
  }
}

.benefit-dot {
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: $brand;
  margin-top: 14rpx;
  margin-right: 18rpx;
  flex-shrink: 0;
}

.benefit-copy {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.benefit-name {
  font-size: $font-body;
  font-weight: 520;
  color: $text-main;
}

.benefit-note {
  font-size: $font-micro;
  color: $text-placeholder;
  line-height: 1.5;
}

.benefit-note-all {
  margin-top: 18rpx;
  padding: 18rpx 20rpx;
  background: $bg-subtle;
  border-radius: $radius-sm;
  font-size: $font-caption;
  color: $text-weak;
  line-height: 1.6;
}

.empty-tip {
  padding: 40rpx 0;
  text-align: center;
  font-size: $font-caption;
  color: $text-placeholder;
  line-height: 1.6;
}

.cap-tip {
  margin-top: 20rpx;
  padding: 18rpx 20rpx;
  background: $bg-subtle;
  border-radius: $radius-sm;
  font-size: $font-caption;
  color: $text-weak;
  line-height: 1.6;
}

.record {
  display: flex;
  align-items: center;
  padding: 18rpx 0;
  border-bottom: 1rpx solid $border;

  &:last-child {
    border-bottom: none;
  }
}

.rec-copy {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5rpx;
}

.rec-name {
  font-size: $font-body;
  color: $text-main;
}

.rec-time {
  font-size: $font-micro;
  color: $text-placeholder;
}

.rec-status {
  font-size: $font-caption;

  &.ok {
    color: $success;
  }

  &.wait {
    color: $warning;
  }
}

.rules-entry {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  padding: 26rpx;
  font-size: $font-caption;
  color: $text-weak;

  &:active {
    color: $brand;
  }

  .arrow {
    color: $text-placeholder;
  }
}
</style>
