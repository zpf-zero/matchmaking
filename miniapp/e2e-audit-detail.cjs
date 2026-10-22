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
  const api = []
  page.on('response', async (res) => {
    const u = res.url()
    if (u.includes('/api/v1/')) {
      let c = '?'
      try { c = (await res.json()).code } catch {}
      api.push(`${c === 0 ? '✓' : '❌'} ${u.split('/api/v1')[1]} ${c}`)
    }
  })

  console.log('=== 登录 admin ===')
  console.log('  ', await login(page, 'admin'))

  // 1) 用户管理 → 筛选待审核 → 进详情
  console.log('\n=== [1] 用户管理 → 待审核 → 详情 ===')
  await page.goto(BASE + '#/pages/admin/user', { waitUntil: 'domcontentloaded' })
  await sleep(3500)
  await page.locator('.fi', { hasText: '待审核' }).first().click()
  await sleep(2500)
  const cnt = await page.locator('.item').count()
  console.log('  待审用户数 =', cnt)
  if (cnt === 0) { console.log('  ⚠️ 无待审用户，跳过'); await browser.close(); return }

  const firstInfo = (await page.locator('.item').first().innerText()).replace(/\n+/g, ' | ')
  console.log('  首个 =', firstInfo.slice(0, 70))
  const uid = (firstInfo.match(/#(\d+)/) || [])[1]
  console.log('  目标 uid =', uid)

  await page.locator('.item').first().click()
  await sleep(3000)
  console.log('  URL =', page.url())

  // 2) 检查审核卡片是否出现
  console.log('\n=== [2] 详情页审核入口 ===')
  const auditCard = await page.locator('.audit').count()
  console.log('  审核卡片 =', auditCard ? '存在 ✓' : '缺失 ✗')
  if (auditCard) {
    const ct = await page.locator('.audit .ct').innerText()
    console.log('  卡片标题 =', ct.replace(/\n/g, ' '))
    const btns = await page.locator('.audit .btn').allInnerTexts()
    console.log('  审核按钮 =', JSON.stringify(btns))
  }
  const allBtns = await page.locator('.actions .btn').allInnerTexts()
  console.log('  全部操作按钮 =', JSON.stringify(allBtns))

  // 3) 点「通过」
  console.log('\n=== [3] 点击「通过」 ===')
  const passBtn = page.locator('.audit .btn.pass')
  if (await passBtn.count()) {
    await passBtn.click()
    await sleep(1200)
    // uni.showModal 确认
    const confirm = page.locator('.uni-modal__btn--primary, .uni-modal__btn').last()
    console.log('  确认弹窗 =', (await page.locator('body').innerText()).includes('通过后该用户') ? '✓' : '✗')
    await confirm.click()
    await sleep(3500)
    const toast = await page.locator('.uni-toast').innerText().catch(() => '(无)')
    console.log('  toast =', toast)
    const tag = await page.locator('.audit .atag').innerText().catch(() => '(无)')
    console.log('  审核标签变为 =', tag)
  } else {
    console.log('  ✗ 没找到「通过」按钮')
  }

  console.log('\n=== API 调用 ===')
  console.log(api.filter((x) => x.includes('audit') || x.includes('user/')).join('\n'))
  console.log('\n非0业务码:', api.filter((x) => x.startsWith('❌')).join(' | ') || '无')
  await browser.close()
})()
