import { BASE_URL } from './request'

/**
 * WebSocket 封装：心跳 + 断线重连 + 事件订阅。
 *
 * 小程序切后台会被系统断连，回前台必须能自动接上，
 * 所以重连不是可选项。指数退避避免服务端重启时被反复冲击。
 */

type Handler = (data: any) => void

const HEARTBEAT_INTERVAL = 25000 // 服务端 60s 超时，25s 一次留足余量
const MAX_RETRY_DELAY = 30000

class WSClient {
  private task: UniApp.SocketTask | null = null
  private handlers = new Map<string, Set<Handler>>()
  private heartbeatTimer: any = null
  private reconnectTimer: any = null
  private retryCount = 0
  private manualClosed = false
  private connecting = false

  get connected() {
    return !!this.task && !this.connecting
  }

  connect() {
    const token = uni.getStorageSync('token')
    if (!token || this.task || this.connecting) return

    this.manualClosed = false
    this.connecting = true

    const task = uni.connectSocket({
      url: `${BASE_URL.replace(/^http/, 'ws')}/ws?token=${token}`,
      complete: () => {},
    })
    this.task = task

    task.onOpen(() => {
      this.connecting = false
      this.retryCount = 0
      this.startHeartbeat()
      this.emit('__open', null)
    })

    task.onMessage((res) => {
      try {
        const env = JSON.parse(res.data as string)
        if (env.type === 'pong') return
        this.emit(env.type, env.data)
      } catch {
        // 忽略无法解析的帧
      }
    })

    task.onClose(() => {
      this.cleanup()
      this.scheduleReconnect()
    })

    task.onError(() => {
      this.connecting = false
      this.cleanup()
      this.scheduleReconnect()
    })
  }

  close() {
    this.manualClosed = true
    clearTimeout(this.reconnectTimer)
    try {
      this.task?.close({})
    } catch {
      // 已经断开则忽略
    }
    this.cleanup()
  }

  send(type: string, data?: any) {
    if (!this.task) {
      this.connect()
      return false
    }
    this.task.send({ data: JSON.stringify({ type, data }) })
    return true
  }

  on(type: string, handler: Handler) {
    if (!this.handlers.has(type)) this.handlers.set(type, new Set())
    this.handlers.get(type)!.add(handler)
  }

  off(type: string, handler: Handler) {
    this.handlers.get(type)?.delete(handler)
  }

  private emit(type: string, data: any) {
    this.handlers.get(type)?.forEach((h) => h(data))
  }

  private startHeartbeat() {
    clearInterval(this.heartbeatTimer)
    this.heartbeatTimer = setInterval(() => {
      this.task?.send({ data: JSON.stringify({ type: 'ping' }) })
    }, HEARTBEAT_INTERVAL)
  }

  private cleanup() {
    clearInterval(this.heartbeatTimer)
    this.heartbeatTimer = null
    this.task = null
    this.connecting = false
  }

  private scheduleReconnect() {
    if (this.manualClosed || !uni.getStorageSync('token')) return

    // 1s, 2s, 4s, 8s... 最多 30s
    const delay = Math.min(1000 * 2 ** this.retryCount, MAX_RETRY_DELAY)
    this.retryCount++

    clearTimeout(this.reconnectTimer)
    this.reconnectTimer = setTimeout(() => this.connect(), delay)
  }
}

export const ws = new WSClient()
