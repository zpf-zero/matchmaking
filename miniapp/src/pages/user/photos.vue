<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { userApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { uploadImage } from '@/utils/request'

const store = useUserStore()
const uploading = ref(false)
const MAX = 9

onShow(() => {
  store.fetchMe()
})

function auditText(s: number) {
  return ['审核中', '', '未通过'][s] || ''
}

async function add() {
  if (store.photos.length >= MAX) {
    uni.showToast({ title: `最多 ${MAX} 张`, icon: 'none' })
    return
  }

  uni.chooseImage({
    count: MAX - store.photos.length,
    sizeType: ['compressed'],
    success: async (res) => {
      const paths = res.tempFilePaths as string[]
      uploading.value = true
      uni.showLoading({ title: '上传中' })
      try {
        for (const p of paths) {
          const url = await uploadImage(p, 'photo')
          await userApi.addPhoto(url)
        }
        await store.fetchMe()
      } catch {
        // 已提示
      } finally {
        uni.hideLoading()
        uploading.value = false
      }
    },
  })
}

function remove(id: number) {
  uni.showModal({
    title: '删除照片',
    content: '确定删除这张照片吗？',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await userApi.deletePhoto(id)
        await store.fetchMe()
      } catch {
        // 已提示
      }
    },
  })
}

function preview(i: number) {
  uni.previewImage({
    urls: store.photos.map((p) => p.url),
    current: i,
  })
}
</script>

<template>
  <view class="page">
    <view class="tip">
      上传真实生活照更容易获得回应。所有照片需经人工审核后才会展示给他人。
    </view>

    <view class="grid">
      <view v-for="(p, i) in store.photos" :key="p.id" class="item">
        <image :src="p.url" class="img" mode="aspectFill" @click="preview(i)" />
        <view v-if="p.audit_status !== 1" class="mask">{{ auditText(p.audit_status) }}</view>
        <view class="del" @click="remove(p.id)">×</view>
      </view>

      <view v-if="store.photos.length < MAX" class="item add" @click="add">
        <text class="plus">＋</text>
        <text class="add-text">添加照片</text>
      </view>
    </view>

    <view class="count">{{ store.photos.length }}/{{ MAX }}</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 24rpx;
}

.tip {
  font-size: 24rpx;
  color: #8a8f99;
  line-height: 1.6;
  padding: 0 8rpx 24rpx;
}

.grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.item {
  position: relative;
  width: 220rpx;
  height: 220rpx;
  border-radius: 12rpx;
  overflow: hidden;
  background: #fff;
}

.img {
  width: 100%;
  height: 100%;
}

.mask {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 44rpx;
  line-height: 44rpx;
  text-align: center;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  font-size: 22rpx;
}

.del {
  position: absolute;
  right: 0;
  top: 0;
  width: 48rpx;
  height: 48rpx;
  line-height: 44rpx;
  text-align: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 32rpx;
  border-bottom-left-radius: 12rpx;
}

.add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1rpx dashed #d9dce0;
  background: #fff;
}

.plus {
  font-size: 52rpx;
  color: #c9cdd4;
  line-height: 1;
  margin-bottom: 12rpx;
}

.add-text {
  font-size: 22rpx;
  color: #bbbfc4;
}

.count {
  text-align: center;
  font-size: 24rpx;
  color: #bbbfc4;
  padding: 32rpx 0;
}
</style>
