const { chromium } = require('playwright')
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const BASE = 'http://localhost:5199/'
const CHROME = 'C:/Program Files/Google/Chrome/Application/chrome.exe'

async function login(page, account) {
  await page.goto(BASE, { waitUntil: 'domcontentloaded', timeout: 60000 })
  await page.evaluate(() => localStorage.clear())
  await page.reload({ waitUntil: 'domcontentloaded' })
  await page.waitForSelector('.btn-primary', { timeout: 30000 })
  await sleep(1200)
  await page.locator('input').first().fill(account)
  await page.locator('.checkbox').first().click()
  await sleep(200)
  await page.locator('.btn-primary').first().click()
  await sleep(6000)
  return page.url()
}

const issues = []
const issue = (s, a, m) => issues.push(`[${s}] ${a}: ${m}`)

;(async () => {
  const browser = await chromium.launch({ executablePath: CHROME, headless: true })
  const ctx = await browser.newContext({ viewport: { width: 414, height: 896 } })

  // ========== 管理端全页走查 ==========
  const page = await ctx.newPage()
  const bad = []
  page.on('response', async (res) => {
    const u = res.url()
    if (u.includes('/api/v1/')) {
      let c = '?'
      try { c = (await res.json()).code } catch {}
      if (c !== 0 && !u.includes('/common/config')) bad.push(`${u.split('/api/v1')[1]} -> ${c}`)
    }
  })
  page.on('pageerror', (e) => issue('JS', 'pageerror', e.message))

  await login(page, 'admin')
  console.log('=== 管理端全页走查 ===')
  const adminPages = [
    ['工作台', '#/pages/admin/index'],
    ['资料审核', '#/pages/admin/audit'],
    ['用户管理', '#/pages/admin/user'],
    ['举报处理', '#/pages/admin/report'],
  ]
  for (const [name, hash] of adminPages) {
    await page.goto(BASE + hash, { waitUntil: 'domcontentloaded' })
    await sleep(3200)
    const txt = (await page.locator('body').innerText()).replace(/\n+/g, ' ').slice(0, 60)
    console.log(`  [${name}] ${txt}`)
    // 检查是否有 AdminBack 返回条（后台子页应有）
    if (name !== '工作台') {
      const back = await page.locator('.back, .admin-back, [class*=back]').count()
      if (back === 0) issue('LOW', name, '没有找到返回条（可能是 class 不匹配，需人工确认）')
    }
  }

  // 工作台各入口可点
  await page.goto(BASE + '#/pages/admin/index', { waitUntil: 'domcontentloaded' })
  await sleep(3000)
  const entries = await page.locator('.q-item, .item').count()
  console.log('  工作台可点入口 =', entries)

  console.log('\n  管理端非0业务码:', bad.length ? bad.join(' | ') : '无')
  await page.close()

  // ========== 用户端全页走查 ==========
  const up = await ctx.newPage()
  const ubad = []
  up.on('response', async (res) => {
    const u = res.url()
    if (u.includes('/api/v1/')) {
      let c = '?'
      try { c = (await res.json()).code } catch {}
      if (c !== 0 && !u.includes('/common/config')) ubad.push(`${u.split('/api/v1')[1]} -> ${c}`)
    }
  })

  console.log('\n=== 用户端全页走查 (test_user_1) ===')
  await login(up, 'test_user_1')
  const userPages = [
    ['推荐', '#/'],
    ['筛选', '#/pages/search/index'],
    ['消息', '#/pages/message/index'],
    ['我的', '#/pages/mine/index'],
    ['设置', '#/pages/setting/index'],
    ['黑名单', '#/pages/setting/blacklist'],
  ]
  for (const [name, hash] of userPages) {
    await up.goto(BASE + hash, { waitUntil: 'domcontentloaded' })
    await sleep(2800)
    const txt = (await up.locator('body').innerText()).replace(/\n+/g, ' ').slice(0, 70)
    console.log(`  [${name}] ${txt}`)
    // 页面不应出现未捕获的空白
    if (txt.trim().length < 3) issue('HIGH', name, '页面内容为空')
  }

  // 用户详情页（点推荐卡片）
  await up.goto(BASE + '#/', { waitUntil: 'domcontentloaded' })
  await sleep(3000)
  const cardCnt = await up.locator('.card, [class*=ucard], [class*=user-card]').count()
  console.log('  [推荐] 卡片数 =', cardCnt)
  if (cardCnt > 0) {
    await up.locator('.card, [class*=ucard], [class*=user-card]').first().click()
    await sleep(3000)
    console.log('  [用户详情] URL =', up.url())
    const txt = (await up.locator('body').innerText()).replace(/\n+/g, ' ').slice(0, 80)
    console.log('  [用户详情]', txt)
  }

  console.log('\n  用户端非0业务码:', ubad.length ? ubad.join(' | ') : '无')

  console.log('\n================ 问题汇总 ================')
  if (!issues.length) console.log('  未发现明显问题')
  else issues.forEach((x, i) => console.log(`  ${i + 1}. ${x}`))

  await browser.close()
})()
