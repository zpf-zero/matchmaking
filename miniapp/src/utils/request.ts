/** 后端地址。开发阶段微信开发者工具需勾选「不校验合法域名」 */
export const BASE_URL = 'http://localhost:8080'
export const API_URL = BASE_URL + '/api/v1'

export interface Resp<T = any> {
  code: number
  message: string
  data: T
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

/** 业务状态码，与后端 internal/pkg/response 保持一致 */
export const Code = {
  Success: 0,
  ParamError: 400,
  Unauthorized: 401,
  Forbidden: 403,
  ProfileIncomplete: 4001,
  ProfileAuditing: 4002,
  ProfileRejected: 4003,
  Banned: 4004,
} as const

/**
 * 跳转去重。
 * 一个页面往往并发发好几个请求，token 过期时它们会同时拿到 401，
 * 不去重的话会连续 reLaunch 好几次，页面栈直接乱掉。
 */
let redirecting = false
function redirectOnce(fn: () => void) {
  if (redirecting) return
  redirecting = true
  fn()
  setTimeout(() => (redirecting = false), 1200)
}

interface Options {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, any>
  /** 不自动弹错误提示，由调用方自己处理 */
  silent?: boolean
  /** 不自动处理登录态跳转 */
  noAuthRedirect?: boolean
}

export function request<T = any>(options: Options): Promise<T> {
  const token = uni.getStorageSync('token')

  return new Promise<T>((resolve, reject) => {
    uni.request({
      url: API_URL + options.url,
      method: options.method || 'GET',
      data: options.data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      timeout: 15000,
      success(res) {
        const body = res.data as Resp<T>

        if (!body || typeof body.code !== 'number') {
          if (!options.silent) uni.showToast({ title: '服务异常', icon: 'none' })
          reject(new Error('响应格式异常'))
          return
        }

        if (body.code === Code.Success) {
          resolve(body.data)
          return
        }

        handleBizError(body, options)
        reject(body)
      },
      fail(err) {
        if (!options.silent) {
          uni.showToast({ title: '网络连接失败', icon: 'none' })
        }
        reject(err)
      },
    })
  })
}

function handleBizError(body: Resp, options: Options) {
  if (options.noAuthRedirect) {
    if (!options.silent) uni.showToast({ title: body.message, icon: 'none' })
    return
  }

  switch (body.code) {
    case Code.Unauthorized:
      uni.removeStorageSync('token')
      redirectOnce(() => uni.reLaunch({ url: '/pages/login/index' }))
      break

    case Code.ProfileIncomplete:
      redirectOnce(() => {
        uni.showToast({ title: '请先完善资料', icon: 'none' })
        setTimeout(() => uni.reLaunch({ url: '/pages/profile/setup' }), 800)
      })
      break

    case Code.ProfileAuditing:
    case Code.ProfileRejected:
    case Code.Banned:
      redirectOnce(() =>
        uni.navigateTo({
          url: `/pages/common/status?code=${body.code}&msg=${encodeURIComponent(body.message)}`,
        })
      )
      break

    default:
      if (!options.silent) {
        uni.showToast({ title: body.message || '操作失败', icon: 'none' })
      }
  }
}

export const http = {
  get: <T = any>(url: string, data?: any, opts?: Partial<Options>) =>
    request<T>({ url, method: 'GET', data, ...opts }),
  post: <T = any>(url: string, data?: any, opts?: Partial<Options>) =>
    request<T>({ url, method: 'POST', data, ...opts }),
  put: <T = any>(url: string, data?: any, opts?: Partial<Options>) =>
    request<T>({ url, method: 'PUT', data, ...opts }),
  del: <T = any>(url: string, data?: any, opts?: Partial<Options>) =>
    request<T>({ url, method: 'DELETE', data, ...opts }),
}

/** 图片上传 */
export function uploadImage(filePath: string, scene: string): Promise<string> {
  const token = uni.getStorageSync('token')
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: `${API_URL}/common/upload/image?scene=${scene}`,
      filePath,
      name: 'file',
      header: token ? { Authorization: `Bearer ${token}` } : {},
      success(res) {
        try {
          const body = JSON.parse(res.data) as Resp<{ url: string }>
          if (body.code === Code.Success) {
            resolve(body.data.url)
          } else {
            uni.showToast({ title: body.message, icon: 'none' })
            reject(body)
          }
        } catch (e) {
          reject(e)
        }
      },
      fail: reject,
    })
  })
}
