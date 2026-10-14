<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import UserAvatar from '@/components/UserAvatar.vue'
import { adminApi } from '@/api'
import { Role, useUserStore } from '@/stores/user'

const store = useUserStore()
const uid = ref(0)
const data = ref<any>(null)
const loaded = ref(false)

const statusText = ['待完善', '待审核', '正常', '未通过', '已封禁']
const reasonText = ['', '虚假资料', '骚扰', '广告', '违法信息', '其他']

onLoad(async (q) => {
  uid.value = Number(q?.id || 0)
  await store.fetchDict()
  await load()
})

async function load() {
  try {
    data.value = await adminApi.userDetail(uid.value)
  } catch {
    // 已提示
  } finally {
    loaded.value = true
  }
}

function ban() {
  uni.showModal({
    title: '封禁用户',
    editable: true,
    placeholderText: '封禁原因，将展示给该用户',
    success: async (r) => {
      if (!r.confirm) return
      const reason = r.content?.trim()
      if (!reason) {
        uni.showToast({ title: '请填写封禁原因', icon: 'none' })
        return
      }
      try {
        await adminApi.ban(uid.value, reason)
        uni.showToast({ title: '已封禁', icon: 'none' })
        load()
      } catch {
        // 已提示
      }
    },
  })
}

function unban() {
  uni.showModal({
    title: '解除封禁',
    content: '解封后该用户资料需重新通过审核才能正常使用',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await adminApi.unban(uid.value)
        uni.showToast({ title: '已解封', icon: 'none' })
        load()
      } catch {
        // 已提示
      }
    },
  })
}

function preview(urls: string[], i: number) {
  uni.previewImage({ urls, current: i })
}
</script>

<template>
  <view v-if="data" class="page">
    <view class="card">
      <view class="top">
        <view class="avatar-wrap" @click="preview([data.user.avatar], 0)">
          <UserAvatar :src="data.user.avatar" :name="data.user.nickname" :size="130" />
        </view>
        <view class="base">
          <view class="name">
            {{ data.user.nickname || '未填写' }}
            <text class="uid">#{{ data.user.id }}</text>
          </view>
          <view class="sub">{{ store.label('gender', data.user.gender) }} / {{ data.phone || '未绑定' }}</view>
          <view class="status" :class="'s' + data.user.status">
            {{ statusText[data.user.status] }}
          </view>
        </view>
      </view>

      <view v-if="data.user.ban_reason" class="ban-reason">
        封禁原因：{{ data.user.ban_reason }}
      </view>
    </view>

    <!-- 数据 -->
    <view class="card stat">
      <view class="s">
        <text class="n">{{ data.stats.like_count }}</text><text class="l">发出的喜欢</text>
      </view>
      <view class="s">
        <text class="n">{{ data.stats.liked_count }}</text><text class="l">收到的喜欢</text>
      </view>
      <view class="s">
        <text class="n">{{ data.stats.match_count }}</text><text class="l">配对数</text>
      </view>
    </view>

    <!-- 资料 -->
    <view v-if="data.profile?.id" class="card">
      <view class="ct">资料信息</view>
      <view class="fields">
        <view class="f"><text class="fk">年龄</text><text class="fv">{{ data.profile.age }}岁</text></view>
        <view class="f"><text class="fk">身高</text><text class="fv">{{ data.profile.height }}cm</text></view>
        <view class="f">
          <text class="fk">学历</text>
          <text class="fv">{{ store.label('education', data.profile.education) }}</text>
        </view>
        <view class="f">
          <text class="fk">收入</text>
          <text class="fv">{{ store.label('income_level', data.profile.income_level) }}</text>
        </view>
        <view class="f">
          <text class="fk">婚况</text>
          <text class="fv">{{ store.label('marital_status', data.profile.marital_status) }}</text>
        </view>
        <view class="f">
          <text class="fk">子女</text>
          <text class="fv">{{ store.label('has_children', data.profile.has_children) }}</text>
        </view>
        <view class="f"><text class="fk">职业</text><text class="fv">{{ data.profile.occupation }}</text></view>
        <view class="f"><text class="fk">城市</text><text class="fv">{{ data.profile.city }}</text></view>
      </view>
      <view class="intro">{{ data.profile.self_intro }}</view>
    </view>

    <!-- 相册 -->
    <view v-if="data.photos?.length" class="card">
      <view class="ct">相册（{{ data.photos.length }}）</view>
      <view class="photos">
        <image
          v-for="(p, i) in data.photos"
          :key="p.id"
          :src="p.url"
          class="photo"
          mode="aspectFill"
          @click="preview(data.photos.map((x: any) => x.url), i)"
        />
      </view>
    </view>

    <!-- 举报记录 -->
    <view v-if="data.reports?.length" class="card">
      <view class="ct">被举报记录（{{ data.reports.length }}）</view>
      <view v-for="r in data.reports" :key="r.id" class="report-item">
        <text class="r-reason">{{ reasonText[r.reason] }}</text>
        <text class="r-desc">{{ r.description || '无描述' }}</text>
        <text class="r-status">{{ ['待处理', '已处理', '已驳回'][r.status] }}</text>
      </view>
    </view>

    <!-- 操作 -->
    <view v-if="store.role >= Role.Admin" class="actions">
      <view v-if="data.user.status !== 4" class="btn danger" @click="ban">封禁该用户</view>
      <view v-else class="btn" @click="unban">解除封禁</view>
    </view>
    <view v-else class="no-perm">封禁用户需要管理员权限</view>
  </view>

  <view v-else-if="loaded" class="empty">用户不存在</view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx 64rpx;
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
  margin-bottom: 10rpx;
}

.uid {
  font-size: $font-micro;
  color: $text-placeholder;
  font-weight: normal;
  margin-left: 10rpx;
}

.sub {
  font-size: $font-mid;
  color: $text-weak;
  margin-bottom: 12rpx;
}

.status {
  display: inline-block;
  font-size: $font-micro;
  padding: 4rpx 16rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
  color: $text-weak;

  &.s1 {
    background: $warning-light;
    color: $warning;
  }

  &.s2 {
    background: $success-light;
    color: $success;
  }

  &.s3,
  &.s4 {
    background: $danger-light;
    color: $danger;
  }
}

.ban-reason {
  margin-top: 20rpx;
  padding: 18rpx;
  background: $danger-light;
  color: $danger;
  border-radius: $radius;
  font-size: $font-caption;
  line-height: 1.6;
}

.stat {
  display: flex;
}

.s {
  flex: 1;
  text-align: center;

  & + .s {
    border-left: 1rpx solid $border;
  }

  .n {
    display: block;
    font-size: $font-title;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    margin-bottom: 6rpx;
  }

  .l {
    font-size: $font-micro;
    color: $text-weak;
  }
}

.ct {
  font-size: $font-lead;
  font-weight: 600;
  margin-bottom: 18rpx;
}

.fields {
  display: flex;
  flex-wrap: wrap;
}

.f {
  width: 50%;
  display: flex;
  padding: 8rpx 0;
  font-size: $font-mid;

  .fk {
    color: $text-weak;
    width: 80rpx;
  }

  .fv {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.intro {
  margin-top: 16rpx;
  padding: 18rpx;
  background: $bg-subtle;
  border-radius: $radius;
  font-size: $font-mid;
  color: $text-sub;
  line-height: 1.7;
}

.photos {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.photo {
  width: 150rpx;
  height: 150rpx;
  border-radius: $radius-sm;
  background: $bg-subtle;
}

.report-item {
  display: flex;
  align-items: center;
  padding: 14rpx 0;
  border-bottom: 1rpx solid $border;
  font-size: $font-caption;

  &:last-child {
    border-bottom: none;
  }
}

.r-reason {
  color: $danger;
  width: 140rpx;
  flex-shrink: 0;
}

.r-desc {
  flex: 1;
  color: $text-sub;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.r-status {
  color: $text-weak;
  margin-left: 12rpx;
}

.actions {
  padding-top: 20rpx;
}

.btn {
  height: 84rpx;
  line-height: 84rpx;
  text-align: center;
  border-radius: $radius;
  background: $bg-card;
  border: 1rpx solid $border;
  color: $text-sub;
  font-size: $font-body;
  font-weight: 600;
  transition: transform 180ms cubic-bezier(0.16, 1, 0.3, 1), background-color 180ms ease;

  &:active {
    transform: scale(0.98);
  }

  &.danger {
    background: $bg-card;
    border-color: $danger;
    color: $danger;
  }
}

.no-perm {
  text-align: center;
  font-size: $font-caption;
  color: $text-placeholder;
  padding: 32rpx;
}
</style>
