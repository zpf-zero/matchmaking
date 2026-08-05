import { defineStore } from 'pinia'
import { authApi, commonApi, userApi } from '@/api'
import type { Dict, MeResp, Photo, Profile, Requirement, User } from '@/api/types'
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

interface State {
  token: string
  user: User | null
  phone: string
  profile: Profile | null
  requirement: Requirement | null
  photos: Photo[]
  dict: Dict
  mockMode: boolean
  unread: { liked_me: number; viewed_me: number; message: number; total: number }
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
    mockMode: true,
    unread: { liked_me: 0, viewed_me: 0, message: 0, total: 0 },
  }),

  getters: {
    isLogin: (s) => !!s.token,
    status: (s) => s.user?.status ?? Status.Incomplete,
    role: (s) => s.user?.role ?? Role.User,
    /** role > 1 时「我的」页才渲染管理入口。真正的拦截在后端 */
    isAdminRole: (s) => (s.user?.role ?? 1) > Role.User,
    canManageUser: (s) => (s.user?.role ?? 1) >= Role.Admin,
    /** 资料已提交，可以浏览他人 */
    canBrowse: (s) => {
      const st = s.user?.status ?? 0
      return st === Status.Auditing || st === Status.Normal || st === Status.Rejected
    },
    /** 审核通过，可以发起互动 */
    canInteract: (s) => s.user?.status === Status.Normal,
  },

  actions: {
    async login(code: string) {
      const res = await authApi.wxLogin(code)
      this.token = res.token
      uni.setStorageSync('token', res.token)
      this.user = res.user
      return res
    },

    async fetchMe() {
      const me: MeResp = await userApi.me()
      this.user = me.user
      this.phone = me.phone
      this.profile = me.profile
      this.requirement = me.requirement
      this.photos = me.photos || []
      return me
    },

    async fetchDict() {
      if (Object.keys(this.dict).length) return this.dict
      this.dict = await commonApi.dict()
      return this.dict
    },

    async fetchConfig() {
      try {
        const cfg = await commonApi.config()
        this.mockMode = cfg.mock_mode
      } catch {
        // 配置拉不到不影响主流程
      }
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
      uni.removeStorageSync('token')
      ws.close()
      uni.reLaunch({ url: '/pages/login/index' })
    },
  },
})
