/**
 * 解析服务端返回的时间字符串，返回时间戳；解析不出来返回 NaN，由调用方决定怎么兜底。
 *
 * 服务端是 Go，time.Time 序列化成 RFC3339（"2026-08-12T21:17:00+08:00"），
 * 直接 Date.parse 就能处理。之前各页面自己写的 `t.replace(/-/g, '/')` 是给
 * MySQL 风格的 "2026-08-12 21:17:00" 准备的老 iOS Safari 兼容写法，
 * 套到 RFC3339 上会把日期部分改成 "2026/08/12T..."，直接解析失败变成 NaN。
 * 这里先按标准格式解析，失败了再退回那套替换，两种格式都能吃。
 */
export function parseTime(t?: string | null): number {
  if (!t) return NaN

  const ts = Date.parse(t)
  if (!Number.isNaN(ts)) return ts

  // 退路：老 WebView 解析不了带连字符的 "2026-08-12 21:17:00"
  return new Date(t.replace(/-/g, '/').replace(/\..*$/, '')).getTime()
}
