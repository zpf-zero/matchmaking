<script setup lang="ts">
import { computed } from 'vue'
import UserAvatar from '@/components/UserAvatar.vue'
import HeartButton from '@/components/HeartButton.vue'
import MatchRing from '@/components/MatchRing.vue'
import type { UserCard } from '@/api/types'

const props = defineProps<{
  user: UserCard
  /** 是否展示右侧喜欢按钮 */
  showLike?: boolean
  liked?: boolean
  /** 正在离场：淡出并收起高度，用于喜欢后移除卡片 */
  leaving?: boolean
  /** 入场动画的错峰序号 */
  index?: number
}>()

const emit = defineEmits<{
  like: [user: UserCard]
}>()

/** 婚姻状况与职业合成一行，空值不留分隔符 */
const metaText = computed(() =>
  [props.user.marital_text, props.user.occupation].filter(Boolean).join(' · ')
)

/** 只给首屏前几张错峰，后面的直接显示，避免长列表末尾等待 */
const enterDelay = computed(() => {
  const i = props.index ?? 0
  return i < 6 ? `${i * 55}ms` : '0ms'
})

const wrapStyle = computed(() => ({
  animationDelay: enterDelay.value,
  ...(props.leaving ? { animationDelay: '0ms' } : {}),
}))

function goDetail() {
  if (props.leaving) return
  uni.navigateTo({ url: `/pages/user/detail?id=${props.user.id}` })
}
</script>

<template>
  <view
    class="card-wrap"
    :class="{ leaving }"
    :style="wrapStyle"
    role="button"
    tabindex="0"
    :aria-label="`查看 ${user.nickname} 的资料`"
    @click="goDetail"
    @keyup.enter="goDetail"
  >
    <view class="user-card">
      <view class="avatar-slot">
        <UserAvatar :src="user.avatar" :name="user.nickname" fill radius="md" />
        <view v-if="user.match_rate > 0" class="ring-slot">
          <MatchRing :value="user.match_rate" :size="76" :thickness="6" />
        </view>
      </view>

      <view class="info">
        <view class="head">
          <text class="name">{{ user.nickname }}</text>
        </view>

        <view class="facts">
          <text class="fact">{{ user.age }} 岁</text>
          <text class="fact-sep"></text>
          <text class="fact">{{ user.height }} cm</text>
          <text class="fact-sep"></text>
          <text class="fact">{{ user.city }}</text>
        </view>

        <view class="tags">
          <text class="tag">{{ user.education_text }}</text>
          <text class="tag tag-brand">{{ user.income_text }}</text>
        </view>

        <view class="meta">{{ metaText }}</view>

        <view class="foot">
          <text class="active">{{ user.active_text }}</text>
          <HeartButton
            v-if="showLike"
            :liked="liked"
            :size="34"
            @like="emit('like', user)"
          />
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.card-wrap {
  margin-bottom: 24rpx;
  overflow: hidden;
  animation: card-enter 460ms cubic-bezier(0.16, 1, 0.3, 1) both;
}

/* 喜欢之后淡出并收起高度，比瞬间消失更容易被感知为「处理完了」 */
.leaving {
  animation: card-leave 380ms cubic-bezier(0.4, 0, 1, 1) both;
  pointer-events: none;
}

.user-card {
  display: grid;
  grid-template-columns: 224rpx minmax(0, 1fr);
  gap: 24rpx;
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 18rpx;
  box-shadow: $shadow-card;
  transition: transform 220ms cubic-bezier(0.16, 1, 0.3, 1),
    border-color 220ms ease, box-shadow 220ms ease;

  &:active {
    transform: scale(0.986);
    border-color: $border-strong;
    box-shadow: none;
  }
}

.avatar-slot {
  position: relative;
  width: 224rpx;
  height: 288rpx;
}

/* 匹配度环压在头像右下角，替掉原来抢眼的小色块 */
.ring-slot {
  position: absolute;
  right: -16rpx;
  bottom: -16rpx;
  padding: 4rpx;
  border-radius: 50%;
  background: $bg-raised;
  box-shadow: 0 4rpx 14rpx rgba(78, 43, 55, 0.14);
}

.info {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12rpx;
  margin: 8rpx 0 12rpx;
}

.name {
  min-width: 0;
  font-size: $font-title;
  font-weight: 650;
  color: $text-main;
  letter-spacing: -0.6rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.facts {
  display: flex;
  align-items: center;
  gap: 12rpx;
  color: $text-sub;
  font-size: $font-mid;
  font-variant-numeric: tabular-nums;
  margin-bottom: 14rpx;
}

.fact {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fact-sep {
  width: 1rpx;
  height: 20rpx;
  flex-shrink: 0;
  background: $border-strong;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 14rpx;
}

.tag {
  padding: 5rpx 12rpx;
  border-radius: $radius-sm;
  font-size: $font-micro;
  background: $bg-subtle;
  color: $text-sub;
}

.tag-brand {
  background: $brand-light;
  color: $brand;
}

.meta {
  font-size: $font-caption;
  color: $text-sub;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.foot {
  margin-top: auto;
  padding-top: 16rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.active {
  min-width: 0;
  font-size: $font-micro;
  color: $text-placeholder;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@keyframes card-enter {
  from {
    opacity: 0;
    transform: translateY(20rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes card-leave {
  0% {
    opacity: 1;
    max-height: 400rpx;
    transform: scale(1);
  }
  55% {
    opacity: 0;
    transform: scale(0.96);
  }
  100% {
    opacity: 0;
    max-height: 0;
    margin-bottom: 0;
    padding: 0;
    transform: scale(0.96);
  }
}
</style>
