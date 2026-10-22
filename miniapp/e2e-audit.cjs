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

;(async () => {
  const browser = await chromium.launch({ executablePath: CHROME, headless: true })
  const ctx = await browser.newContext({ viewport: { width: 414, height: 896 } })
  const page = await ctx.newPage()

  const errs = []
  page.on('pageerror', (e) => errs.push('[pageerror] ' + e.message))
  page.on('console', (m) => {
    if (m.type() === 'error') errs.push('[console.error] ' + m.text())
  })
  const api = []
  page.on('response', async (res) => {
    const u = res.url()
    if (u.includes('/api/v1/')) {
      let c = '?'
      try { c = (await res.json()).code } catch {}
      if (c !== 0) api.push(`❌ ${u.split('/api/v1')[1]} -> ${res.status()}/${c}`)
      else api.push(`✓ ${u.split('/api/v1')[1]}`)
    }
  })

  console.log('=== [管理端] admin 登录 ===')
  console.log('  URL =', await login(page, 'admin'))

  console.log('\n=== [管理端] 进入资料审核页 ===')
  // 点「资料审核」
  const auditEntry = page.locator('.q-item', { hasText: '资料审核' })
  console.log('  审核入口数 =', await auditEntry.count())
  if (await auditEntry.count()) {
    await auditEntry.first().click()
    await sleep(3000)
    console.log('  URL =', page.url())
    const head = await page.locator('.head').innerText().catch(() => '(无)')
    console.log('  列表标题 =', head)
    const cards = await page.locator('.card').count()
    console.log('  卡片数 =', cards)
    // 打印第一张卡内容
    if (cards > 0) {
      const t = (await page.locator('.card').first().innerText()).replace(/\n+/g, ' | ')
      console.log('  首张卡 =', t.slice(0, 160))
      const passBtn = page.locator('.card').first().locator('.btn.pass')
      const rejectBtn = page.locator('.card').first().locator('.btn.reject')
      console.log('  通过按钮 =', await passBtn.count(), ' 拒绝按钮 =', await rejectBtn.count())
    }
  }

  console.log('\n=== [管理端] 点击「通过」实测 ===')
  const btnCount = await page.locator('.card').first().locator('.btn.pass').count()
  if (btnCount > 0) {
    const before = await page.locator('.card').count()
    await page.locator('.card').first().locator('.btn.pass').click()
    await sleep(3500)
    const after = await page.locator('.card').count()
    const toast = await page.locator('.uni-toast, .uni-sample-toast').innerText().catch(() => '(无toast)')
    console.log(`  卡片数 ${before} -> ${after}（应 -1）`)
    console.log('  toast =', toast)
  } else {
    console.log('  ⚠️ 没有待审卡片可点')
  }

  console.log('\n=== 全端错误汇总 ===')
  console.log('  非 0 业务码:', api.filter((x) => x.startsWith('❌')).join(' | ') || '无')
  console.log('  JS 错误:', errs.length ? errs.slice(0, 6).join('\n    ') : '无')

  await browser.close()
})()
