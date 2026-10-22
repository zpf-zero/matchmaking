const { chromium } = require('playwright')
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const BASE = 'http://localhost:5199/'
const CHROME = 'C:/Program Files/Google/Chrome/Application/chrome.exe'

const issues = []
const notes = []
function issue(sev, area, msg) {
  issues.push(`[${sev}] ${area}: ${msg}`)
}
function note(msg) { notes.push(msg) }

async function boot(ctx) {
  const page = await ctx.newPage()
  const apiBad = []
  page.on('response', async (res) => {
    const u = res.url()
    if (u.includes('/api/v1/')) {
      let c = '?'
      try { c = (await res.json()).code } catch {}
      if (c !== 0) apiBad.push(`${u.split('/api/v1')[1]} -> ${res.status()}/${c}`)
    }
  })
  page.on('pageerror', (e) => issue('JS', 'pageerror', e.message))
  page._apiBad = apiBad
  return page
}

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

;(async () => {
  const browser = await chromium.launch({ executablePath: CHROME, headless: true })
  const ctx = await browser.newContext({ viewport: { width: 414, height: 896 } })

  // ============ 管理端 ============
  let page = await boot(ctx)
  console.log('###### 管理端 admin ######')
  console.log('登录落点:', await login(page, 'admin'))

  // 首页各项
  const stat = await page.locator('.q-item, .cell').count()
  console.log('首页统计块 =', stat)

  // 用户管理
  await page.goto(BASE + '#/pages/admin/user', { waitUntil: 'domcontentloaded' })
  await sleep(3500)
  let n = await page.locator('.item').count()
  console.log('\n[用户管理] 列表条数 =', n)
  if (n === 0) issue('HIGH', '用户管理', '列表为空')

  // 筛选「待审核」
  const pendingFilter = page.locator('.fi', { hasText: '待审核' })
  if (await pendingFilter.count()) {
    await pendingFilter.first().click()
    await sleep(2500)
    const c = await page.locator('.item').count()
    console.log('[用户管理] 筛选「待审核」后 =', c, '条')
    // 点进第一个待审用户详情
    if (c > 0) {
      const first = await page.locator('.item').first().innerText()
      console.log('[用户管理] 首个待审用户 =', first.replace(/\n+/g, ' | ').slice(0, 90))
      await page.locator('.item').first().click()
      await sleep(3000)
      console.log('[用户详情] URL =', page.url())
      const acts = await page.locator('.actions .btn').allInnerTexts().catch(() => [])
      console.log('[用户详情] 操作按钮 =', JSON.stringify(acts))
      const body = await page.locator('body').innerText()
      const hasAuditBtn = acts.some((a) => a.includes('通过') || a.includes('审核'))
      if (!hasAuditBtn) {
        issue('HIGH', '用户详情页', `待审核用户在详情页没有「审核通过」入口，只有 ${JSON.stringify(acts)}`)
      }
      if (!body.includes('待审核')) note('用户详情页显示了状态标签')
    }
  }

  // 举报页
  await page.goto(BASE + '#/pages/admin/report', { waitUntil: 'domcontentloaded' })
  await sleep(3000)
  const rpt = await page.locator('.card').count()
  console.log('\n[举报处理] 卡片数 =', rpt)

  console.log('\n[管理端] 非0业务码 =', page._apiBad.length ? page._apiBad.join(' | ') : '无')
  await page.close()

  // ============ 用户端 ============
  page = await boot(ctx)
  console.log('\n###### 用户端 test_user_1 ######')
  const u1url = await login(page, 'test_user_1')
  console.log('登录落点:', u1url)
  if (!u1url.endsWith('#/')) issue('MED', '用户端', '普通用户登录未落到推荐流首页')
  await sleep(2500)

  const cards = await page.locator('.card, .user-card, .ucard').count()
  const recBody = await page.locator('body').innerText()
  console.log('[推荐流] 卡片数 =', cards, '| 含「推荐」=', recBody.includes('推荐'))
  if (cards === 0) issue('HIGH', '推荐流', '首页没有推荐卡片')

  // 各 tab
  for (const [name, hash] of [
    ['筛选', '#/pages/search/index'],
    ['消息', '#/pages/message/index'],
    ['我的', '#/pages/mine/index'],
  ]) {
    await page.goto(BASE + hash, { waitUntil: 'domcontentloaded' })
    await sleep(2600)
    const t = (await page.locator('body').innerText()).replace(/\n+/g, ' ').slice(0, 70)
    console.log(`[${name}] ${t}`)
  }

  // 设置页
  await page.goto(BASE + '#/pages/setting/index', { waitUntil: 'domcontentloaded' })
  await sleep(2500)
  console.log('[设置] ' + (await page.locator('body').innerText()).replace(/\n+/g, ' ').slice(0, 90))

  console.log('\n[用户端] 非0业务码 =', page._apiBad.length ? page._apiBad.join(' | ') : '无')
  await page.close()

  // ============ 汇总 ============
  console.log('\n================ 问题汇总 ================')
  if (issues.length === 0) console.log('  未发现明显问题')
  else issues.forEach((x, i) => console.log(`  ${i + 1}. ${x}`))
  console.log('\n---------------- 备注 ----------------')
  notes.forEach((x) => console.log('  · ' + x))

  await browser.close()
})()
