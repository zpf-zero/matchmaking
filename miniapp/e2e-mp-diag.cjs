#!/usr/bin/env node
/** 诊断脚本：dump 指定页面的 data 与关键元素文本 */
const automator = require('miniprogram-automator')

async function main() {
  const mini = await automator.connect({ wsEndpoint: 'ws://127.0.0.1:9420' })
  try {
    await mini.switchTab('/pages/index/index')
    await new Promise((r) => setTimeout(r, 2500))
    let page = await mini.currentPage()
    console.log('=== 首页 page.path:', page.path)
    const data = await page.data()
    const keys = Object.keys(data)
    console.log('data keys:', keys.join(', '))
    // 找列表类数据
    for (const k of keys) {
      const v = data[k]
      if (Array.isArray(v)) console.log(`  ${k}: Array[${v.length}]`, JSON.stringify(v).slice(0, 200))
      else if (v && typeof v === 'object' && (v.list || v.total !== undefined)) {
        console.log(`  ${k}:`, JSON.stringify(v).slice(0, 300))
      }
    }
    // 空态元素
    const empty = await page.$('.empty, .empty-title, .empty-sub')
    if (empty) console.log('空态文案:', await empty.text())
    const store = data.store || {}
    console.log('store.user:', JSON.stringify(store.user || data.user || null).slice(0, 200))
    console.log('store.status:', store.status, 'canBrowse:', store.canBrowse)

    // 设置页
    await mini.navigateTo('/pages/setting/index')
    await new Promise((r) => setTimeout(r, 2000))
    page = await mini.currentPage()
    const sd = await page.data()
    console.log('\n=== 设置页 keys:', Object.keys(sd).join(', '))
    const els = await page.$$('view')
    let n = 0
    for (const el of els) {
      const t = await el.text().catch(() => '')
      if (t && t.trim() && t.trim().length < 20 && n < 15) {
        console.log('  view:', t.trim())
        n++
      }
    }
  } finally {
    await mini.disconnect()
  }
}
main().catch((e) => { console.error('诊断失败:', e.message); process.exit(1) })
