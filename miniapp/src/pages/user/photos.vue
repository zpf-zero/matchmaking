<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { userApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { uploadImage } from '@/utils/request'

const store = useUserStore()
const uploading = ref(false)
const MAX = 9

/** 正在上传的本地图片路径。多选时一次性建好占位格，
 *  用户能看见「有几张、到第几张」，而不是一句转圈等到底 */
const pending = ref<string[]>([])
/** 已经上传完成的张数，用于算整体进度 */
const doneCount = ref(0)
/** 已加载完成的图片 url，做加载占位 */
const loaded = ref<Record<string, boolean>>({})
/** 加载失败的图片 url。失败后换成明确提示，不留灰块、也不无限转圈 */
const failed = ref<Record<string, boolean>>({})

onShow(() => {
  store.fetchMe()
})

function auditText(s: number) {
  return ['审核中', '', '未通过'][s] || ''
}

async function add() {
  if (uploading.value) return // 上传尚未走完时再点会起第二个上传循环
  if (store.photos.length + pending.value.length >= MAX) {
    uni.showToast({ title: `最多 ${MAX} 张`, icon: 'none' })
    return
  }

  uni.chooseImage({
    count: MAX - store.photos.length - pending.value.length,
    sizeType: ['compressed'],
    success: async (res) => {
      const paths = res.tempFilePaths as string[]
      uploading.value = true
      pending.value = [...paths]
      doneCount.value = 0

      try {
        for (const p of paths) {
          const url = await uploadImage(p, 'photo')
          await userApi.addPhoto(url)
          // 每传完一张就把它从占位队列里摘掉，进度肉眼可见
          pending.value = pending.value.filter((x) => x !== p)
          doneCount.value++
        }
        await store.fetchMe()
      } catch {
        // 已提示
      } finally {
        pending.value = []
        doneCount.value = 0
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
    <view v-if="!store.isVip && store.photos.length > store.freePhotoLimit" class="tip tip-lock">
      前 {{ store.freePhotoLimit }} 张对所有人可见，第 {{ store.freePhotoLimit + 1 }} 张起仅对已解锁的用户展示，其余人看到的是锁定状态。
    </view>

    <!-- 上传进度：多张时给出「第几张 / 共几张」 -->
    <view v-if="pending.length" class="progress">
      <view class="progress-bar" :style="{ width: (doneCount / (doneCount + pending.length)) * 100 + '%' }"></view>
      <text class="progress-text">正在上传 {{ doneCount + 1 }}/{{ doneCount + pending.length }}</text>
    </view>

    <view class="grid">
      <view
        v-for="(p, i) in store.photos"
        :key="p.id"
        class="item photo"
        :class="{ ready: loaded[p.url] }"
      >
        <image
          v-if="!failed[p.url]"
          :src="p.url"
          class="img"
          mode="aspectFill"
          @click="preview(i)"
          @load="loaded[p.url] = true"
          @error="failed[p.url] = true"
        />
        <view v-else class="img-fail">
          <text>图片加载失败</text>
        </view>
        <view v-if="p.audit_status !== 1" class="mask">{{ auditText(p.audit_status) }}</view>
        <view class="del" @click="remove(p.id)">×</view>
      </view>

      <!-- 上传中的占位格：先用本地图铺底 + 半透明遮罩，传完自然替换成正式图 -->
      <view v-for="(path, idx) in pending" :key="'p' + idx" class="item uploading-tile">
        <image :src="path" class="img pending-img" mode="aspectFill" />
        <view class="tile-veil">
          <view class="spinner"></view>
        </view>
      </view>

      <view v-if="store.photos.length + pending.length < MAX" class="item add" @click="add">
        <text class="plus">＋</text>
        <text class="add-text">添加照片</text>
      </view>
    </view>

    <view class="count">{{ store.photos.length }}/{{ MAX }}</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.tip {
  font-size: $font-caption;
  color: $text-weak;
  line-height: 1.6;
  padding: 0 8rpx 24rpx;
}

/* 照片分层说明。用浅色底强调一下，避免和上一句连成一片被略过 */
.tip-lock {
  margin: -10rpx 8rpx 24rpx;
  padding: 18rpx 20rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
  color: $text-placeholder;
}

/* ---- 上传进度条 ---- */
.progress {
  position: relative;
  height: 56rpx;
  margin: 0 8rpx 24rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.progress-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  /* 起步阶段进度条宽度接近 0，给个下限让它始终可见，
     否则用户看不出「到底有没有在动」 */
  min-width: 8%;
  background: $brand-light;
  transition: width 320ms cubic-bezier(0.22, 0.8, 0.3, 1);
}

.progress-text {
  position: relative;
  width: 100%;
  text-align: center;
  font-size: $font-micro;
  color: $brand;
  font-variant-numeric: tabular-nums;
}

.grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
}

.item {
  position: relative;
  width: 100%;
  height: 220rpx;
  border-radius: $radius;
  overflow: hidden;
  background: $bg-card;
  border: 1rpx solid $border;
}

/* 图片未加载完前先露出底色，避免白块闪一下 */
.photo {
  background: $bg-subtle;

  .img {
    opacity: 0;
    transition: opacity 320ms ease;
  }

  &.ready .img {
    opacity: 1;
  }
}

.img {
  width: 100%;
  height: 100%;
}

/* 图片加载失败：给出明确提示，替代原来的灰块 */
.img-fail {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $bg-subtle;
  color: $text-placeholder;
  font-size: $font-micro;
}

/* ---- 上传中的占位格 ---- */
.uploading-tile {
  background: $bg-subtle;
  animation: tile-in 300ms cubic-bezier(0.34, 1.4, 0.5, 1) both;
}

.pending-img {
  filter: blur(6rpx) saturate(0.7);
  transform: scale(1.06); /* 模糊后边缘会露白，放大一点盖住 */
}

.tile-veil {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.42);
  backdrop-filter: blur(2rpx);
}

.spinner {
  width: 46rpx;
  height: 46rpx;
  border-radius: 50%;
  border: 4rpx solid rgba(150, 62, 88, 0.22);
  border-top-color: $brand;
  animation: spin 720ms linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes tile-in {
  from {
    transform: scale(0.86);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
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
  color: #fffaf8;
  font-size: $font-micro;
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
  color: #fffaf8;
  font-size: 32rpx;
  border-bottom-left-radius: $radius;
}

.add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1rpx dashed $border-strong;
  background: $bg-card;
  transition: transform 180ms ease, background-color 180ms ease;

  &:active {
    transform: scale(0.98);
    background: $bg-subtle;
  }
}

.plus {
  font-size: 52rpx;
  color: $text-placeholder;
  line-height: 1;
  margin-bottom: 12rpx;
}

.add-text {
  font-size: $font-micro;
  color: $text-placeholder;
}

.count {
  text-align: center;
  font-size: $font-caption;
  color: $text-placeholder;
  padding: 36rpx 0;
  font-variant-numeric: tabular-nums;
}
</style>
