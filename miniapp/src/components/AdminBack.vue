<script setup lang="ts">
/**
 * 后台子页顶部的返回条。
 *
 * 管理页不是 tabBar 页，走 navigateTo 进来时靠系统返回键即可；
 * 但冷启动深链（或 reLaunch 落地）时没有返回栈，会卡在这一页出不去。
 * 这里统一给一个「工作台」入口：有栈就回退，没栈就回工作台。
 */
const props = withDefaults(defineProps<{ to?: string; label?: string }>(), {
  to: '/pages/admin/index',
  label: '工作台',
})

function back() {
  const pages = getCurrentPages()
  if (pages.length > 1) {
    uni.navigateBack()
  } else {
    uni.reLaunch({ url: props.to })
  }
}
</script>

<template>
  <view class="admin-back" @click="back">
    <view class="chev"></view>
    <text class="txt">{{ label }}</text>
  </view>
</template>

<style lang="scss" scoped>
.admin-back {
  display: inline-flex;
  align-items: center;
  gap: 10rpx;
  padding: 10rpx 20rpx 10rpx 16rpx;
  margin-bottom: 20rpx;
  border-radius: $radius-round;
  background: $bg-subtle;
  transition: background-color 180ms ease;

  &:active {
    background: $bg-card;
  }
}

.chev {
  width: 14rpx;
  height: 14rpx;
  border-left: 2rpx solid $text-sub;
  border-bottom: 2rpx solid $text-sub;
  transform: rotate(45deg);
}

.txt {
  font-size: $font-caption;
  color: $text-sub;
}
</style>
