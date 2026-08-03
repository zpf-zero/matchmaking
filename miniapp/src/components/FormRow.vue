<script setup lang="ts">
defineProps<{
  label: string
  required?: boolean
  /** 值区域是否右对齐（选择器类用右对齐，输入框用左对齐） */
  alignRight?: boolean
}>()
</script>

<template>
  <view class="form-row">
    <view class="label">
      <text>{{ label }}</text>
      <text v-if="required" class="req">*</text>
    </view>
    <view class="content" :class="{ right: alignRight }">
      <slot />
    </view>
  </view>
</template>

<style lang="scss" scoped>
.form-row {
  display: flex;
  align-items: center;
  min-height: 100rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f2f3f5;

  &:last-child {
    border-bottom: none;
  }
}

.label {
  width: 180rpx;
  flex-shrink: 0;
  font-size: 28rpx;
  color: #1f2329;

  .req {
    color: #f53f3f;
    margin-left: 4rpx;
  }
}

.content {
  flex: 1;
  font-size: 28rpx;
  color: #1f2329;
  overflow: hidden;

  &.right {
    text-align: right;
  }
}
</style>
