import { http, type PageData } from '@/utils/request'
import type {
  AdminUserItem,
  AuditItem,
  Conversation,
  Dict,
  MeResp,
  Message,
  Profile,
  Requirement,
  UserCard,
  UserDetail,
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

export const userApi = {
  me: () => http.get<MeResp>('/user/me'),
  saveProfile: (data: Partial<Profile> & Record<string, any>) => http.put('/user/profile', data),
  saveRequirement: (data: Requirement) => http.put('/user/requirement', data),
  submitAudit: () => http.post('/user/submit-audit'),
  detail: (id: number) => http.get<UserDetail>(`/user/detail/${id}`),
  addPhoto: (url: string) => http.post('/user/photo', { url }),
  deletePhoto: (id: number) => http.del(`/user/photo/${id}`),
  sortPhotos: (ids: number[]) => http.put('/user/photo/sort', { ids }),
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
  viewedMe: (page = 1) => http.get<PageData<UserCard>>('/interaction/viewed-me', { page }),
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
}

export const commonApi = {
  dict: () => http.get<Dict>('/common/dict'),
  config: () => http.get<{ mock_mode: boolean; notice: string; max_photo_num: number }>(
    '/common/config',
    undefined,
    { noAuthRedirect: true }
  ),
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
}
