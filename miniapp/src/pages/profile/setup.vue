<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import FormRow from '@/components/FormRow.vue'
import { authApi, userApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { uploadImage } from '@/utils/request'

const store = useUserStore()

/** edit=编辑已有资料，不走提交审核；否则是首次填写向导 */
const isEdit = ref(false)
const step = ref(1)
const submitting = ref(false)

const form = reactive({
  phone: '',
  nickname: '',
  avatar: '',
  gender: 0,
  birthday: '',
  height: 0,
  weight: 0,
  marital_status: 0,
  has_children: 0,
  education: 0,
  school: '',
  occupation: '',
  income_level: 0,
  house_status: 0,
  car_status: 0,
  province: '',
  city: '',
  city_code: '',
  home_province: '',
  home_city: '',
  self_intro: '',
})

const req = reactive({
  age_min: 22,
  age_max: 40,
  height_min: 150,
  height_max: 195,
  education_min: 0,
  income_min: 0,
  marital_status: '1',
  accept_children: 1,
  city_codes: '',
  other: '',
})

/** 择偶婚史多选的本地状态 */
const acceptMarital = reactive<Record<number, boolean>>({ 1: true, 2: false, 3: false })

const heightRange = Array.from({ length: 81 }, (_, i) => 140 + i)
const ageRange = Array.from({ length: 63 }, (_, i) => 18 + i)

/** 生日上限：必须年满 18 周岁 */
const maxBirthday = computed(() => {
  const d = new Date()
  d.setFullYear(d.getFullYear() - 18)
  return d.toISOString().slice(0, 10)
})

onLoad(async (query) => {
  isEdit.value = query?.mode === 'edit'
  await store.fetchDict()

  if (!store.user) await store.fetchMe()

  // 已有数据回填
  const u = store.user
  if (u) {
    form.nickname = u.nickname
    form.avatar = u.avatar
    form.gender = u.gender
  }
  form.phone = store.phone

  const p = store.profile
  if (p) {
    Object.assign(form, {
      birthday: p.birthday ? String(p.birthday).slice(0, 10) : '',
      height: p.height,
      weight: p.weight,
      marital_status: p.marital_status,
      has_children: p.has_children,
      education: p.education,
      school: p.school,
      occupation: p.occupation,
      income_level: p.income_level,
      house_status: p.house_status,
      car_status: p.car_status,
      province: p.province,
      city: p.city,
      city_code: p.city_code,
      home_province: p.home_province,
      home_city: p.home_city,
      self_intro: p.self_intro,
    })
  }

  const r = store.requirement
  if (r) {
    Object.assign(req, r)
    Object.keys(acceptMarital).forEach((k) => (acceptMarital[+k] = false))
    ;(r.marital_status || '').split(',').forEach((v) => {
      if (v) acceptMarital[+v] = true
    })
  }

  if (isEdit.value) step.value = 1

  // 首次注册且服务端没有资料时，尝试恢复本地草稿。
  // 4 步表单填一半退出全丢，等于直接流失一个注册
  if (!isEdit.value && !store.profile) {
    restoreDraft()
  }
})

// ---------- 草稿 ----------
const DRAFT_KEY = 'profile_draft'

function saveDraft() {
  if (isEdit.value) return // 编辑模式以服务端数据为准，不存草稿
  uni.setStorageSync(DRAFT_KEY, {
    form: { ...form },
    req: { ...req },
    acceptMarital: { ...acceptMarital },
    step: step.value,
  })
}

function restoreDraft() {
  const draft = uni.getStorageSync(DRAFT_KEY)
  if (!draft || !draft.form) return
  // 手机号可能已经绑定过，服务端的值优先
  const phone = form.phone
  Object.assign(form, draft.form)
  if (phone) form.phone = phone
  Object.assign(req, draft.req || {})
  Object.assign(acceptMarital, draft.acceptMarital || {})
  if (draft.step > 1) {
    step.value = draft.step
    uni.showToast({ title: '已恢复上次填写的内容', icon: 'none' })
  }
}

function clearDraft() {
  uni.removeStorageSync(DRAFT_KEY)
}

// 中途退出（返回、杀进程前的页面卸载）时兜底保存
onUnload(saveDraft)

// ---------- 选择器 ----------
function pickDict(type: string, field: keyof typeof form) {
  const list = store.dict[type] || []
  uni.showActionSheet({
    itemList: list.map((i) => i.label),
    success: (res) => {
      // @ts-ignore 动态字段赋值
      form[field] = list[res.tapIndex].value
    },
  })
}

function pickReqDict(type: string, field: 'education_min' | 'income_min') {
  const list = [{ value: 0, label: '不限' }, ...(store.dict[type] || [])]
  uni.showActionSheet({
    itemList: list.map((i) => i.label),
    success: (res) => (req[field] = list[res.tapIndex].value),
  })
}

function onRegionChange(e: any, target: 'live' | 'home') {
  const [prov, city] = e.detail.value
  const codes = e.detail.code || []
  if (target === 'live') {
    form.province = prov
    form.city = city
    form.city_code = codes[1] || ''
  } else {
    form.home_province = prov
    form.home_city = city
  }
}

async function chooseAvatar() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    success: async (res) => {
      const path = (res.tempFilePaths as string[])[0]
      uni.showLoading({ title: '上传中' })
      try {
        form.avatar = await uploadImage(path, 'avatar')
      } finally {
        uni.hideLoading()
      }
    },
  })
}

// ---------- 分步校验 ----------
function validateStep(n: number): string {
  if (n === 1) {
    if (!form.phone) return '请填写手机号'
    if (!/^1[3-9]\d{9}$/.test(form.phone)) return '手机号格式不正确'
    if (!form.avatar) return '请上传头像'
    if (form.nickname.length < 2) return '昵称至少 2 个字'
    if (!form.gender) return '请选择性别'
    if (!form.birthday) return '请选择出生日期'
  }
  if (n === 2) {
    if (!form.height) return '请选择身高'
    if (!form.marital_status) return '请选择婚姻状况'
    if (!form.province) return '请选择现居城市'
    if (!form.home_province) return '请选择户籍地'
  }
  if (n === 3) {
    if (!form.education) return '请选择学历'
    if (!form.occupation) return '请填写职业'
    if (!form.income_level) return '请选择月收入'
    if (!form.house_status) return '请选择购房情况'
    if (!form.car_status) return '请选择购车情况'
  }
  if (n === 4) {
    if (form.self_intro.trim().length < 30) return '自我介绍至少 30 字'
  }
  return ''
}

async function next() {
  const err = validateStep(step.value)
  if (err) {
    uni.showToast({ title: err, icon: 'none' })
    return
  }

  // 手机号在第一步就绑定，后面步骤失败也不用重填
  if (step.value === 1 && form.phone !== store.phone) {
    try {
      await authApi.bindPhone({ phone: form.phone })
      store.phone = form.phone
    } catch {
      return
    }
  }

  if (step.value < 4) {
    step.value++
    saveDraft() // 每完成一步存一次，中途退出不丢
    uni.pageScrollTo({ scrollTop: 0, duration: 0 })
    return
  }
  await submit()
}

async function submit() {
  if (submitting.value) return
  submitting.value = true
  uni.showLoading({ title: '提交中' })

  try {
    req.marital_status = Object.keys(acceptMarital)
      .filter((k) => acceptMarital[+k])
      .join(',')

    await userApi.saveProfile(form)
    await userApi.saveRequirement(req)

    if (!isEdit.value) {
      await userApi.submitAudit()
    }

    await store.fetchMe()
    clearDraft()
    uni.hideLoading()

    uni.showModal({
      title: isEdit.value ? '已保存' : '提交成功',
      content: isEdit.value
        ? '资料修改后需要重新审核，期间不影响正常使用'
        : '资料已提交审核，通常 24 小时内完成',
      showCancel: false,
      success: () => uni.reLaunch({ url: '/pages/index/index' }),
    })
  } catch {
    uni.hideLoading()
  } finally {
    submitting.value = false
  }
}

function prev() {
  if (step.value > 1) step.value--
}
</script>

<template>
  <view class="setup">
    <!-- 进度 -->
    <view class="steps">
      <view v-for="n in 4" :key="n" class="step-item" :class="{ active: n <= step }">
        <view class="dot">{{ n }}</view>
        <view class="line" v-if="n < 4"></view>
      </view>
    </view>
    <view class="step-title">
      {{ ['基础信息', '个人情况', '硬性条件', '介绍与期望'][step - 1] }}
      <text class="step-sub">{{ step }}/4</text>
    </view>

    <!-- 第 1 步 -->
    <view v-show="step === 1" class="card">
      <FormRow label="手机号" required>
        <input
          v-model="form.phone"
          type="number"
          maxlength="11"
          placeholder="用于账号安全，不会公开"
          placeholder-class="ph"
        />
      </FormRow>

      <FormRow label="头像" required align-right>
        <view class="avatar-row" @click="chooseAvatar">
          <image v-if="form.avatar" :src="form.avatar" class="avatar" mode="aspectFill" />
          <view v-else class="avatar placeholder">＋</view>
        </view>
      </FormRow>

      <FormRow label="昵称" required>
        <input v-model="form.nickname" maxlength="12" placeholder="2-12 个字" placeholder-class="ph" />
      </FormRow>

      <FormRow label="性别" required align-right>
        <view v-if="store.user?.gender" class="locked">
          {{ store.label('gender', form.gender) }}
          <text class="lock-tip">（不可修改）</text>
        </view>
        <view v-else class="gender-group">
          <view
            v-for="g in [
              { v: 1, t: '男' },
              { v: 2, t: '女' },
            ]"
            :key="g.v"
            class="gender-item"
            :class="{ on: form.gender === g.v }"
            @click="form.gender = g.v"
          >
            {{ g.t }}
          </view>
        </view>
      </FormRow>

      <FormRow label="出生日期" required align-right>
        <picker mode="date" :end="maxBirthday" :value="form.birthday" @change="form.birthday = $event.detail.value">
          <text :class="{ ph: !form.birthday }">{{ form.birthday || '请选择' }}</text>
        </picker>
      </FormRow>
    </view>

    <!-- 第 2 步 -->
    <view v-show="step === 2" class="card">
      <FormRow label="身高" required align-right>
        <picker
          mode="selector"
          :range="heightRange"
          @change="form.height = heightRange[$event.detail.value]"
        >
          <text :class="{ ph: !form.height }">{{ form.height ? form.height + ' cm' : '请选择' }}</text>
        </picker>
      </FormRow>

      <FormRow label="体重" align-right>
        <input
          v-model.number="form.weight"
          type="number"
          class="ta-right"
          placeholder="选填"
          placeholder-class="ph"
        />
      </FormRow>

      <FormRow label="婚姻状况" required align-right>
        <text
          :class="{ ph: !form.marital_status }"
          @click="pickDict('marital_status', 'marital_status')"
        >
          {{ store.label('marital_status', form.marital_status) || '请选择' }}
        </text>
      </FormRow>

      <FormRow label="子女情况" required align-right>
        <text @click="pickDict('has_children', 'has_children')">
          {{ store.label('has_children', form.has_children) || '无' }}
        </text>
      </FormRow>

      <FormRow label="现居城市" required align-right>
        <picker mode="region" @change="onRegionChange($event, 'live')">
          <text :class="{ ph: !form.province }">
            {{ form.province ? form.province + ' ' + form.city : '请选择' }}
          </text>
        </picker>
      </FormRow>

      <FormRow label="户籍地" required align-right>
        <picker mode="region" @change="onRegionChange($event, 'home')">
          <text :class="{ ph: !form.home_province }">
            {{ form.home_province ? form.home_province + ' ' + form.home_city : '请选择' }}
          </text>
        </picker>
      </FormRow>
    </view>

    <!-- 第 3 步 -->
    <view v-show="step === 3" class="card">
      <FormRow label="学历" required align-right>
        <text :class="{ ph: !form.education }" @click="pickDict('education', 'education')">
          {{ store.label('education', form.education) || '请选择' }}
        </text>
      </FormRow>

      <FormRow label="毕业院校" align-right>
        <input v-model="form.school" class="ta-right" placeholder="选填" placeholder-class="ph" />
      </FormRow>

      <FormRow label="职业" required align-right>
        <input
          v-model="form.occupation"
          class="ta-right"
          maxlength="20"
          placeholder="如 软件工程师"
          placeholder-class="ph"
        />
      </FormRow>

      <FormRow label="月收入" required align-right>
        <text :class="{ ph: !form.income_level }" @click="pickDict('income_level', 'income_level')">
          {{ store.label('income_level', form.income_level) || '请选择' }}
        </text>
      </FormRow>

      <FormRow label="购房情况" required align-right>
        <text :class="{ ph: !form.house_status }" @click="pickDict('house_status', 'house_status')">
          {{ store.label('house_status', form.house_status) || '请选择' }}
        </text>
      </FormRow>

      <FormRow label="购车情况" required align-right>
        <text :class="{ ph: !form.car_status }" @click="pickDict('car_status', 'car_status')">
          {{ store.label('car_status', form.car_status) || '请选择' }}
        </text>
      </FormRow>
    </view>

    <!-- 第 4 步 -->
    <view v-show="step === 4">
      <view class="card">
        <view class="section-title">自我介绍</view>
        <textarea
          v-model="form.self_intro"
          class="intro"
          maxlength="500"
          placeholder="聊聊你的性格、生活状态和对另一半的期待。写得越具体，越容易遇到合适的人。"
          placeholder-class="ph"
        />
        <view class="count" :class="{ warn: form.self_intro.length < 30 }">
          {{ form.self_intro.length }}/500（至少 30 字）
        </view>
        <view class="tip">请勿填写微信号、手机号等联系方式，配对成功后可在站内聊天</view>
      </view>

      <view class="card">
        <view class="section-title">择偶要求</view>

        <FormRow label="年龄范围" align-right>
          <picker mode="selector" :range="ageRange" @change="req.age_min = ageRange[$event.detail.value]">
            <text class="pick-inline">{{ req.age_min }}</text>
          </picker>
          <text class="dash">-</text>
          <picker mode="selector" :range="ageRange" @change="req.age_max = ageRange[$event.detail.value]">
            <text class="pick-inline">{{ req.age_max }}</text>
          </picker>
          <text class="unit">岁</text>
        </FormRow>

        <FormRow label="身高范围" align-right>
          <picker mode="selector" :range="heightRange" @change="req.height_min = heightRange[$event.detail.value]">
            <text class="pick-inline">{{ req.height_min }}</text>
          </picker>
          <text class="dash">-</text>
          <picker mode="selector" :range="heightRange" @change="req.height_max = heightRange[$event.detail.value]">
            <text class="pick-inline">{{ req.height_max }}</text>
          </picker>
          <text class="unit">cm</text>
        </FormRow>

        <FormRow label="学历要求" align-right>
          <text @click="pickReqDict('education', 'education_min')">
            {{ req.education_min ? store.label('education', req.education_min) + '及以上' : '不限' }}
          </text>
        </FormRow>

        <FormRow label="收入要求" align-right>
          <text @click="pickReqDict('income_level', 'income_min')">
            {{ req.income_min ? store.label('income_level', req.income_min) + '及以上' : '不限' }}
          </text>
        </FormRow>

        <FormRow label="可接受婚史" align-right>
          <view class="multi">
            <view
              v-for="m in store.dict['marital_status'] || []"
              :key="m.value"
              class="multi-item"
              :class="{ on: acceptMarital[m.value] }"
              @click="acceptMarital[m.value] = !acceptMarital[m.value]"
            >
              {{ m.label }}
            </view>
          </view>
        </FormRow>

        <FormRow label="接受有子女" align-right>
          <switch
            :checked="req.accept_children === 1"
            color="#e8506d"
            @change="req.accept_children = $event.detail.value ? 1 : 0"
          />
        </FormRow>
      </view>
    </view>

    <!-- 底部操作 -->
    <view class="actions">
      <view v-if="step > 1" class="btn-plain half" @click="prev">上一步</view>
      <view class="btn-primary" :class="{ half: step > 1, disabled: submitting }" @click="next">
        {{ step < 4 ? '下一步' : isEdit ? '保存' : '提交审核' }}
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.setup {
  padding: 24rpx 24rpx 180rpx;
}

.steps {
  display: flex;
  align-items: center;
  padding: 20rpx 40rpx 0;
}

.step-item {
  display: flex;
  align-items: center;
  flex: 1;

  &:last-child {
    flex: 0;
  }

  .dot {
    width: 44rpx;
    height: 44rpx;
    line-height: 44rpx;
    text-align: center;
    border-radius: 50%;
    background: #ebedf0;
    color: #8a8f99;
    font-size: 24rpx;
    flex-shrink: 0;
  }

  .line {
    flex: 1;
    height: 4rpx;
    background: #ebedf0;
  }

  &.active .dot {
    background: #e8506d;
    color: #fff;
  }

  &.active .line {
    background: #f5b3c1;
  }
}

.step-title {
  font-size: 34rpx;
  font-weight: 600;
  padding: 32rpx 8rpx 24rpx;

  .step-sub {
    font-size: 24rpx;
    color: #8a8f99;
    font-weight: normal;
    margin-left: 12rpx;
  }
}

.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 8rpx 24rpx;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  padding: 24rpx 0 8rpx;
}

.ph {
  color: #bbbfc4;
}

.ta-right {
  text-align: right;
}

.avatar-row {
  display: flex;
  justify-content: flex-end;
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: 16rpx;
  background: #f5f6f8;

  &.placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #bbbfc4;
    font-size: 48rpx;
    border: 1rpx dashed #d9dce0;
  }
}

.gender-group {
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
}

.gender-item {
  width: 120rpx;
  height: 64rpx;
  line-height: 64rpx;
  text-align: center;
  border-radius: 32rpx;
  background: #f5f6f8;
  color: #646a73;
  font-size: 26rpx;

  &.on {
    background: #e8506d;
    color: #fff;
  }
}

.locked {
  color: #8a8f99;

  .lock-tip {
    font-size: 22rpx;
    color: #bbbfc4;
  }
}

.intro {
  width: 100%;
  height: 240rpx;
  font-size: 28rpx;
  line-height: 1.6;
  padding: 16rpx 0;
}

.count {
  text-align: right;
  font-size: 22rpx;
  color: #8a8f99;

  &.warn {
    color: #ff8f1f;
  }
}

.tip {
  font-size: 22rpx;
  color: #bbbfc4;
  padding: 12rpx 0 20rpx;
  line-height: 1.5;
}

.pick-inline {
  display: inline-block;
  min-width: 70rpx;
  text-align: center;
  color: #e8506d;
}

.dash,
.unit {
  color: #8a8f99;
  margin: 0 6rpx;
}

.multi {
  display: flex;
  justify-content: flex-end;
  gap: 12rpx;
}

.multi-item {
  padding: 8rpx 20rpx;
  border-radius: 24rpx;
  background: #f5f6f8;
  color: #646a73;
  font-size: 24rpx;

  &.on {
    background: #fff0f3;
    color: #e8506d;
  }
}

.actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  display: flex;
  gap: 20rpx;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.04);
}

.half {
  flex: 1;
}

.actions .btn-primary {
  flex: 2;
}
</style>
