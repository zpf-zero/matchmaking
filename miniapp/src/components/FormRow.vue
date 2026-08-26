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
  min-height: 112rpx;
  padding: 20rpx 0;
  border-bottom: 1rpx solid $border;

  &:last-child {
    border-bottom: none;
  }
}

.label {
  width: 180rpx;
  flex-shrink: 0;
  font-size: 26rpx;
  font-weight: 550;
  color: $text-main;

  .req {
    color: $danger;
    margin-left: 4rpx;
  }
}

.content {
  flex: 1;
  font-size: 27rpx;
  color: $text-main;
  overflow: hidden;

  &.right {
    text-align: right;
  }
}

@media (max-width: 340px) {
  .label {
    width: 156rpx;
  }
}
</style>
