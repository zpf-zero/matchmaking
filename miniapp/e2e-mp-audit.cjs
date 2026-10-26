#!/usr/bin/env node
/**
 * 深度测试 1：审核闭环（小程序端管理后台）
 *
 * 流程：
 *   API 侧：造一个「待审核」用户（注册 → 资料 → 择偶要求 → 手机号 → 提交审核）
 *   模拟器：登录超管 → 管理中心 → 审核队列 → 通过该用户
 *   验证：数据库 user.status=2、user_profile.audit_status=1
 *
 * 前提：本地后端 localhost:8080 + 自动化端口 9420 已就绪
 */
const automator = require('miniprogram-automator')
const http = require('http')

const BASE = 'http://localhost:8080/api/v1'
const TS = Date.now().toString().slice(-8)
const NEW_OPENID = `qa_audit_${TS}`
const NICKNAME = `审核测试${TS.slice(-4)}`
const PHONE = '137' + TS

function api(method, path, token, body) {
  return new Promise((resolve, reject) => {
    const data = body ? JSON.stringify(body) : null
    const req = http.request(BASE + path, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    }, (res) => {
      let buf = ''
      res.on('data', (c) => (buf += c))
      res.on('end', () => {
        try { resolve(JSON.parse(buf)) } catch (e) { reject(e) }
      })
    })
    req.on('error', reject)
    if (data) req.write(data)
    req.end()
  })
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const results = []
function record(name, ok, detail = '') {
  results.push({ name, ok })
  console.log(`  [${ok ? 'PASS' : 'FAIL'}] ${name}` + (detail ? `  -> ${detail}` : ''))
}

async function main() {
  console.log('[1] API 造待审核用户')
  let r = await api('POST', '/auth/wx-login', null, { code: NEW_OPENID })
  const token = r.data.token
  const uid = r.data.user.id
  record('注册成功', r.code === 0, `uid=${uid}`)

  r = await api('PUT', '/user/profile', token, {
    nickname: NICKNAME, avatar: '/uploads/test.png', gender: 1,
    birthday: '1996-05-20', height: 176, weight: 68,
    marital_status: 1, has_children: 0, education: 3,
    school: '测试大学', occupation: '工程师', income_level: 3,
    house_status: 1, car_status: 1, province: '浙江省', city: '杭州市',
    city_code: '330100', home_province: '浙江省', home_city: '杭州市',
    self_intro: '这是一段用于审核闭环测试的自我介绍，长度超过三十个字符以满足校验要求。',
  })
  record('资料保存', r.code === 0, r.message)

  r = await api('PUT', '/user/requirement', token, {
    age_min: 22, age_max: 40, height_min: 155, height_max: 195,
    education_min: 0, income_min: 0, marital_status: '1',
    accept_children: 1, city_codes: '', other: '',
  })
  record('择偶要求保存', r.code === 0, r.message)

  r = await api('POST', '/auth/bind-phone', token, { phone: PHONE })
  record('绑定手机号', r.code === 0, r.message)

  r = await api('POST', '/user/submit-audit', token)
  record('提交审核', r.code === 0, r.message)

  console.log('\n[2] 模拟器：超管登录 → 审核通过')
  const mini = await automator.connect({ wsEndpoint: 'ws://127.0.0.1:9420' })
  const loginErrs = []
  mini.on('console', (m) => {
    if (m.type === 'error') loginErrs.push((m.args || []).map((a) => (typeof a === 'string' ? a : JSON.stringify(a))).join(' ').slice(0, 200))
  })
  try {
    await mini.callWxMethod('clearStorageSync').catch(() => {})
    await mini.reLaunch('/pages/login/index')
    await sleep(1500)
    let page = await mini.currentPage()

    // 预设第一项是 admin（超管）
    const presets = await page.$$('.preset')
    if (presets.length) await presets[0].tap()
    await sleep(300)
    ;(await page.$('.agreement')).tap()
    await sleep(300)
    ;(await page.$('.btn-primary')).tap()

    // 轮询等进入管理中心（登录 + reLaunch 耗时波动大）
    let ok = false
    for (let i = 0; i < 30; i++) {
      await sleep(800)
      page = await mini.currentPage()
      if (/admin/.test(page.path)) { ok = true; break }
    }
    record('超管登录进入管理中心', ok, ok ? page.path : `停在 ${page.path}，console: ${loginErrs[0] || '无'}`)
    if (!ok) throw new Error('登录未进入管理中心')
    await mini.screenshot({ path: 'e2e-mp-shots/d1-admin-home.png' })

    // 概览数据
    const cells = await page.$$('.num')
    record('概览数据渲染', cells.length >= 4, `cells=${cells.length}`)

    // 进审核队列
    await mini.navigateTo('/pages/admin/audit')
    await sleep(2500)
    page = await mini.currentPage()
    record('审核队列打开', /admin\/audit/.test(page.path), page.path)
    await mini.screenshot({ path: 'e2e-mp-shots/d2-audit-list.png' })

    // 找到目标用户并点「通过」
    const cards = await page.$$('.card')
    record('队列有待审项', cards.length > 0, `items=${cards.length}`)
    let passed = false
    for (const card of cards) {
      const text = await card.text().catch(() => '')
      if (text && text.includes(NICKNAME)) {
        const btn = await card.$('.btn.pass')
        if (btn) {
          await btn.tap()
          passed = true
        }
        break
      }
    }
    await sleep(2000)
    await mini.screenshot({ path: 'e2e-mp-shots/d3-after-pass.png' })
    record(`找到「${NICKNAME}」并点击通过`, passed)
  } finally {
    await mini.disconnect()
  }

  console.log('\n[3] 验证落库')
  // 口令从环境变量取，不进仓库：DB_PASSWORD=xxx node e2e-mp-audit.cjs
  const { execSync } = require('child_process')
  const dbPassword = process.env.DB_PASSWORD
  if (!dbPassword) {
    console.log('  跳过落库校验与清理：未设置 DB_PASSWORD 环境变量')
  } else {
    const sql = (q) => execSync(
      `mysql --host=101.35.245.54 --user=marry_link_dev --password=${dbPassword} marry_link_dev -N -B -e "${q}"`,
      { encoding: 'utf8' }
    ).trim()
    const st = sql(`SELECT CONCAT(u.status,'/',p.audit_status) FROM user u JOIN user_profile p ON p.user_id=u.id WHERE u.id=${uid};`)
    record('用户状态变为 正常/已通过(2/1)', st === '2/1', `实际=${st}`)

    // 清理：删掉测试用户与资料
    for (const t of ['user_profile', 'user_requirement', 'user_photo', 'three_view_answer']) {
      sql(`DELETE FROM ${t} WHERE user_id=${uid};`)
    }
    sql(`DELETE FROM interaction WHERE from_user_id=${uid} OR to_user_id=${uid};`)
    sql(`DELETE FROM invite_relation WHERE inviter_id=${uid} OR invitee_id=${uid};`)
    sql(`DELETE FROM user WHERE id=${uid};`)
    console.log(`已清理测试用户 uid=${uid}`)
  }

  const pass = results.filter((r) => r.ok).length
  console.log('\n' + '='.repeat(46))
  console.log(`审核闭环 ${pass} / ${results.length}`)
  console.log('='.repeat(46))
  process.exit(pass === results.length ? 0 : 1)
}

main().catch((e) => { console.error('脚本异常:', e.message); process.exit(1) })
