const { chromium } = require('playwright')
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const BASE = 'http://localhost:5199/'
const CHROME = 'C:/Program Files/Google/Chrome/Application/chrome.exe'
const UID = 1 // test_user_1

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

async function mineStatus(ctx, account) {
  const p = await ctx.newPage()
  await login(p, account)
  await p.goto(BASE + '#/pages/mine/index', { waitUntil: 'domcontentloaded' })
  await sleep(2800)
  const t = (await p.locator('body').innerText()).replace(/\n+/g, ' ')
  await p.close()
  return {
    status: (t.match(/(已通过审核|审核中|审核未通过|资料未完善)/) || ['?'])[0],
    hasReason: t.includes('未通过原因'),
  }
}

async function detail(ctx, account) {
  const p = await ctx.newPage()
  await login(p, account)
  await p.goto(BASE + `#/pages/admin/user-detail?id=${UID}`, { waitUntil: 'domcontentloaded' })
  await sleep(3200)
  const hasAudit = await p.locator('.audit').count()
  const tag = hasAudit ? await p.locator('.audit .atag').innerText().catch(() => '?') : '(无卡片)'
  const btns = hasAudit ? await p.locator('.audit .btn').allInnerTexts() : []
  return { p, hasAudit: !!hasAudit, tag, btns }
}

function fmt(r) {
  return `状态=${r.status}${r.hasReason ? ' [有原因]' : ''}`
}

;(async () => {
  const browser = await chromium.launch({ executablePath: CHROME, headless: true })
  const ctx = await browser.newContext({ viewport: { width: 414, height: 896 } })
  const results = []
  const check = (name, ok, extra) => {
    results.push({ name, ok })
    console.log(`  ${ok ? '✅' : '❌'} ${name}${extra ? ' — ' + extra : ''}`)
  }

  // ── 前置：恢复 uid=1 为「已通过」
  console.log('\n=== [0] 前置：管理端先把它审过 ===')
  let d = await detail(ctx, 'admin')
  if (d.hasAudit && d.btns.includes('通过')) {
    await d.p.locator('.audit .btn.pass').click()
    await sleep(1200)
    await d.p.locator('.uni-modal__btn--primary, .uni-modal__btn').last().click()
    await sleep(3000)
    d = await detail(ctx, 'admin')
  }
  console.log(`  当前标签=${d.tag} 按钮=${JSON.stringify(d.btns)}`)
  check('恢复到「已通过」', d.tag === '已通过', d.tag)
  await d.p.close()

  // ── 场景 A：误拒 → 管理员改判回来
  console.log('\n=== [A] 误拒后能否纠正（原 bug：卡片消失） ===')
  d = await detail(ctx, 'admin')
  await d.p.locator('.audit .btn').first().click() // 拒绝
  await sleep(1200)
  const items = d.p.locator('.uni-actionsheet__cell')
  console.log('  拒绝原因选项数 =', await items.count())
  await items.first().click()
  await sleep(3500)
  await d.p.close()

  let st = await mineStatus(ctx, 'test_user_1')
  console.log('  用户端看到：', fmt(st))
  check('用户端变为「审核未通过」', st.status === '审核未通过', st.status)

  console.log('  >>> 关键：管理端重新进入详情页')
  d = await detail(ctx, 'admin')
  console.log(`  标签=${d.tag} 按钮=${JSON.stringify(d.btns)}`)
  check('被拒后审核卡片仍在', d.hasAudit)
  check('标签显示「未通过」', d.tag === '未通过', d.tag)
  check('仍有「通过」按钮可纠正', d.btns.includes('通过'), JSON.stringify(d.btns))

  if (d.btns.includes('通过')) {
    await d.p.locator('.audit .btn.pass').click()
    await sleep(1200)
    await d.p.locator('.uni-modal__btn--primary, .uni-modal__btn').last().click()
    await sleep(3500)
    const tag2 = await d.p.locator('.audit .atag').innerText().catch(() => '?')
    check('改判后标签变「已通过」', tag2 === '已通过', tag2)
  }
  await d.p.close()

  st = await mineStatus(ctx, 'test_user_1')
  console.log('  用户端看到：', fmt(st))
  check('用户端恢复「已通过审核」', st.status === '已通过审核', st.status)

  // ── 场景 B：被拒用户自己改资料 → 自动重新排队
  console.log('\n=== [B] 被拒用户改资料后能否重新排队（原 bug：永久卡死） ===')
  d = await detail(ctx, 'admin')
  await d.p.locator('.audit .btn').first().click() // 拒绝
  await sleep(1200)
  await d.p.locator('.uni-actionsheet__cell').first().click()
  await sleep(3500)
  await d.p.close()
  st = await mineStatus(ctx, 'test_user_1')
  check('先被拒', st.status === '审核未通过', st.status)

  console.log('  >>> 用户端点「我的资料 → 编辑」，走完 4 步后保存')
  const up = await ctx.newPage()
  await login(up, 'test_user_1')
  await up.goto(BASE + '#/pages/profile/setup?mode=edit', { waitUntil: 'domcontentloaded' })
  await sleep(3500)
  let body = (await up.locator('body').innerText()).replace(/\n+/g, ' ')
  console.log('  起始 =', body.slice(0, 60))
  // 编辑模式数据已在，直接连点「下一步」到最后一步
  for (let i = 0; i < 3; i++) {
    const nx = up.locator('text=下一步').last()
    if (await nx.count()) {
      await nx.click()
      await sleep(1500)
    }
  }
  body = (await up.locator('body').innerText()).replace(/\n+/g, ' ')
  console.log('  末步 =', body.slice(-70))
  const saveBtn = up.locator('text=保存').last()
  console.log('  「保存」按钮数 =', await up.locator('text=保存').count())
  if (await up.locator('text=保存').count()) {
    await saveBtn.click()
    await sleep(4500)
  }
  await up.close()

  st = await mineStatus(ctx, 'test_user_1')
  console.log('  用户端看到：', fmt(st))
  check('改资料后变「审核中」', st.status === '审核中', st.status)

  // ── 收尾：管理员再通过，恢复干净状态
  console.log('\n=== [C] 收尾：管理员通过，恢复初始 ===')
  d = await detail(ctx, 'admin')
  if (d.btns.includes('通过')) {
    await d.p.locator('.audit .btn.pass').click()
    await sleep(1200)
    await d.p.locator('.uni-modal__btn--primary, .uni-modal__btn').last().click()
    await sleep(3200)
  }
  await d.p.close()
  st = await mineStatus(ctx, 'test_user_1')
  check('回到「已通过审核」', st.status === '已通过审核', st.status)

  console.log('\n================ 汇总 ================')
  const pass = results.filter((r) => r.ok).length
  console.log(`通过 ${pass}/${results.length}`)
  results.filter((r) => !r.ok).forEach((r) => console.log('  ❌ ' + r.name))

  await browser.close()
  process.exit(pass === results.length ? 0 : 1)
})()
