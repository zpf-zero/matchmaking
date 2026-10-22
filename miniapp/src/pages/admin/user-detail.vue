<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import AdminBack from '@/components/AdminBack.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { adminApi } from '@/api'
import { Role, Status, useUserStore } from '@/stores/user'

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

/** 只有超管能任命角色；且不能操作其他超管与自己 */
const canSetRole = computed(
  () =>
    store.role >= Role.SuperAdmin &&
    data.value?.user?.role !== Role.SuperAdmin &&
    data.value?.user?.id !== store.user?.id
)

/**
 * 是否可审核这份资料。
 *
 * 从「用户管理」点进来的待审用户，原先只有「变更角色 / 封禁」两个按钮 ——
 * 管理员看到有人卡在待审核，却没有任何办法让他通过，只能绕回「资料审核」页
 * 去列表里再找一遍。这里把审核入口补到详情页。
 *
 * 后端 AuditUser 的约束：只审 role=1 的普通用户（运营号不进队列），
 * 且 status 不能是「待完善」（还没提交）或「已封禁」（要先解封）。
 * 这里保持一致，避免点了才报错。
 *
 * 注意 status=3（未通过）也要给入口 —— 拒错了、或用户改完资料想再争取，
 * 都靠这一步纠正。少了它，误拒的用户在详情页就再也翻不了身。
 */
const canAudit = computed(() => {
  const u = data.value?.user
  if (!u) return false
  return (
    store.role >= Role.Auditor &&
    u.role === Role.User &&
    // 待审核 / 正常 / 未通过 均可复审（未通过的能改判，不能永久卡死）
    [Status.Auditing, Status.Normal, Status.Rejected].includes(u.status)
  )
})

const roleNames = ['', '普通用户', '审核员', '管理员', '超级管理员']

// 0 待审核 / 1 已通过 / 2 未通过
const auditTag = computed(() => {
  const s = data.value?.profile?.audit_status
  if (s === 1) return { text: '已通过', cls: 'passed' }
  if (s === 2) return { text: '未通过', cls: 'rejected' }
  return { text: '待审核', cls: 'pending' }
})

function auditPass() {
  uni.showModal({
    title: '通过审核',
    content: '通过后该用户可正常浏览与互动。',
    success: async (r) => {
      if (!r.confirm) return
      try {
        await adminApi.audit(uid.value, true)
        uni.showToast({ title: '已通过', icon: 'none' })
        load()
      } catch {
        // 已提示
      }
    },
  })
}

const REJECT_PRESETS = [
  '头像非真人照片',
  '资料信息不真实',
  '自我介绍含联系方式',
  '照片含违规内容',
  '自定义原因',
]

function auditReject() {
  uni.showActionSheet({
    itemList: REJECT_PRESETS,
    success: (res) => {
      if (res.tapIndex === REJECT_PRESETS.length - 1) {
        customReject()
      } else {
        doReject(REJECT_PRESETS[res.tapIndex])
      }
    },
  })
}

function customReject() {
  uni.showModal({
    title: '拒绝原因',
    editable: true,
    placeholderText: '请填写具体原因，用户会看到',
    success: (r) => {
      if (r.confirm && r.content?.trim()) doReject(r.content.trim())
    },
  })
}

async function doReject(remark: string) {
  try {
    await adminApi.audit(uid.value, false, remark)
    uni.showToast({ title: '已拒绝', icon: 'none' })
    load()
  } catch {
    // 已提示
  }
}

function setRole() {
  const cur = Number(data.value?.user?.role || 1)
  const options: { label: string; role: number }[] = [
    { label: '设为审核员', role: Role.Auditor },
    { label: '设为管理员', role: Role.Admin },
    { label: '降为普通用户', role: Role.User },
  ].filter((o) => o.role !== cur) as { label: string; role: number }[]

  uni.showActionSheet({
    itemList: options.map((o) => o.label),
    success: (res) => {
      const picked = options[res.tapIndex]
      if (!picked) return
      confirmSetRole(picked.role, picked.label)
    },
  })
}

function confirmSetRole(role: number, label: string) {
  const extra =
    role > Role.User
      ? '\n该账号将免于资料审核、不参与择偶，并可直接进入管理后台。'
      : '\n撤销后该账号按普通用户规则使用（资料需经审核）。'
  uni.showModal({
    title: '确认变更角色',
    content: `确定「${label}」吗？${extra}`,
    success: async (r) => {
      if (!r.confirm) return
      try {
        await adminApi.setRole(uid.value, role)
        uni.showToast({ title: '已变更', icon: 'none' })
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
    <AdminBack label="用户管理" to="/pages/admin/user" />
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
          <view class="chips">
            <view class="status" :class="'s' + data.user.status">
              {{ statusText[data.user.status] }}
            </view>
            <view v-if="data.user.role > Role.User" class="role-chip">{{ roleNames[data.user.role] }}</view>
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

    <!-- 资料审核：待审 / 正常 / 未通过 均可复审，不必绕回列表页 -->
    <view v-if="canAudit" class="card audit">
      <view class="ct">
        资料审核
        <text class="atag" :class="auditTag.cls">{{ auditTag.text }}</text>
      </view>
      <view v-if="data.profile?.audit_remark" class="audit-remark">
        上次拒绝原因：{{ data.profile.audit_remark }}
      </view>
      <view class="actions">
        <view v-if="auditTag.cls !== 'rejected'" class="btn" @click="auditReject">拒绝</view>
        <view v-if="auditTag.cls !== 'passed'" class="btn pass" @click="auditPass">通过</view>
        <view v-else class="btn" @click="auditReject">改判为未通过</view>
      </view>
    </view>

    <!-- 操作 -->
    <view class="actions">
      <view v-if="canSetRole" class="btn" @click="setRole">变更角色</view>
      <template v-if="store.role >= Role.Admin">
        <view v-if="data.user.status !== 4" class="btn danger" @click="ban">封禁该用户</view>
        <view v-else class="btn" @click="unban">解除封禁</view>
      </template>
      <view v-else-if="!canSetRole && !canAudit" class="no-perm">封禁用户需要管理员权限</view>
    </view>
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

.chips {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

/* 运营角色标记：和状态标签并列，用品牌色区分 */
.role-chip {
  display: inline-block;
  font-size: $font-micro;
  padding: 4rpx 16rpx;
  border-radius: $radius-sm;
  background: $brand-light;
  color: $brand;
  font-weight: 600;
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

  /* 审核区的按钮和其他操作按钮同款，但要有主次之分 */
  .btn.pass {
    background: $brand;
    border-color: $brand;
    color: $text-on-accent;
  }
}

/* 审核卡片：和「资料信息」卡区分开，用品牌色描边突出是待办 */
.audit {
  border-color: $brand-light;
  background: $bg-accent;
}

.atag {
  font-size: $font-micro;
  padding: 4rpx 14rpx;
  border-radius: $radius-sm;
  margin-left: 12rpx;
  background: $warning-light;
  color: $warning;
  font-weight: normal;

  &.passed {
    background: $success-light;
    color: $success;
  }

  &.rejected {
    background: $danger-light;
    color: $danger;
  }
}

.audit-remark {
  font-size: $font-caption;
  color: $danger;
  line-height: 1.6;
  padding-bottom: 8rpx;
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
