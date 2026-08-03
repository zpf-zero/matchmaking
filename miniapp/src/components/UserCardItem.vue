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
  <view class="user-card" @click="goDetail">
    <image :src="user.avatar" class="avatar" mode="aspectFill" />

    <view class="info">
      <view class="name-row">
        <text class="name">{{ user.nickname }}</text>
        <text v-if="user.match_rate > 0" class="match">匹配 {{ user.match_rate }}%</text>
      </view>

      <view class="tags">
        <text class="tag">{{ user.age }}岁</text>
        <text class="tag">{{ user.height }}cm</text>
        <text class="tag">{{ user.education_text }}</text>
        <text class="tag tag-brand">{{ user.income_text }}</text>
      </view>

      <view class="meta">
        <text>{{ user.city }}</text>
        <text class="sep">·</text>
        <text>{{ user.marital_text }}</text>
        <text v-if="user.occupation" class="sep">·</text>
        <text v-if="user.occupation">{{ user.occupation }}</text>
      </view>

      <view class="active">{{ user.active_text }}</view>
    </view>

    <view v-if="showLike" class="like" :class="{ on: liked }" @click.stop="emit('like', user)">
      {{ liked ? '♥' : '♡' }}
    </view>
  </view>
</template>

<style lang="scss" scoped>
.user-card {
  display: flex;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  position: relative;
}

.avatar {
  width: 160rpx;
  height: 200rpx;
  border-radius: 12rpx;
  background: #f5f6f8;
  flex-shrink: 0;
}

.info {
  flex: 1;
  margin-left: 24rpx;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.name-row {
  display: flex;
  align-items: center;
  margin-bottom: 14rpx;
}

.name {
  font-size: 32rpx;
  font-weight: 600;
  color: #1f2329;
  max-width: 240rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.match {
  margin-left: 14rpx;
  font-size: 22rpx;
  color: #e8506d;
  background: #fff0f3;
  padding: 2rpx 12rpx;
  border-radius: 6rpx;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 14rpx;
}

.tag {
  padding: 4rpx 14rpx;
  border-radius: 6rpx;
  font-size: 22rpx;
  background: #f5f6f8;
  color: #646a73;
}

.tag-brand {
  background: #fff0f3;
  color: #e8506d;
}

.meta {
  font-size: 24rpx;
  color: #646a73;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sep {
  margin: 0 8rpx;
  color: #d9dce0;
}

.active {
  margin-top: auto;
  font-size: 22rpx;
  color: #bbbfc4;
}

.like {
  position: absolute;
  right: 24rpx;
  bottom: 24rpx;
  width: 64rpx;
  height: 64rpx;
  line-height: 64rpx;
  text-align: center;
  border-radius: 50%;
  background: #f5f6f8;
  color: #8a8f99;
  font-size: 34rpx;

  &.on {
    background: #fff0f3;
    color: #e8506d;
  }
}
</style>
