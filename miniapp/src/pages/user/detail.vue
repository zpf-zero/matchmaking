<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import UserAvatar from '@/components/UserAvatar.vue'
import HeartButton from '@/components/HeartButton.vue'
import MatchRing from '@/components/MatchRing.vue'
import MatchSuccessOverlay from '@/components/MatchSuccessOverlay.vue'
import { commonApi, interactionApi, userApi } from '@/api'
import type { UserDetail } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const detail = ref<UserDetail | null>(null)
const loading = ref(true)
const userId = ref(0)
const matchVisible = ref(false)
const heartRef = ref<any>(null)
/** 相册里加载失败的下标。失败后画「加载失败」占位，不留灰块 */
const photoFailed = ref<Record<number, boolean>>({})

onLoad(async (q) => {
  userId.value = Number(q?.id || 0)
  if (!userId.value) {
    uni.showToast({ title: '参数错误', icon: 'none' })
    return
  }
  // 字典、题库、资料三件事互不依赖，并发发出去。
  // 之前是串行 await，在远程接口上等于白等两个往返，页面骨架会多停 100~300ms。
  await Promise.all([
    store.fetchDict().catch(() => {}),
    store.fetchThreeViewQuestions().catch(() => {}),
    load(),
  ])
  // 题库可能比资料晚到，补算一次三观对照，避免整块空白
  buildThreeViewRows()
})

async function load() {
  loading.value = true
  try {
    detail.value = await userApi.detail(userId.value)
    buildThreeViewRows()
  } catch {
    setTimeout(() => uni.navigateBack(), 1200)
  } finally {
    loading.value = false
  }
}

function previewPhoto(index: number) {
  if (!detail.value?.photos?.length) return
  uni.previewImage({ urls: detail.value.photos, current: index })
}

async function handleLike() {
  if (!store.canInteract) {
    uni.showToast({ title: '资料审核通过后才能表达心意', icon: 'none' })
    return
  }
  const d = detail.value!

  try {
    const res = await interactionApi.like(d.id)
    d.is_liked = true

    if (res.is_match) {
      d.is_matched = true
      d.conversation_id = res.conversation_id
      matchVisible.value = true
    } else {
      uni.showToast({ title: '已表达心意，等待对方回应', icon: 'none' })
    }
  } catch {
    // 已提示
  }
}

async function handlePass() {
  try {
    await interactionApi.pass(detail.value!.id)
    uni.showToast({ title: '已跳过', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 600)
  } catch {
    // 已提示
  }
}

function goChat() {
  const d = detail.value!
  matchVisible.value = false
  uni.navigateTo({
    url: `/pages/chat/index?id=${d.conversation_id}&name=${encodeURIComponent(d.nickname)}&peer=${d.id}`,
  })
}

/** 三观题目的展示行：题目 + 我的选择 + 对方的选择 */
const threeViewRows = ref<{ title: string; mine: string; theirs: string; same: boolean }[]>([])

/**
 * 把三观答案摊平成可读的行。
 *
 * 只在「我已解锁」时才有对方答案（后端未解锁只回 answered 计数）。
 * 一致率由后端算好，这里只负责配对文案 —— 前端不重复计算，
 * 避免两边口径漂移。
 */
function buildThreeViewRows() {
  const tv = detail.value?.three_view
  const questions = store.threeViewQuestions
  if (!tv || !tv.answers?.length || !questions.length) {
    threeViewRows.value = []
    return
  }
  threeViewRows.value = questions.map((q, i) => {
    const mineIdx = tv.my_answers?.[i] ?? -1
    const theirsIdx = tv.answers?.[i] ?? -1
    return {
      title: q.title,
      mine: mineIdx >= 0 ? q.options[mineIdx] || '—' : '未作答',
      theirs: theirsIdx >= 0 ? q.options[theirsIdx] || '—' : '未作答',
      same: mineIdx >= 0 && theirsIdx >= 0 && mineIdx === theirsIdx,
    }
  })
}

/** 去邀请页解锁更多内容 */
function goInvite() {
  uni.navigateTo({ url: '/pages/invite/index' })
}

function moreActions() {
  uni.showActionSheet({
    itemList: ['举报该用户', '拉黑该用户'],
    success: (res) => {
      if (res.tapIndex === 0) doReport()
      else doBlock()
    },
  })
}

function doReport() {
  const reasons = store.dict['report_reason'] || []
  uni.showActionSheet({
    itemList: reasons.map((r) => r.label),
    success: async (res) => {
      try {
        await commonApi.report({
          target_user_id: detail.value!.id,
          reason: reasons[res.tapIndex].value,
        })
        uni.showToast({ title: '举报已提交，我们会尽快核实', icon: 'none' })
      } catch {
        // 已提示
      }
    },
  })
}

function doBlock() {
  uni.showModal({
    title: '确认拉黑',
    content: '拉黑后双方将互相不可见，已有的配对也会解除',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await commonApi.addBlacklist(detail.value!.id)
        uni.showToast({ title: '已拉黑', icon: 'none' })
        setTimeout(() => uni.navigateBack(), 800)
      } catch {
        // 已提示
      }
    },
  })
}
</script>

<template>
  <view v-if="detail" class="page">
    <!-- 头图 -->
    <view class="hero">
      <view class="hero-media" @click="previewPhoto(0)">
        <UserAvatar :src="detail.avatar" :name="detail.nickname" fill radius="none" />
        <view class="hero-veil"></view>
      </view>
      <view class="hero-info">
        <view class="hero-heading">
          <view class="hero-name">{{ detail.nickname }}</view>
          <view v-if="detail.match_rate > 0" class="ring-wrap">
            <MatchRing :value="detail.match_rate" :size="104" :thickness="8" />
          </view>
        </view>
        <view class="hero-meta">
          <text>{{ detail.age }} 岁</text>
          <text>{{ detail.height }} cm</text>
          <text>{{ detail.city }}</text>
        </view>
      </view>
      <view class="more" @click="moreActions">更多</view>
    </view>

    <!-- 关系提示 -->
    <view v-if="detail.liked_me && !detail.is_matched" class="hint">
      TA 已经喜欢你了，回应一下试试？
    </view>

    <!-- 核心条件 -->
    <view class="card">
      <view class="card-title">基本条件</view>
      <view class="grid">
        <view class="grid-item">
          <text class="k">学历</text><text class="v">{{ detail.education_text }}</text>
        </view>
        <view class="grid-item">
          <text class="k">月收入</text><text class="v">{{ detail.income_text }}</text>
        </view>
        <view class="grid-item">
          <text class="k">婚姻状况</text><text class="v">{{ detail.marital_text }}</text>
        </view>
        <view class="grid-item">
          <text class="k">子女</text>
          <text class="v">{{ store.label('has_children', detail.has_children) }}</text>
        </view>
        <view class="grid-item">
          <text class="k">购房</text>
          <text class="v">{{ store.label('house_status', detail.house_status) }}</text>
        </view>
        <view class="grid-item">
          <text class="k">购车</text>
          <text class="v">{{ store.label('car_status', detail.car_status) }}</text>
        </view>
        <view class="grid-item">
          <text class="k">职业</text><text class="v">{{ detail.occupation || '未填写' }}</text>
        </view>
        <view class="grid-item">
          <text class="k">行业</text>
          <text class="v">{{ detail.industry_text || '未填写' }}</text>
        </view>
        <view class="grid-item">
          <text class="k">院校</text><text class="v">{{ detail.school || '未填写' }}</text>
        </view>
        <view class="grid-item">
          <text class="k">现居</text>
          <text class="v">{{ detail.province }} {{ detail.city }}</text>
        </view>
        <view class="grid-item">
          <text class="k">户籍</text>
          <text class="v">{{ detail.home_province }} {{ detail.home_city }}</text>
        </view>
      </view>
    </view>

    <!-- 自我介绍 -->
    <view class="card">
      <view class="card-title">自我介绍</view>
      <view class="intro">{{ detail.self_intro }}</view>
    </view>

    <!-- 三观问答（L2：答免费，看答案需权益） -->
    <view v-if="detail.three_view_answered" class="card">
      <view class="card-title">
        三观问答
        <text v-if="detail.three_view?.match_rate != null && detail.three_view_unlocked" class="card-badge">
          一致率 {{ detail.three_view.match_rate }}%
        </text>
      </view>

      <!-- 已解锁：展示逐题对照 -->
      <template v-if="detail.three_view_unlocked && threeViewRows.length">
        <view v-for="(row, i) in threeViewRows" :key="i" class="tv-row">
          <view class="tv-title">
            <text>{{ row.title }}</text>
            <text v-if="row.same" class="tv-same">一致</text>
          </view>
          <view class="tv-pair">
            <view class="tv-cell">
              <text class="tv-who">TA</text>
              <text class="tv-val">{{ row.theirs }}</text>
            </view>
            <view class="tv-cell">
              <text class="tv-who">你</text>
              <text class="tv-val">{{ row.mine }}</text>
            </view>
          </view>
        </view>
      </template>

      <!-- 未解锁：只告诉「答了几题」，不泄露任何选项 -->
      <view v-else class="tv-locked">
        <view class="locked-line">
          TA 已完成 {{ detail.three_view?.answered || detail.three_view?.total || 0 }} 道三观问答
        </view>
        <view class="locked-hint">邀请好友通过审核后，可查看彼此的三观选择与一致率</view>
        <view class="locked-btn" @click="goInvite">去解锁</view>
      </view>
    </view>

    <!-- 原生家庭（L2，且需对方主动授权） -->
    <view v-if="detail.viewed_family && detail.family" class="card">
      <view class="card-title">原生家庭</view>
      <view class="grid">
        <view v-if="detail.family.parent_occ_text" class="grid-item">
          <text class="k">父母职业</text><text class="v">{{ detail.family.parent_occ_text }}</text>
        </view>
        <view v-if="detail.family.atmosphere_text" class="grid-item">
          <text class="k">家庭氛围</text><text class="v">{{ detail.family.atmosphere_text }}</text>
        </view>
        <view v-if="detail.family.only_child_text" class="grid-item">
          <text class="k">独生子女</text><text class="v">{{ detail.family.only_child_text }}</text>
        </view>
        <view v-if="detail.family.city" class="grid-item">
          <text class="k">家庭所在地</text><text class="v">{{ detail.family.city }}</text>
        </view>
      </view>
      <view v-if="detail.family.note" class="intro family-note">{{ detail.family.note }}</view>
      <view class="family-tip">由 TA 主动授权公开</view>
    </view>

    <!-- 相册 -->
    <view v-if="detail.photo_total > 0" class="card">
      <view class="card-title">相册</view>
      <view class="photos">
        <template v-for="(p, i) in detail.photos" :key="i">
          <image
            v-if="!photoFailed[i]"
            :src="p"
            class="photo"
            mode="aspectFill"
            @click="previewPhoto(i)"
            @error="photoFailed[i] = true"
          />
          <view v-else class="photo photo-fail"><text>加载失败</text></view>
        </template>
        <!-- 被锁住的照片用占位格表示数量，不泄露图片本身 -->
        <view v-for="n in detail.locked_photo_count" :key="'lock' + n" class="photo photo-locked">
          <text class="lock-mark">🔒</text>
        </view>
      </view>
      <view v-if="detail.photos_locked" class="photos-lock-tip" @click="goInvite">
        还有 {{ detail.locked_photo_count }} 张照片，邀请好友通过审核即可查看 ›
      </view>
    </view>

    <view class="safe-tip">
      平台不会以任何形式索要钱财。遇到索要红包、诱导投资等情况请立即举报。
    </view>

    <!-- 底部操作 -->
    <view class="actions">
      <template v-if="detail.is_matched">
        <view class="btn-primary full" @click="goChat">去聊天</view>
      </template>
      <template v-else>
        <view class="btn-plain half" @click="handlePass">跳过</view>
        <HeartButton
          class="half"
          block
          :size="36"
          :liked="detail.is_liked"
          :label="'喜欢 TA'"
          :active-label="'已喜欢'"
          @like="handleLike"
        />
      </template>
    </view>

    <MatchSuccessOverlay
      :visible="matchVisible"
      :my-avatar="store.user?.avatar"
      :peer-avatar="detail.avatar"
      :peer-name="detail.nickname"
      @chat="goChat"
      @close="matchVisible = false"
    />
  </view>

  <view v-else-if="loading" class="detail-skeleton" aria-label="正在加载资料">
    <view class="skeleton skeleton-hero"></view>
    <view class="skeleton-block">
      <view class="skeleton skeleton-title"></view>
      <view class="skeleton skeleton-line"></view>
      <view class="skeleton skeleton-line short"></view>
    </view>
    <view class="skeleton-block compact">
      <view class="skeleton skeleton-title"></view>
      <view class="skeleton skeleton-line"></view>
    </view>
  </view>

  <view v-else class="empty">资料加载失败</view>
</template>

<style lang="scss" scoped>
.page {
  padding-bottom: 180rpx;
}

.detail-skeleton {
  min-height: 100dvh;
  padding-bottom: 80rpx;
}

.skeleton-hero {
  width: 100%;
  height: 600rpx;
}

.skeleton-block {
  margin: 20rpx 24rpx 0;
  padding: 34rpx 28rpx;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  background: $bg-raised;

  &.compact {
    padding-bottom: 40rpx;
  }
}

.skeleton-title {
  width: 32%;
  height: 30rpx;
  margin-bottom: 30rpx;
  border-radius: $radius-sm;
}

.skeleton-line {
  width: 88%;
  height: 22rpx;
  margin-bottom: 20rpx;
  border-radius: $radius-sm;

  &.short {
    width: 62%;
  }
}

.hero {
  position: relative;
  background: $bg-card;
  border-bottom: 1rpx solid $border;
}

.hero-media {
  position: relative;
  width: 100%;
  height: 600rpx;
  overflow: hidden;
}

/* 头图底部压一层极淡的纱，让「更多」浮层与照片之间有一道呼吸 */
.hero-veil {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 180rpx;
  background: linear-gradient(to top, rgba(34, 31, 32, 0.24), rgba(34, 31, 32, 0));
  pointer-events: none;
}

.hero-info {
  padding: 28rpx 28rpx 30rpx;
}

.hero-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  margin-bottom: 12rpx;
}

.hero-name {
  font-size: $font-title;
  font-weight: 600;
  color: $text-main;
  letter-spacing: -1rpx;
}

/* 匹配度环压在头图下沿，跨越图片与信息区，成为视觉锚点 */
.ring-wrap {
  margin-top: -60rpx;
  padding: 5rpx;
  border-radius: 50%;
  background: $bg-card;
  box-shadow: 0 6rpx 20rpx rgba(78, 43, 55, 0.16);
}

.hero-meta {
  display: flex;
  align-items: center;
  gap: 0;
  font-size: $font-mid;
  color: $text-weak;
  font-variant-numeric: tabular-nums;
}

.hero-meta text + text::before {
  content: '/';
  margin: 0 14rpx;
  color: $text-placeholder;
}

.more {
  position: absolute;
  right: 24rpx;
  top: 24rpx;
  min-width: 80rpx;
  height: 56rpx;
  line-height: 56rpx;
  padding: 0 16rpx;
  text-align: center;
  border-radius: $radius;
  border: 1rpx solid rgba(255, 255, 255, 0.22);
  background: rgba(37, 35, 33, 0.62);
  color: #fffaf8;
  font-size: $font-caption;
  backdrop-filter: blur(8rpx);
}

.hint {
  margin: 20rpx 24rpx 0;
  padding: 20rpx 24rpx;
  background: $brand-light;
  color: $brand;
  border-left: 4rpx solid $brand;
  border-radius: $radius-sm;
  font-size: $font-mid;
}

.card {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 30rpx 28rpx;
  margin: 20rpx 24rpx 0;
}

.card-title {
  font-size: $font-section;
  font-weight: 600;
  color: $text-main;
  margin-bottom: 22rpx;
}

.grid {
  display: flex;
  flex-wrap: wrap;
}

.grid-item {
  width: 50%;
  display: flex;
  padding: 14rpx 0;
  font-size: $font-mid;

  .k {
    color: $text-weak;
    width: 130rpx;
    flex-shrink: 0;
  }

  .v {
    color: $text-main;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.intro {
  font-size: $font-body;
  line-height: 1.8;
  color: $text-sub;
  white-space: pre-wrap;
}

.photos {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.photo {
  width: 210rpx;
  height: 210rpx;
  border-radius: $radius;
  background: $bg-subtle;
}

/* 加载失败的照片：明确提示，避免用户以为是还在加载 */
.photo-fail {
  display: flex;
  align-items: center;
  justify-content: center;
  color: $text-placeholder;
  font-size: $font-micro;
}

/* 被权益锁住的照片：只占位不显示内容，避免通过缩略图泄露 */
.photo-locked {
  display: flex;
  align-items: center;
  justify-content: center;
  background: $bg-subtle;
  border: 1rpx dashed $border;

  .lock-mark {
    font-size: 40rpx;
    opacity: 0.42;
  }
}

.photos-lock-tip {
  margin-top: 20rpx;
  padding: 20rpx 22rpx;
  border-radius: $radius-sm;
  background: $brand-light;
  color: $brand;
  font-size: $font-caption;
  line-height: 1.5;

  &:active {
    opacity: 0.82;
  }
}

/* ---- 三观问答 ---- */
.card-badge {
  margin-left: 14rpx;
  padding: 4rpx 14rpx;
  border-radius: $radius-sm;
  background: $brand-light;
  color: $brand;
  font-size: $font-micro;
  font-weight: 500;
}

.tv-row {
  padding: 20rpx 0;
  border-bottom: 1rpx solid $border;

  &:last-of-type {
    border-bottom: none;
  }
}

.tv-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 14rpx;
  font-size: $font-mid;
  color: $text-main;
  line-height: 1.5;
}

.tv-same {
  flex-shrink: 0;
  padding: 3rpx 12rpx;
  border-radius: $radius-sm;
  background: $success-light;
  color: $success;
  font-size: $font-micro;
}

.tv-pair {
  display: flex;
  gap: 14rpx;
}

.tv-cell {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: flex-start;
  gap: 10rpx;
  padding: 14rpx 16rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
}

.tv-who {
  flex-shrink: 0;
  font-size: $font-micro;
  color: $text-placeholder;
  margin-top: 2rpx;
}

.tv-val {
  flex: 1;
  min-width: 0;
  font-size: $font-caption;
  color: $text-sub;
  line-height: 1.5;
}

.tv-locked {
  padding: 8rpx 0;
}

.locked-line {
  font-size: $font-mid;
  color: $text-main;
  margin-bottom: 10rpx;
}

.locked-hint {
  font-size: $font-caption;
  color: $text-placeholder;
  line-height: 1.6;
  margin-bottom: 22rpx;
}

.locked-btn {
  height: 76rpx;
  line-height: 76rpx;
  text-align: center;
  border-radius: $radius;
  background: $brand;
  color: $text-on-accent;
  font-size: $font-mid;
  font-weight: 600;

  &:active {
    opacity: 0.88;
  }
}

.family-note {
  margin-top: 16rpx;
}

.family-tip {
  margin-top: 18rpx;
  font-size: $font-micro;
  color: $text-placeholder;
}

.safe-tip {
  margin: 24rpx;
  font-size: $font-micro;
  color: $text-placeholder;
  line-height: 1.6;
  text-align: center;
}

.actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: $bg-card;
  border-top: 1rpx solid $border;
  display: flex;
  gap: 20rpx;
}

.half {
  flex: 1;
}

.full {
  flex: 1;
}
</style>
