<script setup lang="ts">
import { ref } from 'vue'
import { onShow, onReachBottom } from '@dcloudio/uni-app'
import UserAvatar from '@/components/UserAvatar.vue'
import { adminApi } from '@/api'
import type { AuditItem } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const list = ref<AuditItem[]>([])
const page = ref(1)
const total = ref(0)
const finished = ref(false)
const loading = ref(false)
const loaded = ref(false)

onShow(async () => {
  await store.fetchDict()
  load(true)
})

onReachBottom(() => {
  if (!finished.value) load(false)
})

async function load(reset: boolean) {
  if (loading.value) return
  loading.value = true
  if (reset) {
    page.value = 1
    finished.value = false
  }

  try {
    const res = await adminApi.auditList(page.value)
    total.value = res.total
    list.value = reset ? res.list || [] : [...list.value, ...(res.list || [])]
    finished.value = list.value.length >= res.total
    if (!finished.value) page.value++
    loaded.value = true
  } catch {
    loaded.value = true
  } finally {
    loading.value = false
  }
}

function previewAll(item: AuditItem, index: number) {
  const urls = [item.avatar, ...item.photos.map((p) => p.url)].filter(Boolean)
  uni.previewImage({ urls, current: index })
}

async function pass(item: AuditItem) {
  try {
    await adminApi.audit(item.user_id, true)
    list.value = list.value.filter((i) => i.user_id !== item.user_id)
    total.value = Math.max(0, total.value - 1)
    uni.showToast({ title: '已通过', icon: 'none' })
  } catch {
    // 已提示
  }
}

function reject(item: AuditItem) {
  const presets = ['头像非真人照片', '资料信息不真实', '自我介绍含联系方式', '照片含违规内容', '自定义原因']
  uni.showActionSheet({
    itemList: presets,
    success: (res) => {
      if (res.tapIndex === presets.length - 1) {
        customReject(item)
      } else {
        doReject(item, presets[res.tapIndex])
      }
    },
  })
}

function customReject(item: AuditItem) {
  // 小程序没有带输入框的 modal，用 editable 版本
  uni.showModal({
    title: '拒绝原因',
    editable: true,
    placeholderText: '请填写具体原因，用户会看到',
    success: (r) => {
      if (r.confirm && r.content?.trim()) {
        doReject(item, r.content.trim())
      }
    },
  })
}

async function doReject(item: AuditItem, remark: string) {
  try {
    await adminApi.audit(item.user_id, false, remark)
    list.value = list.value.filter((i) => i.user_id !== item.user_id)
    total.value = Math.max(0, total.value - 1)
    uni.showToast({ title: '已拒绝', icon: 'none' })
  } catch {
    // 已提示
  }
}

function calcAge(item: AuditItem) {
  return item.profile?.age || 0
}
</script>

<template>
  <view class="page">
    <view class="head">待审核 {{ total }} 人</view>

    <view v-for="item in list" :key="item.user_id" class="card">
      <!-- 基本信息 -->
      <view class="top">
        <view class="avatar-wrap" @click="previewAll(item, 0)">
          <UserAvatar :src="item.avatar" :name="item.nickname" :size="140" />
        </view>
        <view class="base">
          <view class="name">
            {{ item.nickname }}
            <text class="gender">{{ store.label('gender', item.gender) }}</text>
          </view>
          <view class="sub">{{ calcAge(item) }}岁 · {{ item.profile?.height }}cm · {{ item.profile?.city }}</view>
          <view class="sub">手机 {{ item.phone || '未绑定' }}</view>
        </view>
      </view>

      <!-- 条件 -->
      <view class="fields">
        <view class="f">
          <text class="fk">学历</text>
          <text class="fv">{{ store.label('education', item.profile?.education) }}</text>
        </view>
        <view class="f">
          <text class="fk">收入</text>
          <text class="fv">{{ store.label('income_level', item.profile?.income_level) }}</text>
        </view>
        <view class="f">
          <text class="fk">婚况</text>
          <text class="fv">{{ store.label('marital_status', item.profile?.marital_status) }}</text>
        </view>
        <view class="f">
          <text class="fk">子女</text>
          <text class="fv">{{ store.label('has_children', item.profile?.has_children) }}</text>
        </view>
        <view class="f">
          <text class="fk">职业</text>
          <text class="fv">{{ item.profile?.occupation }}</text>
        </view>
        <view class="f">
          <text class="fk">院校</text>
          <text class="fv">{{ item.profile?.school || '未填写' }}</text>
        </view>
      </view>

      <!-- 自我介绍 -->
      <view class="intro">{{ item.profile?.self_intro }}</view>

      <!-- 相册 -->
      <view v-if="item.photos?.length" class="photos">
        <image
          v-for="(p, i) in item.photos"
          :key="p.id"
          :src="p.url"
          class="photo"
          mode="aspectFill"
          @click="previewAll(item, i + 1)"
        />
      </view>

      <view class="actions">
        <view class="btn reject" @click="reject(item)">拒绝</view>
        <view class="btn pass" @click="pass(item)">通过</view>
      </view>
    </view>

    <view v-if="loaded && !list.length" class="empty">没有待审核的资料</view>
    <view v-else-if="list.length && finished" class="end">已经到底啦</view>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 48rpx;
}

.head {
  font-size: $font-mid;
  font-variant-numeric: tabular-nums;
  color: $text-weak;
  padding: 0 8rpx 20rpx;
}

.card {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 28rpx;
  margin-bottom: 24rpx;
}

.top {
  display: flex;
  margin-bottom: 20rpx;
}

.avatar-wrap {
  flex-shrink: 0;
}

.base {
  flex: 1;
  margin-left: 24rpx;
}

.name {
  font-size: $font-section;
  font-weight: 600;
  margin-bottom: 12rpx;
}

.gender {
  font-size: $font-micro;
  color: $text-weak;
  font-weight: normal;
  margin-left: 10rpx;
}

.sub {
  font-size: $font-caption;
  color: $text-weak;
  margin-bottom: 6rpx;
}

.fields {
  display: flex;
  flex-wrap: wrap;
  padding: 16rpx 0;
  border-top: 1rpx solid $border;
}

.f {
  width: 50%;
  display: flex;
  padding: 8rpx 0;
  font-size: $font-mid;

  .fk {
    color: $text-weak;
    width: 90rpx;
    flex-shrink: 0;
  }

  .fv {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.intro {
  font-size: $font-mid;
  color: $text-sub;
  line-height: 1.7;
  background: $bg-subtle;
  border-radius: $radius;
  padding: 20rpx;
  margin: 12rpx 0 20rpx;
}

.photos {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 24rpx;
}

.photo {
  width: 150rpx;
  height: 150rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
}

.actions {
  display: flex;
  gap: 20rpx;
  padding-top: 8rpx;
}

.btn {
  flex: 1;
  height: 80rpx;
  line-height: 80rpx;
  text-align: center;
  border-radius: $radius;
  font-size: $font-body;
  font-weight: 600;
  transition: transform 180ms cubic-bezier(0.16, 1, 0.3, 1), background-color 180ms ease;

  &:active {
    transform: scale(0.98);
  }

  &.reject {
    background: $bg-card;
    border: 1rpx solid $border;
    color: $text-sub;
  }

  &.pass {
    background: $brand;
    color: $text-on-accent;
  }
}

.end {
  text-align: center;
  color: $text-placeholder;
  font-size: $font-caption;
  padding: 30rpx 0;
}
</style>
