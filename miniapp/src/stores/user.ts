import { defineStore } from 'pinia'
import { authApi, commonApi, userApi } from '@/api'
import type {
  Dict,
  MeResp,
  Photo,
  Profile,
  Requirement,
  ThreeViewQuestion,
  User,
} from '@/api/types'
import { VipLevelUnlocked } from '@/api/types'
import { ws } from '@/utils/ws'

/** 用户状态，与后端 model/enum.go 保持一致 */
export const Status = {
  Incomplete: 0,
  Auditing: 1,
  Normal: 2,
  Rejected: 3,
  Banned: 4,
} as const

/** 角色，向上包含 */
export const Role = {
  User: 1,
  Auditor: 2,
  Admin: 3,
  SuperAdmin: 4,
} as const

/** 举报处理状态，与后端 model/enum.go 保持一致 */
export const ReportStatus = {
  Pending: 0,
  Handled: 1,
  Rejected: 2,
} as const

interface State {
  token: string
  user: User | null
  phone: string
  profile: Profile | null
  requirement: Requirement | null
  photos: Photo[]
  dict: Dict
  mockMode: boolean
  configLoaded: boolean
  configError: boolean
  /**
   * 是否已完成过至少一次 fetchMe。
   *
   * 页面 onShow 早于 App.vue onLaunch 里的 `await fetchMe()` 触发，
   * 那一刻 user 还是 null、canBrowse/canInteract 一律为 false。
   * 只有这个标志能区分「还没拿到用户」和「用户确实没权限」——
   * 少了它，页面会把「等待中」误判成「无权限」。
   */
  meLoaded: boolean
  /** fetchMe 失败（网络异常等）。配合 meLoaded 让页面能给出重试入口，而不是无限转圈 */
  meError: boolean
  unread: { liked_me: number; viewed_me: number; message: number; total: number }
  /**
   * 三观题库。全局缓存一份 —— 题目不随用户变化，
   * 答题页、他人资料页都要用，没必要每页各拉一次。
   */
  threeViewQuestions: ThreeViewQuestion[]
  /** 免费可见的照片张数。来自 /common/config，用于渲染「还有 N 张」解锁提示 */
  freePhotoLimit: number
  /** 相册上限。来自 /common/config，与后端 MaxPhotoNum 同源 */
  maxPhotoNum: number
  /**
   * 私聊聊天远程开关。来自 /common/config，默认 true ——
   * 配置还没拉回来时不要把功能藏了，否则冷启动会闪一下「未开放」
   */
  privateChat: boolean
  /**
   * 已结算但尚未查看的邀请数。好友过审的那一刻双方权益到账，
   * 邀请人却多半不在场 —— 这个数字驱动「我的」页「邀请有礼」红点，
   * 进邀请页即清除（POST /invite/seen）
   */
  inviteUnseen: number
}

export const useUserStore = defineStore('user', {
  state: (): State => ({
    token: uni.getStorageSync('token') || '',
    user: null,
    phone: '',
    profile: null,
    requirement: null,
    photos: [],
    dict: {},
    mockMode: false,
    configLoaded: false,
    configError: false,
    meLoaded: false,
    meError: false,
    unread: { liked_me: 0, viewed_me: 0, message: 0, total: 0 },
    threeViewQuestions: [],
    freePhotoLimit: 3,
    maxPhotoNum: 9,
    privateChat: true,
    inviteUnseen: 0,
  }),

  getters: {
    isLogin: (s) => !!s.token,
    status: (s) => s.user?.status ?? Status.Incomplete,
    role: (s) => s.user?.role ?? Role.User,
    /** role > 1 时「我的」页才渲染管理入口。真正的拦截在后端 */
    isAdminRole: (s) => (s.user?.role ?? 1) > Role.User,
    canManageUser: (s) => (s.user?.role ?? 1) >= Role.Admin,
    /**
     * 是否需要先去完善资料。
     *
     * 判据是「普通用户且资料未提交」—— 运营角色（审核员及以上）不参与择偶，
     * 不该被资料向导拦住。把这条判断收敛到这里，避免登录页、App 守卫
     * 各写一遍 status===0 而漏掉 role。
     */
    needsProfile: (s) => (s.user?.role ?? 1) === Role.User && (s.user?.status ?? 0) === Status.Incomplete,
    /** 运营角色的落地页：后台工作台 */
    homePath: (s) =>
      (s.user?.role ?? 1) > Role.User ? '/pages/admin/index' : '/pages/index/index',
    /** 资料已提交，可以浏览他人 */
    canBrowse: (s) => {
      const st = s.user?.status ?? 0
      return st === Status.Auditing || st === Status.Normal || st === Status.Rejected
    },
    /** 审核通过，可以发起互动 */
    canInteract: (s) => s.user?.status === Status.Normal,
    /**
     * 是否已解锁完整资料。
     *
     * 自 2026-10 权益改为「永久等级」：判定只看 vip_level ≥ 1，
     * 不再有到期时间。这样用户不需要理解任何规则，也不会产生
     * 「快到期了要不要续」的咨询工单（见 docs/08-权益体系重构方案.md 4.4）。
     */
    isVip: (s) => (s.user?.vip_level ?? 0) >= VipLevelUnlocked,
    /** 私聊聊天是否开放。关闭时消息列表、聊天按钮、配对后「去聊天」全部隐藏 */
    privateChatEnabled: (s) => s.privateChat,
  },

  actions: {
    async login(code: string) {
      const res = await authApi.wxLogin(code)
      this.token = res.token
      uni.setStorageSync('token', res.token)
      this.user = res.user
      // 登录响应里已带回 user，但资料/择偶要求要等 fetchMe。
      // 这里不置 meLoaded，由调用方紧接着的 fetchMe 统一置位。
      return res
    },

    /** H5 手机号验证码登录。手机号不存在时后端自动建号 */
    async phoneLogin(phone: string, code: string) {
      const res = await authApi.phoneLogin(phone, code)
      this.token = res.token
      uni.setStorageSync('token', res.token)
      this.user = res.user
      return res
    },

    async fetchMe() {
      try {
        const me: MeResp = await userApi.me()
        this.user = me.user
        this.phone = me.phone
        this.profile = me.profile
        this.requirement = me.requirement
        this.photos = me.photos || []
        this.inviteUnseen = me.invite_unseen || 0
        this.meLoaded = true
        this.meError = false
        return me
      } catch (e) {
        // 标记失败，页面据此显示重试入口而不是无限转圈。
        // meLoaded 不置位：下次调用仍算「首次加载」
        this.meError = true
        throw e
      }
    },

    async fetchDict() {
      if (Object.keys(this.dict).length) return this.dict
      this.dict = await commonApi.dict()
      return this.dict
    },

    async fetchConfig() {
      this.configError = false
      try {
        const cfg = await commonApi.config()
        this.mockMode = cfg.mock_mode
        this.freePhotoLimit = cfg.free_photo_limit ?? 3
        this.maxPhotoNum = cfg.max_photo_num ?? 9
        // 老后端没有这个字段时按「开放」处理，避免升级顺序不当把聊天误关
        this.privateChat = cfg.private_chat !== false
      } catch {
        this.mockMode = false
        this.configError = true
      } finally {
        this.configLoaded = true
      }
    },

    /** 拉三观题库（全局缓存一次）。答题页与展示页共用 */
    async fetchThreeViewQuestions(force = false) {
      if (!force && this.threeViewQuestions.length) return this.threeViewQuestions
      this.threeViewQuestions = await commonApi.threeViewQuestions()
      return this.threeViewQuestions
    },

    async fetchUnread() {
      if (!this.canBrowse) return
      try {
        this.unread = await (await import('@/api')).interactionApi.unreadCount()
        this.syncTabBadge()
      } catch {
        // 红点失败静默处理
      }
    },

    /** 消息 Tab 角标。挂在 fetchUnread 后自动同步 */
    syncTabBadge() {
      const n = this.unread.message + this.unread.liked_me
      if (n > 0) {
        uni.setTabBarBadge({
          index: 2,
          text: n > 99 ? '99+' : String(n),
          fail: () => {}, // 当前不在 tabBar 页面时会失败，忽略
        })
      } else {
        uni.removeTabBarBadge({ index: 2, fail: () => {} })
      }
    },

    /** 字典值转文字 */
    label(type: string, value: number) {
      return this.dict[type]?.find((i) => i.value === value)?.label || ''
    },

    logout() {
      this.token = ''
      this.user = null
      this.profile = null
      this.requirement = null
      this.photos = []
      this.inviteUnseen = 0
      this.meLoaded = false
      this.meError = false
      uni.removeStorageSync('token')
      ws.close()
      uni.reLaunch({ url: '/pages/login/index' })
    },
  },
})
