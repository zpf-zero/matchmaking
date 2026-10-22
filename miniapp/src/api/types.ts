export interface User {
  id: number
  nickname: string
  avatar: string
  gender: number
  role: number
  status: number
  ban_reason?: string
  /**
   * 权益等级。0 = 无权益，≥1 = 已解锁完整资料（永久）。
   *
   * 自 2026-10 起业务从「到期时间」改为「永久等级」：
   * 赠送权益用倒计时没有心理锚点（用户不知道剩 7 天能换什么），
   * 且倒计时会加速犹豫型用户流失。详见 docs/08-权益体系重构方案.md 4.4。
   */
  vip_level?: number
  last_active_at?: string
  created_at?: string
}

/** 权益等级常量，与后端 model.VipLevelUnlocked 保持一致 */
export const VipLevelUnlocked = 1

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
  /** 行业标签（粗粒度，免费内容）。0 表示未填 */
  industry: number

  // ---- 原生家庭（最隐私一级，默认私有，需主动授权）----
  /** 父母职业类型。刻意只到「类型」不到具体单位，防反查 */
  family_parent_occ: number
  family_atmosphere: number
  /** 1 是 2 否 */
  family_only_child: number
  family_city: string
  family_note: string
  /** 1 表示同意对已解锁的用户展示 */
  family_visible: number

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
  /** 已结算但尚未查看的邀请数，「我的」页红点 */
  invite_unseen: number
}

/** 「谁看过我」列表。非会员只返回前 free_limit 条，其余 locked 条需权益解锁 */
export interface ViewedMeResp {
  list: UserCard[]
  /** 实际看过我的人数 */
  total: number
  /** 被锁住、需要权益才能看的人数 */
  locked: number
  /** 我是否已解锁完整资料（vip_level ≥ 1） */
  unlocked: boolean
  free_limit: number
}

/** 邀请记录。被邀请人昵称脱敏 */
export interface InviteRecord {
  nickname: string
  /** 0待达成（被邀请人尚未通过审核） 1已结算 2已作废 */
  status: number
  created_at: string
  settled_at: string | null
}

export interface MyInviteInfo {
  code: string
  invite_url: string
  settled_count: number
  pending_count: number
  max_settle: number
  /** 我是否已解锁完整资料（永久，无到期时间概念） */
  unlocked: boolean
  /** 已结算但我尚未查看。进邀请页时弹一次性回执，并据此清红点 */
  unseen_settled: number
  records: InviteRecord[]
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
  /** 行业标签（粗粒度），免费内容，用于「同系统」归属感 */
  industry: number
  industry_text: string
  match_rate: number
  active_text: string
}

/** 原生家庭（L2 权益内容，需对方主动授权公开） */
export interface FamilyInfo {
  parent_occ: number
  parent_occ_text: string
  atmosphere: number
  atmosphere_text: string
  only_child: number
  only_child_text: string
  city: string
  note: string
}

/** 对方的三观展示数据。未解锁时 answers 为空，仅 answered/total 有值 */
export interface ThreeViewInfo {
  /** 对方按题库顺序的选项下标，未答为 -1 */
  answers: number[]
  answered: number
  total: number
  /** 与我答案的一致率 0~100，仅在双方都答过时才有意义 */
  match_rate: number
  my_answers: number[]
  my_answered: boolean
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

  /** 审核通过的照片总数 */
  photo_total: number
  /** 被权益锁住、需解锁才能看的张数 */
  locked_photo_count: number
  photos_locked: boolean

  /** 原生家庭。仅当对方授权且我已解锁时才有值 */
  family: FamilyInfo | null
  /** 我是否有权限看对方的原生家庭。false 时 family 必为 null */
  viewed_family: boolean
  /** 对方三观答案。未解锁时为 null，但 three_view_answered 仍会告知对方答过题 */
  three_view: ThreeViewInfo | null
  three_view_answered: boolean
  three_view_unlocked: boolean

  is_liked: boolean
  liked_me: boolean
  is_matched: boolean
  conversation_id: number
  is_blocked: boolean
}

/** 三观题 */
export interface ThreeViewQuestion {
  id: number
  title: string
  options: string[]
}

/** 我的三观答题状态 */
export interface MyThreeViewResp {
  questions: ThreeViewQuestion[]
  /** 与 questions 同序，未答为 -1 */
  answers: number[]
  answered: number
  total: number
}

/** 客户端配置 */
export interface ClientConfig {
  mock_mode: boolean
  notice: string
  max_photo_num: number
  /** 免费可见照片张数，用于渲染「还有 N 张」的解锁提示 */
  free_photo_limit: number
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
