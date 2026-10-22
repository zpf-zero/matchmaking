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
}

;(async () => {
  const browser = await chromium.launch({ executablePath: CHROME, headless: true })
  const ctx = await browser.newContext({ viewport: { width: 414, height: 896 } })

  // STEP 1: 用户端确认初始状态（已通过）
  console.log('=== [1] 用户端 test_user_1 初始状态 ===')
  let up = await ctx.newPage()
  await login(up, 'test_user_1')
  await up.goto(BASE + '#/pages/mine/index', { waitUntil: 'domcontentloaded' })
  await sleep(2800)
  let t = (await up.locator('body').innerText()).replace(/\n+/g, ' ')
  console.log('  我的页状态 =', (t.match(/(已通过审核|审核中|审核未通过|资料未完善)/) || ['?'])[0])
  await up.close()

  console.log('\n>>> 外部把 test_user_1 置为「待审核」')
  // 通过 node 执行 SQL 太重，这里用接口：先让 admin 拒绝它（置为未通过），再提交审核
  // 简化：直接用后端内部方式 —— 用 admin 拒绝 → 用户端应变「审核未通过」
  const ap = await ctx.newPage()
  await login(ap, 'admin')
  await ap.goto(BASE + '#/pages/admin/user-detail?id=1', { waitUntil: 'domcontentloaded' })
  await sleep(3200)
  const btns = await ap.locator('.actions .btn').allInnerTexts()
  console.log('  详情页按钮 =', JSON.stringify(btns))
  const hasAudit = await ap.locator('.audit').count()
  console.log('  审核卡片 =', hasAudit ? '有 ✓' : '无 ✗')

  // 点拒绝
  if (hasAudit) {
    await ap.locator('.audit .btn').first().click() // 拒绝
    await sleep(1200)
    // actionSheet 选「头像非真人照片」
    const items = ap.locator('.uni-actionsheet__cell')
    console.log('  拒绝原因选项数 =', await items.count())
    if (await items.count()) {
      await items.first().click()
      await sleep(3500)
      const tag = await ap.locator('.audit .atag').innerText().catch(() => '?')
      console.log('  拒绝后标签 =', tag)
      const toast = await ap.locator('.uni-toast').innerText().catch(() => '(无)')
      console.log('  toast =', toast)
    }
  }
  await ap.close()

  // STEP 3: 用户端重新登录，看状态是否变
  console.log('\n=== [3] 用户端再看状态（应为「审核未通过」）===')
  up = await ctx.newPage()
  await login(up, 'test_user_1')
  await up.goto(BASE + '#/pages/mine/index', { waitUntil: 'domcontentloaded' })
  await sleep(2800)
  t = (await up.locator('body').innerText()).replace(/\n+/g, ' ')
  console.log('  我的页状态 =', (t.match(/(已通过审核|审核中|审核未通过|资料未完善)/) || ['?'])[0])
  console.log('  有未通过提示 =', t.includes('未通过原因') ? '✓' : '✗')

  await up.goto(BASE + '#/', { waitUntil: 'domcontentloaded' })
  await sleep(3000)
  t = (await up.locator('body').innerText()).replace(/\n+/g, ' ')
  console.log('  推荐页提示 =', t.includes('资料未通过') ? '显示「资料未通过」✓' : '(无提示)')
  await up.close()

  // STEP 4: 管理端再通过
  console.log('\n=== [4] 管理端重新通过 ===')
  const ap2 = await ctx.newPage()
  await login(ap2, 'admin')
  await ap2.goto(BASE + '#/pages/admin/user-detail?id=1', { waitUntil: 'domcontentloaded' })
  await sleep(3200)
  await ap2.locator('.audit .btn.pass').click()
  await sleep(1200)
  await ap2.locator('.uni-modal__btn--primary, .uni-modal__btn').last().click()
  await sleep(3500)
  const tag2 = await ap2.locator('.audit .atag').innerText().catch(() => '?')
  console.log('  通过后标签 =', tag2)
  await ap2.close()

  // STEP 5: 用户端确认恢复
  console.log('\n=== [5] 用户端确认恢复「已通过审核」===')
  up = await ctx.newPage()
  await login(up, 'test_user_1')
  await up.goto(BASE + '#/pages/mine/index', { waitUntil: 'domcontentloaded' })
  await sleep(2800)
  t = (await up.locator('body').innerText()).replace(/\n+/g, ' ')
  console.log('  我的页状态 =', (t.match(/(已通过审核|审核中|审核未通过|资料未完善)/) || ['?'])[0])
  await up.close()

  await browser.close()
})()
