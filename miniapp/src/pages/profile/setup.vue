<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import FormRow from '@/components/FormRow.vue'
import RegionPicker from '@/components/RegionPicker.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { authApi, inviteApi, userApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { uploadImage } from '@/utils/request'

const store = useUserStore()

/*
 * 微信官方授权能力只在微信小程序端存在：
 *   - getPhoneNumber：一键获取手机号（企业主体 + 微信认证，0.03 元/次）
 *   - chooseAvatar：获取用户头像（getUserProfile 已于基础库 2.27.1 起回收该能力）
 * H5 端没有这些组件，只能走手填 + chooseImage 兜底。
 *
 * ============================ 本页涉及的待申请事项 ============================
 * 【待申请】手机号快速验证组件 —— docs/07-待申请事项清单.md 第 5 项
 *   0.03 元/次，每小程序 1000 次免费额度。无需单独申请，但要在微信后台确认额度。
 *
 * 【待申请】隐私保护指引 —— 清单第 6 项（最容易漏，漏了一定卡真机）
 *   必须在微信后台「设置 → 服务内容声明 → 用户隐私保护指引」声明收集：
 *   手机号、相册/摄像头、昵称头像。未配置时真实模式下这几个接口会直接调用失败，
 *   而开发者工具里可能看不出问题 —— 真机上线才暴露。
 *
 * 【待申请】AppID / AppSecret —— 清单第 4 项
 *   当前 store.mockMode=true，canWxAuth 为 false，因此显示手填输入框而非授权按钮。
 *   关闭 mock 后本页会自动切到官方授权组件，无需改代码。
 * ============================================================================
 */
// #ifdef MP-WEIXIN
const IS_MP_WEIXIN = true
// #endif
// #ifndef MP-WEIXIN
const IS_MP_WEIXIN = false
// #endif

/** 走微信官方授权组件的条件：小程序端 + 非 mock（mock 模式手填，方便开发者工具调试） */
const canWxAuth = computed(() => IS_MP_WEIXIN && !store.mockMode)

/**
 * 与 App.vue 共用的「待绑定邀请码」storage key。
 * App.vue 在启动/被唤起时捕获分享或扫码带来的码存进去，本页在向导里预填。
 */
const PENDING_INVITE_KEY = 'pending_invite_code'

/** edit=编辑已有资料，不走提交审核；否则是首次填写向导 */
const isEdit = ref(false)
const step = ref(1)
const submitting = ref(false)

// ---------- 步骤切换动画 ----------
// 4 步向导是一整条注册转化链路，硬切会让用户失去方向感。
// 做两段式：旧内容先滑出（out），再换内容滑入（in），方向由前进/后退决定。
const TRANS_OUT = 170 // 滑出时长，要略短于滑入，避免节奏拖沓
const TRANS_IN = 300
const dir = ref<'next' | 'prev'>('next')
const phase = ref<'idle' | 'out' | 'in'>('idle')
let phaseTimer: ReturnType<typeof setTimeout> | undefined

/** 提交成功覆盖层 */
const doneVisible = ref(false)
let doneTimer: ReturnType<typeof setTimeout> | undefined

/** 舞台容器的动画 class。phase 参与类名，保证两次切换之间类名会变、动画能重新触发 */
const stageClass = computed(() => (phase.value === 'idle' ? '' : `s-${phase.value}-${dir.value}`))

/** 带动画地切到第 n 步 */
function goStep(n: number) {
  if (n === step.value || phase.value !== 'idle') return // 动画期间忽略连点，避免状态错乱
  dir.value = n > step.value ? 'next' : 'prev'
  phase.value = 'out'
  clearTimeout(phaseTimer)
  phaseTimer = setTimeout(() => {
    step.value = n
    phase.value = 'in'
    uni.pageScrollTo({ scrollTop: 0, duration: 0 })
    phaseTimer = setTimeout(() => (phase.value = 'idle'), TRANS_IN)
  }, TRANS_OUT)
}

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
  /** 选填邀请码。不参与资料提交，单独调 /invite/bind */
  invite_code: '',
  /** 行业标签（粗粒度），选填，免费内容 */
  industry: 0,

  // ---- 原生家庭（选填，默认私有）----
  // 必须用户主动勾选「对已解锁用户可见」才会展示。这是最敏感的资料块，
  // 默认公开会劝退大量用户（见 docs/08 6-2）
  family_parent_occ: 0,
  family_atmosphere: 0,
  family_only_child: 0,
  family_city: '',
  family_note: '',
  family_visible: 0,
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
      industry: p.industry || 0,
      family_parent_occ: p.family_parent_occ || 0,
      family_atmosphere: p.family_atmosphere || 0,
      family_only_child: p.family_only_child || 0,
      family_city: p.family_city || '',
      family_note: p.family_note || '',
      family_visible: p.family_visible || 0,
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

  // 预填分享带进来的邀请码。
  //
  // 这条链路此前是断的：App.vue 的 captureInvite 把码存进 storage，
  // 但只有「重启小程序且 status≠0」时才会被 flushPendingInvite 补交，
  // 新用户走完向导提交时 form.invite_code 是空的 —— 码要等下次启动才生效。
  // 这里在向导里直接读出来预填，用户提交时就随主流程一起绑定了。
  //
  // 只读不删：用户可能中途退出，删了就丢了。真正清除在提交成功后。
  if (!isEdit.value && !form.invite_code) {
    const pending = uni.getStorageSync(PENDING_INVITE_KEY) || ''
    if (pending) form.invite_code = String(pending).toUpperCase()
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
  }}

function clearDraft() {
  uni.removeStorageSync(DRAFT_KEY)
}

/**
 * 退出登录 / 切换账号。
 *
 * 首次填写向导是登录后 reLaunch 进来的 —— 页面栈是空的，系统返回键直接退出整个
 * H5，用户被永久困在这一页。mock 模式下把开发账号名输错（比如把 mock_admin 敲成
 * test_admin）就会当场注册出一个普通用户，然后卡死在这里，无处可退。
 * 编辑模式（从「我的」点进来）有正常的返回，不需要这个出口。
 */
function switchAccount() {
  uni.showModal({
    title: '切换账号',
    content: '将退出当前账号，已填写的内容会保留，下次登录可继续填写。',
    confirmText: '退出登录',
    success: async (r) => {
      if (!r.confirm) return
      saveDraft() // 草稿先落地，退出不丢
      try {
        await authApi.logout()
      } catch {
        // 退出接口失败不影响本地清理
      }
      store.logout() // 内部会 reLaunch 到登录页
    },
  })
}

// 中途退出（返回、杀进程前的页面卸载）时兜底保存
onUnload(() => {
  saveDraft()
  clearTimeout(phaseTimer)
  clearTimeout(doneTimer)
})

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

function onRegionChange(v: { province: string; city: string; code: string }, target: 'live' | 'home') {
  if (target === 'live') {
    form.province = v.province
    form.city = v.city
    form.city_code = v.code
  } else {
    form.home_province = v.province
    form.home_city = v.city
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

/**
 * 微信官方头像选择。
 *
 * wx.getUserProfile 自基础库 2.27.1 起已回收头像昵称（统一返回灰色头像 +
 * 「微信用户」），官方现行为「用户自主填写」：头像走 open-type="chooseAvatar"，
 * 昵称走 <input type="nickname">。
 * chooseAvatar 给到的是**临时文件路径**，必须上传到自己的服务器持久化。
 */
async function onChooseAvatar(e: any) {
  const path = e?.detail?.avatarUrl
  if (!path) return
  uni.showLoading({ title: '上传中' })
  try {
    form.avatar = await uploadImage(path, 'avatar')
  } catch {
    // uploadImage 内部已提示
  } finally {
    uni.hideLoading()
  }
}

/**
 * 微信一键获取手机号。
 *
 * e.detail.code 是一次性凭证，交给后端调 getuserphonenumber 换取真实号码
 * （小程序端拿不到明文手机号，只有后端解密才有）。
 * 未认证主体调用会直接失败，此时提示走手填兜底。
 */
async function onGetPhoneNumber(e: any) {
  const code = e?.detail?.code
  if (!code) {
    // 用户点了「拒绝」，或主体未完成微信认证导致组件不可用
    uni.showToast({ title: '未获取到手机号，可手动填写', icon: 'none' })
    return
  }
  uni.showLoading({ title: '获取中' })
  try {
    await authApi.bindPhone({ code })
    await store.fetchMe()
    form.phone = store.phone
    uni.showToast({ title: '已绑定微信手机号', icon: 'none' })
  } catch {
    // request 层已提示
  } finally {
    uni.hideLoading()
  }
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
    goStep(step.value + 1)
    saveDraft() // 每完成一步存一次，中途退出不丢
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

    // 邀请码绑定要在提交审核之前：审核通过时由后端结算双方权益。
    // 绑定失败不阻断主流程 —— 邀请码填错不该让用户提交不了资料。
    if (!isEdit.value && form.invite_code.trim()) {
      try {
        await inviteApi.bind(form.invite_code.trim().toUpperCase())
      } catch {
        uni.showToast({ title: '邀请码无效，已跳过', icon: 'none' })
      }
    }

    if (!isEdit.value) {
      await userApi.submitAudit()
      // 邀请码已随本次提交绑定，清掉本地暂存，避免 App.vue 下次启动再补交一遍
      uni.removeStorageSync(PENDING_INVITE_KEY)
    }

    await store.fetchMe()
    clearDraft()
    uni.hideLoading()

    // 提交成功用一个轻量的成功态覆盖层收尾，替代原生灰弹窗：
    // 这是整条注册链路的终点，值得给一个明确的「完成了」信号
    doneVisible.value = true
    doneTimer = setTimeout(() => uni.reLaunch({ url: '/pages/index/index' }), 1800)
  } catch {
    uni.hideLoading()
  } finally {
    submitting.value = false
  }
}

function prev() {
  if (step.value > 1) goStep(step.value - 1)
}

/** 点击顶部进度点跳转。往回退自由放行；往前走必须把中间步骤挨个校验通过，
 *  否则用户可以直接跳到第 4 步提交一份缺项的资料 */
async function jump(n: number) {
  if (n === step.value) return
  if (n < step.value) {
    goStep(n)
    return
  }
  for (let i = step.value; i < n; i++) {
    const err = validateStep(i)
    if (err) {
      uni.showToast({ title: err, icon: 'none' })
      return
    }
  }
  // 跳过第 1 步时，手机号绑定这件事不能漏（原本挂在 next() 里）
  if (step.value === 1 && form.phone !== store.phone) {
    try {
      await authApi.bindPhone({ phone: form.phone })
      store.phone = form.phone
    } catch {
      return
    }
  }
  goStep(n)
}
</script>

<template>
  <view class="setup">
    <!-- 首次填写向导是 reLaunch 进来的、页面栈为空，上一页不存在也没有返回键。
         登录错账号时必须有路可退，否则只能清缓存重来。编辑模式有正常返回，不显示。 -->
    <view v-if="!isEdit" class="topbar">
      <view class="topbar-exit" @click="switchAccount">切换账号</view>
    </view>

    <!-- 进度 -->
    <view class="steps">
      <view
        v-for="n in 4"
        :key="n"
        class="step-item"
        :class="{ active: n <= step, done: n < step }"
        @click="jump(n)"
      >
        <view class="dot">
          <text v-if="n < step" class="tick">✓</text>
          <text v-else>{{ n }}</text>
        </view>
        <view class="line" v-if="n < 4"></view>
      </view>
    </view>

    <!-- 步骤内容舞台：切步时整块滑出/滑入，给用户方向感 -->
    <view class="stage" :class="stageClass">
      <view class="step-title">
        {{ ['基础信息', '个人情况', '硬性条件', '介绍与期望'][step - 1] }}
        <text class="step-sub">{{ step }}/4</text>
      </view>

      <!-- 第 1 步 -->
      <view v-show="step === 1" class="card">
      <FormRow label="手机号" required>
        <!-- 真实模式走微信一键授权；mock/H5 手填。两条路径都落同一个 /auth/bind-phone -->
        <button
          v-if="canWxAuth && !form.phone"
          class="wx-phone-btn"
          open-type="getPhoneNumber"
          @getphonenumber="onGetPhoneNumber"
        >
          微信一键获取
        </button>
        <input
          v-else
          v-model="form.phone"
          type="number"
          maxlength="11"
          :disabled="canWxAuth && !!form.phone"
          placeholder="用于账号安全，不会公开"
          placeholder-class="ph"
        />
      </FormRow>

      <FormRow label="头像" required align-right>
        <!-- 小程序端用官方 chooseAvatar，H5 端退回相册选择 -->
        <button
          v-if="canWxAuth"
          class="avatar-btn"
          open-type="chooseAvatar"
          @chooseavatar="onChooseAvatar"
        >
          <UserAvatar v-if="form.avatar" :src="form.avatar" :name="form.nickname" :size="120" />
          <view v-else class="placeholder">＋</view>
        </button>
        <view v-else class="avatar-row" @click="chooseAvatar">
          <UserAvatar v-if="form.avatar" :src="form.avatar" :name="form.nickname" :size="120" />
          <view v-else class="placeholder">＋</view>
        </view>
      </FormRow>

      <FormRow label="昵称" required>
        <!-- type="nickname" 会让键盘顶部出现「使用微信昵称」的快捷填入 -->
        <input
          v-model="form.nickname"
          :type="canWxAuth ? 'nickname' : 'text'"
          maxlength="12"
          placeholder="2-12 个字"
          placeholder-class="ph"
        />
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

      <!-- 邀请码：选填，不影响任何功能使用。权益由系统在审核通过后自动发放 -->
      <FormRow v-if="!isEdit" label="邀请码" align-right>
        <input
          v-model="form.invite_code"
          class="ta-right"
          maxlength="12"
          placeholder="选填"
          placeholder-class="ph"
        />
      </FormRow>
      <view v-if="!isEdit" class="invite-hint">
        有朋友邀请你来？填写 TA 的邀请码，双方在资料审核通过后各解锁完整资料查看权益
      </view>
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
        <RegionPicker
          :province="form.province"
          :city="form.city"
          @change="onRegionChange($event, 'live')"
        />
      </FormRow>

      <FormRow label="户籍地" required align-right>
        <RegionPicker
          :province="form.home_province"
          :city="form.home_city"
          @change="onRegionChange($event, 'home')"
        />
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

      <!-- 行业标签：粗粒度，只到「系统」级。
           写「教育系统」是归属感，写「XX市第三中学」就是可反查到人，
           体制内用户对可被识别的信息极其敏感（见 docs/08 3.3） -->
      <FormRow label="所属行业" align-right>
        <text :class="{ ph: !form.industry }" @click="pickDict('industry', 'industry')">
          {{ store.label('industry', form.industry) || '选填' }}
        </text>
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
          color="#963E58"
            @change="req.accept_children = $event.detail.value ? 1 : 0"
          />
        </FormRow>
      </view>

      <!-- 原生家庭：最隐私的资料块，默认仅自己与审核员可见。
           「默认私有 + 用户主动授权」是刻意的设计 —— 默认公开会劝退大量用户，
           尤其是一期以体制内为主的用户群（见 docs/08 4.2① 与 6-2）。
           父母职业只到「类型」不到具体单位，防组合反查。 -->
      <view class="card">
        <view class="section-title">
          原生家庭
          <text class="section-sub">选填 · 默认仅审核员可见</text>
        </view>

        <FormRow label="父母职业" align-right>
          <text :class="{ ph: !form.family_parent_occ }" @click="pickDict('family_parent_occ', 'family_parent_occ')">
            {{ store.label('family_parent_occ', form.family_parent_occ) || '请选择' }}
          </text>
        </FormRow>

        <FormRow label="家庭氛围" align-right>
          <text :class="{ ph: !form.family_atmosphere }" @click="pickDict('family_atmosphere', 'family_atmosphere')">
            {{ store.label('family_atmosphere', form.family_atmosphere) || '请选择' }}
          </text>
        </FormRow>

        <FormRow label="是否独生子女" align-right>
          <text :class="{ ph: !form.family_only_child }" @click="pickDict('family_only_child', 'family_only_child')">
            {{ store.label('family_only_child', form.family_only_child) || '请选择' }}
          </text>
        </FormRow>

        <FormRow label="家庭所在地" align-right>
          <input
            v-model="form.family_city"
            class="ta-right"
            maxlength="16"
            placeholder="填到城市即可，如 河南 洛阳"
            placeholder-class="ph"
          />
        </FormRow>

        <FormRow label="补充说明" align-right>
          <input
            v-model="form.family_note"
            class="ta-right"
            maxlength="100"
            placeholder="选填，不超过 100 字"
            placeholder-class="ph"
          />
        </FormRow>

        <FormRow label="对已解锁用户可见" align-right>
          <switch
            :checked="form.family_visible === 1"
            color="#963E58"
            @change="form.family_visible = $event.detail.value ? 1 : 0"
          />
        </FormRow>

        <view class="tip">
          开启后，邀请好友通过审核的用户可看到以上内容；请勿填写单位全称、住址等可被反查到的信息。
        </view>
      </view>
    </view>
    </view>

    <!-- 底部操作 -->
    <view class="actions">
      <view v-if="step > 1" class="btn-plain half" @click="prev">上一步</view>
      <view
        class="btn-primary"
        :class="{ half: step > 1, disabled: submitting }"
        @click="next"
      >
        {{ submitting ? '提交中…' : step < 4 ? '下一步' : isEdit ? '保存' : '提交审核' }}
      </view>
    </view>
    <!-- 提交成功 -->
    <view v-if="doneVisible" class="done-mask">
      <view class="done-card">
        <view class="done-ring">
          <view class="done-check">✓</view>
        </view>
        <view class="done-title">{{ isEdit ? '已保存' : '提交成功' }}</view>
        <view class="done-desc">
          {{ isEdit ? '资料修改后需重新审核，期间不影响正常使用' : '资料已提交审核，通常 24 小时内完成' }}
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.setup {
  padding: 28rpx 24rpx 196rpx;
}

/* 退出出口：靠右、低对比，不做成按钮抢主流程的注意力 */
.topbar {
  display: flex;
  justify-content: flex-end;
  margin: -8rpx 0 12rpx;
}

.topbar-exit {
  padding: 8rpx 4rpx 8rpx 24rpx;
  font-size: $font-caption;
  color: $text-weak;
  transition: color 180ms ease;

  &:active {
    color: $brand;
  }
}

.steps {
  display: flex;
  align-items: center;
  padding: 12rpx 8rpx 0;
}

.step-item {
  display: flex;
  align-items: center;
  flex: 1;

  &:last-child {
    flex: 0;
  }

  .dot {
    width: 48rpx;
    height: 48rpx;
    line-height: 48rpx;
    text-align: center;
    border-radius: $radius-sm;
    border: 1rpx solid $border;
    background: $bg-card;
    color: $text-weak;
    font-size: $font-caption;
    flex-shrink: 0;
    font-variant-numeric: tabular-nums;
    position: relative;
    transition: background-color 220ms ease, border-color 220ms ease, color 220ms ease;

    /* 已完成的步骤：数字变对勾，弹一下，给「又推进了一格」的正反馈 */
    .tick {
      display: block;
      animation: tick-pop 320ms cubic-bezier(0.34, 1.56, 0.64, 1);
    }
  }

  .line {
    flex: 1;
    height: 2rpx;
    background: $border;
    transition: background-color 260ms ease;
  }

  &.active .dot {
    background: $brand;
    border-color: $brand;
    color: $text-on-accent;
  }

  &.active .line {
    background: $brand;
  }
}

@keyframes tick-pop {
  0% {
    transform: scale(0.4);
    opacity: 0;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

/* ---- 步骤切换舞台 ---- */
.stage {
  will-change: transform, opacity;
}

/* 前进：旧内容往左退场，新内容从右侧跟进；后退反之。
   时长与脚本里的 TRANS_OUT / TRANS_IN 保持一致（170ms / 300ms） */
.s-out-next {
  animation: stage-out-l 170ms ease-in forwards;
}
.s-in-next {
  animation: stage-in-r 300ms cubic-bezier(0.22, 0.8, 0.3, 1) both;
}
.s-out-prev {
  animation: stage-out-r 170ms ease-in forwards;
}
.s-in-prev {
  animation: stage-in-l 300ms cubic-bezier(0.22, 0.8, 0.3, 1) both;
}

@keyframes stage-out-l {
  from {
    transform: translateX(0);
    opacity: 1;
  }
  to {
    transform: translateX(-72rpx);
    opacity: 0;
  }
}
@keyframes stage-in-r {
  from {
    transform: translateX(72rpx);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
@keyframes stage-out-r {
  from {
    transform: translateX(0);
    opacity: 1;
  }
  to {
    transform: translateX(72rpx);
    opacity: 0;
  }
}
@keyframes stage-in-l {
  from {
    transform: translateX(-72rpx);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.step-title {
  font-size: $font-title;
  font-weight: 600;
  color: $text-main;
  letter-spacing: -1rpx;
  padding: 40rpx 4rpx 24rpx;

  .step-sub {
    font-size: $font-caption;
    color: $text-weak;
    font-weight: normal;
    margin-left: 12rpx;
  }
}

.card {
  background: $bg-raised;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 8rpx 28rpx;
  margin-bottom: 24rpx;
  box-shadow: $shadow-card;
}

.section-title {
  font-size: $font-lead;
  font-weight: 600;
  color: $text-main;
  padding: 28rpx 0 10rpx;
}

.ph {
  color: $text-placeholder;
}

.ta-right {
  text-align: right;
}

.avatar-row {
  display: flex;
  justify-content: flex-end;
}

/* 微信官方授权按钮：必须用原生 button，且要去掉默认样式让它看起来像普通控件 */
.avatar-btn {
  padding: 0;
  margin: 0;
  line-height: normal;
  background: transparent;
  border: none;
  display: flex;
  justify-content: flex-end;
  width: 100%;

  &::after {
    border: none;
  }
}

.wx-phone-btn {
  padding: 0;
  margin: 0;
  height: 60rpx;
  line-height: 60rpx;
  font-size: $font-body;
  background: transparent;
  border: none;
  color: $brand;
  font-weight: 600;
  text-align: left;

  &::after {
    border: none;
  }
}

.invite-hint {
  margin: -6rpx 4rpx 20rpx;
  font-size: $font-micro;
  color: $text-placeholder;
  line-height: 1.6;
}

.placeholder {
  width: 120rpx;
  height: 120rpx;
  border-radius: $radius;
  display: flex;
  align-items: center;
  justify-content: center;
  color: $text-placeholder;
  font-size: 44rpx;
  background: $bg-subtle;
  border: 1rpx dashed $border-strong;
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
  border-radius: $radius-sm;
  border: 1rpx solid $border;
  background: $bg-card;
  color: $text-sub;
  font-size: $font-mid;

  &.on {
    background: $brand;
    border-color: $brand;
    color: $text-on-accent;
  }
}

.locked {
  color: $text-weak;

  .lock-tip {
    font-size: $font-micro;
    color: $text-placeholder;
  }
}

.intro {
  width: 100%;
  height: 240rpx;
  font-size: $font-body;
  line-height: 1.6;
  padding: 18rpx 0;
  color: $text-main;
}

.count {
  text-align: right;
  font-size: $font-caption;
  color: $text-weak;

  &.warn {
    color: $warning;
  }
}

.tip {
  font-size: $font-caption;
  color: $text-placeholder;
  padding: 12rpx 0 20rpx;
  line-height: 1.5;
}

/* 「选填 · 默认仅审核员可见」这类小字，跟在 section 标题后 */
.section-sub {
  margin-left: 14rpx;
  font-size: $font-micro;
  font-weight: 400;
  color: $text-placeholder;
}

.pick-inline {
  display: inline-block;
  min-width: 70rpx;
  text-align: center;
  color: $brand;
  font-variant-numeric: tabular-nums;
}

.dash,
.unit {
  color: $text-weak;
  margin: 0 6rpx;
}

.multi {
  display: flex;
  justify-content: flex-end;
  gap: 12rpx;
}

.multi-item {
  padding: 8rpx 20rpx;
  border-radius: $radius-sm;
  border: 1rpx solid $border;
  background: $bg-card;
  color: $text-sub;
  font-size: $font-mid;
  transition: background-color 180ms ease, color 180ms ease, border-color 180ms ease;

  &.on {
    background: $brand-light;
    border-color: $brand-light;
    color: $brand;
  }
}

.actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: $bg-raised;
  border-top: 1rpx solid $border;
  display: flex;
  gap: 20rpx;
  box-shadow: 0 -12rpx 34rpx var(--color-shadow);
}

.half {
  flex: 1;
}

.actions .btn-primary {
  flex: 2;
}

/* ---- 提交成功覆盖层 ---- */
.done-mask {
  position: fixed;
  inset: 0;
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(34, 31, 32, 0.72);
  backdrop-filter: blur(6rpx);
  animation: done-mask-in 240ms ease both;
}

.done-card {
  width: 520rpx;
  padding: 56rpx 40rpx 48rpx;
  border-radius: $radius-lg;
  background: $bg-raised;
  box-shadow: $shadow-card;
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: done-card-in 420ms cubic-bezier(0.34, 1.4, 0.5, 1) both;
}

.done-ring {
  width: 132rpx;
  height: 132rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $brand-light;
  position: relative;
  margin-bottom: 32rpx;

  /* 外扩一圈涟漪，做出「完成」的余韵 */
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 2rpx solid $brand;
    animation: done-ripple 900ms ease-out 200ms both;
  }
}

.done-check {
  width: 84rpx;
  height: 84rpx;
  border-radius: 50%;
  background: $brand;
  color: $text-on-accent;
  font-size: 48rpx;
  line-height: 84rpx;
  text-align: center;
  /* 对勾用 stroke 的感觉：先缩放再回弹 */
  animation: done-check 460ms cubic-bezier(0.34, 1.56, 0.64, 1) 120ms both;
}

.done-title {
  font-size: $font-title;
  font-weight: 600;
  color: $text-main;
  letter-spacing: -1rpx;
  animation: done-text 360ms ease 300ms both;
}

.done-desc {
  margin-top: 14rpx;
  font-size: $font-caption;
  color: $text-weak;
  line-height: 1.6;
  text-align: center;
  animation: done-text 360ms ease 380ms both;
}

@keyframes done-mask-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
@keyframes done-card-in {
  from {
    transform: scale(0.86) translateY(24rpx);
    opacity: 0;
  }
  to {
    transform: scale(1) translateY(0);
    opacity: 1;
  }
}
@keyframes done-check {
  from {
    transform: scale(0);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}
@keyframes done-ripple {
  from {
    transform: scale(1);
    opacity: 0.7;
  }
  to {
    transform: scale(1.75);
    opacity: 0;
  }
}
@keyframes done-text {
  from {
    transform: translateY(12rpx);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
</style>
