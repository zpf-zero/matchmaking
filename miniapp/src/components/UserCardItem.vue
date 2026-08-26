<script setup lang="ts">
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
    <image :src="user.avatar" class="avatar" mode="aspectFill" lazy-load />

    <view class="info">
      <view class="name-row">
        <text class="name">{{ user.nickname }}</text>
        <text v-if="user.match_rate > 0" class="match">{{ user.match_rate }}% 匹配</text>
      </view>

      <view class="facts">
        <text>{{ user.age }} 岁</text>
        <text class="fact-divider"></text>
        <text>{{ user.height }} cm</text>
      </view>

      <view class="tags">
        <text class="tag">{{ user.education_text }}</text>
        <text class="tag tag-brand">{{ user.income_text }}</text>
      </view>

      <view class="meta">
        <text class="meta-item">{{ user.city }}</text>
        <text class="meta-item">{{ user.marital_text }}</text>
        <text v-if="user.occupation" class="meta-item">{{ user.occupation }}</text>
      </view>

      <view class="active" :class="{ 'with-like': showLike }">{{ user.active_text }}</view>
    </view>

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
  margin-bottom: 26rpx;
  position: relative;
  overflow: hidden;
  box-shadow: $shadow-card;
  transition: transform 220ms cubic-bezier(0.16, 1, 0.3, 1),
    border-color 220ms ease, box-shadow 220ms ease;

  &:active {
    transform: scale(0.986);
    border-color: $border-strong;
    box-shadow: none;
  }
}

.avatar {
  width: 224rpx;
  height: 288rpx;
  border-radius: $radius;
  background: $bg-subtle;
}

.info {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  margin: 6rpx 0 10rpx;
}

.name {
  font-size: 35rpx;
  font-weight: 650;
  color: $text-main;
  letter-spacing: -0.5rpx;
  max-width: 180rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.match {
  font-size: 22rpx;
  color: $brand;
  font-weight: 600;
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
  padding: 5rpx 9rpx;
  border-radius: $radius-sm;
  background: $brand-light;
}

.facts {
  display: flex;
  align-items: center;
  gap: 12rpx;
  color: $text-sub;
  font-size: 25rpx;
  font-variant-numeric: tabular-nums;
  margin-bottom: 14rpx;
}

.fact-divider {
  width: 1rpx;
  height: 20rpx;
  background: $border-strong;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 15rpx;
}

.tag {
  padding: 5rpx 12rpx;
  border-radius: $radius-sm;
  font-size: 22rpx;
  background: $bg-subtle;
  color: $text-sub;
}

.tag-brand {
  background: $brand-light;
  color: $brand;
}

.meta {
  font-size: 24rpx;
  color: $text-sub;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta-item + .meta-item::before {
  content: '/';
  margin: 0 10rpx;
  color: $text-placeholder;
}

.active {
  margin-top: auto;
  font-size: 22rpx;
  color: $text-placeholder;
}

.active.with-like {
  padding-right: 72rpx;
}

.like {
  position: absolute;
  right: 20rpx;
  bottom: 20rpx;
  min-width: 92rpx;
  height: 62rpx;
  line-height: 62rpx;
  padding: 0 18rpx;
  text-align: center;
  border-radius: $radius-sm;
  border: 1rpx solid $brand;
  background: $brand;
  color: $text-on-accent;
  font-size: 23rpx;
  font-weight: 600;
  transition: transform 160ms ease, background-color 160ms ease, color 160ms ease;

  &:active {
    transform: scale(0.92);
  }

  &.on {
    background: $brand-light;
    border-color: $brand;
    color: $brand;
  }
}
</style>
