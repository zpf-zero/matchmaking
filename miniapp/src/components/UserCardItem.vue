<script setup lang="ts">
import { computed } from 'vue'
import UserAvatar from '@/components/UserAvatar.vue'
import type { UserCard } from '@/api/types'

const props = defineProps<{
  user: UserCard
  /** 是否展示右侧喜欢按钮 */
  showLike?: boolean
  liked?: boolean
}>()

const emit = defineEmits<{
  like: [user: UserCard]
}>()

/** 婚姻状况与职业合成一行，空值不留分隔符 */
const metaText = computed(() =>
  [props.user.marital_text, props.user.occupation].filter(Boolean).join(' · ')
)

function goDetail() {
  uni.navigateTo({ url: `/pages/user/detail?id=${props.user.id}` })
}
</script>

<template>
  <view
    class="user-card"
    role="button"
    tabindex="0"
    :aria-label="`查看 ${user.nickname} 的资料`"
    @click="goDetail"
    @keyup.enter="goDetail"
  >
    <UserAvatar :src="user.avatar" :name="user.nickname" :size="224" />

    <view class="info">
      <view class="head">
        <text class="name">{{ user.nickname }}</text>
        <text v-if="user.match_rate > 0" class="match">{{ user.match_rate }}% 匹配</text>
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
        <view
          v-if="showLike"
          class="like"
          :class="{ on: liked }"
          role="button"
          tabindex="0"
          :aria-label="liked ? '已喜欢' : '喜欢'"
          @click.stop="emit('like', user)"
          @keyup.enter.stop="emit('like', user)"
        >
          {{ liked ? '已喜欢' : '喜欢' }}
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.user-card {
  display: grid;
  grid-template-columns: 224rpx minmax(0, 1fr);
  gap: 24rpx;
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 18rpx;
  margin-bottom: 24rpx;
  box-shadow: $shadow-card;
  transition: transform 220ms cubic-bezier(0.16, 1, 0.3, 1),
    border-color 220ms ease, box-shadow 220ms ease;

  &:active {
    transform: scale(0.986);
    border-color: $border-strong;
    box-shadow: none;
  }
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

.match {
  flex-shrink: 0;
  padding: 4rpx 10rpx;
  border-radius: $radius-sm;
  background: $brand-light;
  color: $brand;
  font-size: $font-micro;
  font-weight: 500;
  font-variant-numeric: tabular-nums;
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

.like {
  flex-shrink: 0;
  min-width: 132rpx;
  height: 64rpx;
  line-height: 62rpx;
  padding: 0 22rpx;
  text-align: center;
  border-radius: $radius;
  border: 1rpx solid $brand;
  background: $brand;
  color: $text-on-accent;
  font-size: $font-caption;
  font-weight: 600;
  transition: transform 160ms ease, background-color 160ms ease, color 160ms ease;

  &:active {
    transform: scale(0.94);
  }

  &.on {
    background: $brand-light;
    color: $brand;
  }
}
</style>
