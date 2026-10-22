import { http, type PageData } from '@/utils/request'
import type {
  AdminUserItem,
  AuditItem,
  ClientConfig,
  Conversation,
  Dict,
  MeResp,
  Message,
  MyInviteInfo,
  MyThreeViewResp,
  Profile,
  Requirement,
  ThreeViewQuestion,
  UserCard,
  UserDetail,
  ViewedMeResp,
} from './types'

export const authApi = {
  /** mock 模式下 code 直接作为 openid，换不同 code 可模拟多个账号 */
  wxLogin: (code: string) =>
    http.post<{ token: string; user: any; status: number; role: number; need_profile: boolean }>(
      '/auth/wx-login',
      { code },
      { noAuthRedirect: true }
    ),
  bindPhone: (data: { code?: string; phone?: string }) => http.post('/auth/bind-phone', data),
  logout: () => http.post('/auth/logout'),
  deleteAccount: () => http.del('/auth/account'),
}

export const inviteApi = {
  /** 我的邀请码与邀请进度 */
  myCode: () => http.get<MyInviteInfo>('/invite/my-code'),
  /** 绑定邀请码。在被邀请人通过人工审核后由后端结算双方权益 */
  bind: (code: string) => http.post('/invite/bind', { code }),
  /** 查看结算结果，清除「我的」页红点 */
  seen: () => http.post('/invite/seen'),
}

export const userApi = {  me: () => http.get<MeResp>('/user/me'),
  saveProfile: (data: Partial<Profile> & Record<string, any>) => http.put('/user/profile', data),
  saveRequirement: (data: Requirement) => http.put('/user/requirement', data),
  submitAudit: () => http.post('/user/submit-audit'),
  detail: (id: number) => http.get<UserDetail>(`/user/detail/${id}`),
  addPhoto: (url: string) => http.post('/user/photo', { url }),
  deletePhoto: (id: number) => http.del(`/user/photo/${id}`),
  sortPhotos: (ids: number[]) => http.put('/user/photo/sort', { ids }),

  /** 我的三观答题状态（含题库与我已选项） */
  myThreeView: () => http.get<MyThreeViewResp>('/user/three-view'),
  /** 提交三观答案。答题免费；-1 表示未答 */
  saveThreeView: (answers: number[]) => http.put('/user/three-view', { answers }),
}

export const discoverApi = {
  recommend: (page = 1, pageSize = 20) =>
    http.get<PageData<UserCard>>('/discover/recommend', { page, page_size: pageSize }),
  search: (params: Record<string, any>) =>
    http.get<PageData<UserCard>>('/discover/search', params),
}

export const interactionApi = {
  like: (toUserId: number) =>
    http.post<{ is_match: boolean; conversation_id: number; peer?: any }>('/interaction/like', {
      to_user_id: toUserId,
    }),
  pass: (toUserId: number) => http.post('/interaction/pass', { to_user_id: toUserId }),
  unlike: (toUserId: number) => http.del(`/interaction/like/${toUserId}`),
  likedMe: (page = 1) => http.get<PageData<UserCard>>('/interaction/liked-me', { page }),
  myLikes: (page = 1) => http.get<PageData<UserCard>>('/interaction/my-likes', { page }),
  viewedMe: (page = 1) => http.get<ViewedMeResp>('/interaction/viewed-me', { page }),
  unreadCount: () =>
    http.get<{ liked_me: number; viewed_me: number; message: number; total: number }>(
      '/interaction/unread-count',
      undefined,
      { silent: true }
    ),
  markRead: (type: number) => http.put(`/interaction/read?type=${type}`),
}

export const chatApi = {
  conversations: () => http.get<Conversation[]>('/conversation/list'),
  closeConversation: (id: number) => http.del(`/conversation/${id}`),
  messages: (conversationId: number, beforeId = 0, limit = 20) =>
    http.get<Message[]>('/message/list', {
      conversation_id: conversationId,
      before_id: beforeId,
      limit,
    }),
  send: (conversationId: number, content: string, msgType = 1) =>
    http.post<Message>('/message/send', {
      conversation_id: conversationId,
      content,
      msg_type: msgType,
    }),
  read: (conversationId: number) => http.put('/message/read', { conversation_id: conversationId }),
  recall: (messageId: number) => http.put('/message/recall', { message_id: messageId }),
}

export const commonApi = {
  dict: () => http.get<Dict>('/common/dict'),
  config: () => http.get<ClientConfig>(
    '/common/config',
    undefined,
    { noAuthRedirect: true }
  ),
  /** 三观题库。答题免费，题目本身无隐私，登录即可取 */
  threeViewQuestions: () => http.get<ThreeViewQuestion[]>('/common/three-view/questions'),
  report: (data: {
    target_user_id: number
    reason: number
    description?: string
    images?: string
  }) => http.post('/common/report', data),
  addBlacklist: (toUserId: number) => http.post('/common/blacklist', { to_user_id: toUserId }),
  removeBlacklist: (id: number) => http.del(`/common/blacklist/${id}`),
  blacklist: () => http.get<{ user_id: number; nickname: string; avatar: string }[]>(
    '/common/blacklist'
  ),
}

export const adminApi = {
  overview: () =>
    http.get<{
      total_user: number
      today_new: number
      pending_audit: number
      pending_report: number
      matched: number
      active_week: number
    }>('/admin/overview'),
  auditList: (page = 1) => http.get<PageData<AuditItem>>('/admin/audit/list', { page }),
  audit: (userId: number, pass: boolean, remark = '') =>
    http.post(`/admin/audit/${userId}`, { pass, remark }),
  userList: (params: { keyword?: string; status?: number; page?: number }) =>
    http.get<PageData<AdminUserItem>>('/admin/user/list', params),
  userDetail: (id: number) => http.get<any>(`/admin/user/${id}`),
  ban: (id: number, reason: string) => http.post(`/admin/user/${id}/ban`, { reason }),
  unban: (id: number) => http.post(`/admin/user/${id}/unban`),
  reportList: (status = -1, page = 1) =>
    http.get<PageData<any>>('/admin/report/list', { status, page }),
  handleReport: (id: number, action: 'ignore' | 'ban', remark = '') =>
    http.post(`/admin/report/${id}/handle`, { action, remark }),
  /** 任命/撤销管理角色。仅超级管理员可用，role: 1 普通用户 / 2 审核员 / 3 管理员 */
  setRole: (id: number, role: number) => http.post(`/admin/user/${id}/role`, { role }),
}
