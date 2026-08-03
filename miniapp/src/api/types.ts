export interface User {
  id: number
  nickname: string
  avatar: string
  gender: number
  role: number
  status: number
  ban_reason?: string
  last_active_at?: string
  created_at?: string
}

export interface Profile {
  id?: number
  user_id?: number
  birthday: string
  age: number
  height: number
  weight: number
  marital_status: number
  has_children: number
  education: number
  school: string
  occupation: string
  income_level: number
  house_status: number
  car_status: number
  province: string
  city: string
  city_code: string
  home_province: string
  home_city: string
  self_intro: string
  complete_rate: number
  audit_status: number
  audit_remark: string
}

export interface Requirement {
  age_min: number
  age_max: number
  height_min: number
  height_max: number
  education_min: number
  income_min: number
  marital_status: string
  accept_children: number
  city_codes: string
  other: string
}

export interface Photo {
  id: number
  url: string
  sort: number
  audit_status: number
}

export interface MeResp {
  user: User
  phone: string
  profile: Profile | null
  requirement: Requirement | null
  photos: Photo[]
}

export interface UserCard {
  id: number
  nickname: string
  avatar: string
  gender: number
  age: number
  height: number
  education: number
  education_text: string
  income_level: number
  income_text: string
  marital_status: number
  marital_text: string
  city: string
  occupation: string
  match_rate: number
  active_text: string
}

export interface UserDetail extends UserCard {
  weight: number
  has_children: number
  school: string
  house_status: number
  car_status: number
  province: string
  home_province: string
  home_city: string
  self_intro: string
  photos: string[]
  is_liked: boolean
  liked_me: boolean
  is_matched: boolean
  conversation_id: number
  is_blocked: boolean
}

export interface Conversation {
  id: number
  peer_id: number
  peer_nickname: string
  peer_avatar: string
  last_message: string
  last_message_at: string | null
  unread: number
  matched_at: string | null
  peer_online: boolean
}

export interface Message {
  id: number
  conversation_id: number
  from_user_id: number
  type: number
  content: string
  is_read: number
  created_at: string
  /** 前端本地字段：发送中的临时标记 */
  _temp_id?: string
  _sending?: boolean
  _failed?: boolean
}

export interface DictItem {
  value: number
  label: string
}

export type Dict = Record<string, DictItem[]>

export interface AuditItem {
  user_id: number
  nickname: string
  avatar: string
  gender: number
  phone: string
  status: number
  created_at: string
  profile: Profile
  photos: Photo[]
}

export interface AdminUserItem {
  id: number
  nickname: string
  avatar: string
  phone: string
  gender: number
  role: number
  status: number
  ban_reason: string
  created_at: string
  last_active_at: string
  age: number
  city: string
  education: number
  income_level: number
  audit_status: number
}
