/**
 * 把用户头像收拢成本站自托管的文件，让部署出去的站点不依赖外网图床，
 * 也不会残留指向开发机的地址。
 *
 * 对每个头像地址判断该怎么拿到图片文件，保证它一定进到部署包的 uploads 里：
 *   dicebear 外链      → 按原地址下载，画面和现在完全一致
 *   localhost 地址     → 本地文件正常就沿用，缺失或是几十字节的空图就补一张
 *   已经是生产地址     → 本地有文件就带走；没有但属于 avatar/seed/ 的可以按
 *                        openid 重新生成（同 seed 出同一张图），其余的报出来
 *
 * 最后一条是关键：脚本要能重复执行。第二遍跑的时候数据库里已经是生产地址了，
 * 如果这时不认这类地址，包里就会没有图，而数据库还指着它们。
 *
 * 用法：
 *   node fetch-avatars.mjs --input users.tsv --uploads-src <本地uploads> \
 *        --uploads-dst <包内uploads> --base-url http://1.2.3.4 --sql-out fix-avatars.sql
 *
 * 输入 TSV 由 build.ps1 从数据库导出，三列：openid、gender、avatar（含表头）。
 */
import fs from 'node:fs'
import path from 'node:path'

const args = {}
for (let i = 2; i < process.argv.length; i += 2) {
  args[process.argv[i].replace(/^--/, '')] = process.argv[i + 1]
}
for (const k of ['input', 'uploads-src', 'uploads-dst', 'base-url', 'sql-out']) {
  if (!args[k]) {
    console.error(`缺少参数 --${k}`)
    process.exit(1)
  }
}

const UPLOADS_SRC = args['uploads-src']
const UPLOADS_DST = args['uploads-dst']
const BASE_URL = args['base-url'].replace(/\/$/, '')

/** 生成的占位头像统一放这里，和用户真实上传的目录分开，也便于按 openid 重建 */
const SEED_DIR = 'avatar/seed'
/** 小于这个尺寸的多半是 1x1 占位图，当成没有头像 */
const MIN_VALID_BYTES = 512

const HAIR_MALE = 'shortFlat,shortRound,shortWaved,shortCurly,theCaesar,theCaesarAndSidePart,sides,frizzle'
const HAIR_FEMALE = 'longButNotTooLong,straight01,straight02,straightAndStrand,bob,bun,curly,curvy,bigHair'

function dicebearURL(seed, gender) {
  const p = new URLSearchParams({
    size: '200',
    seed,
    top: gender === 1 ? HAIR_MALE : HAIR_FEMALE,
    facialHairProbability: gender === 1 ? '35' : '0',
  })
  return `https://api.dicebear.com/9.x/avataaars/png?${p}`
}

const sqlStr = (s) => `'${String(s).replace(/\\/g, '\\\\').replace(/'/g, "''")}'`
const seedRel = (openid) => `${SEED_DIR}/${openid}.png`

function sizeOf(absPath) {
  try {
    return fs.statSync(absPath).size
  } catch {
    return -1
  }
}

function writeBoth(rel, buf) {
  for (const base of [UPLOADS_DST, UPLOADS_SRC]) {
    const abs = path.join(base, rel)
    fs.mkdirSync(path.dirname(abs), { recursive: true })
    fs.writeFileSync(abs, buf)
  }
}

/** 下载并同时落到包里和本地 uploads，本地留一份，下次打包直接复用 */
async function downloadTo(rel, url) {
  const res = await fetch(url, { redirect: 'follow' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const buf = Buffer.from(await res.arrayBuffer())
  if (buf.length < MIN_VALID_BYTES) throw new Error(`返回内容过小（${buf.length} 字节）`)
  writeBoth(rel, buf)
  return buf.length
}

/** 本地 uploads 里已有的文件，复制进包 */
function copyIntoBundle(rel) {
  const src = path.join(UPLOADS_SRC, rel)
  const dst = path.join(UPLOADS_DST, rel)
  fs.mkdirSync(path.dirname(dst), { recursive: true })
  fs.copyFileSync(src, dst)
}

/** 从 http://host[:port]/uploads/xxx 里取出 uploads 后面那段相对路径 */
function relFromUploadURL(url) {
  const i = url.indexOf('/uploads/')
  return i < 0 ? null : url.slice(i + '/uploads/'.length)
}

// ---------- 读入待处理用户 ----------
const rows = fs
  .readFileSync(args.input, 'utf8')
  .split(/\r?\n/)
  .map((l) => l.trim())
  .filter(Boolean)
  .slice(1) // 表头
  .map((line) => {
    const [openid, gender, avatar] = line.split('\t')
    return { openid, gender: Number(gender), avatar: (avatar || '').trim() }
  })
  .filter((r) => r.openid && r.avatar)

const updates = []
const stat = { downloaded: 0, reused: 0, rebuilt: 0, unchanged: 0, failed: 0 }
const unrecoverable = []

for (const r of rows) {
  const pad = r.openid.padEnd(14)
  let rel = null

  try {
    if (r.avatar.startsWith('https://api.dicebear.com')) {
      // 外链占位图：抓原图自托管
      rel = seedRel(r.openid)
      const n = await downloadTo(rel, r.avatar)
      console.log(`  下载  ${pad} ${n} 字节`)
      stat.downloaded++
    } else {
      const urlRel = relFromUploadURL(r.avatar)
      const isLocalHost = /^https?:\/\/(localhost|127\.0\.0\.1)/.test(r.avatar)
      const isProd = r.avatar.startsWith(`${BASE_URL}/uploads/`)

      if (!urlRel || (!isLocalHost && !isProd)) {
        console.log(`  跳过  ${pad} 无法识别的地址 ${r.avatar}`)
        continue
      }

      const localSize = sizeOf(path.join(UPLOADS_SRC, urlRel))
      if (localSize >= MIN_VALID_BYTES) {
        // 本地有正常文件，原样带走
        rel = urlRel
        copyIntoBundle(rel)
        console.log(`  沿用  ${pad} ${rel}（${localSize} 字节）`)
        stat.reused++
      } else if (urlRel.startsWith(`${SEED_DIR}/`)) {
        // 生成的占位头像，同一个 openid 重新生成得到的还是同一张图
        rel = urlRel
        const n = await downloadTo(rel, dicebearURL(r.openid, r.gender))
        console.log(`  重建  ${pad} ${rel}（${n} 字节）`)
        stat.rebuilt++
      } else if (isLocalHost) {
        // 开发机上传过但文件没了，或者只是个空图，补一张
        rel = seedRel(r.openid)
        const n = await downloadTo(rel, dicebearURL(r.openid, r.gender))
        const why = localSize < 0 ? '本地文件缺失' : `本地文件仅 ${localSize} 字节`
        console.log(`  补图  ${pad} ${why} -> ${n} 字节`)
        stat.downloaded++
      } else {
        // 生产地址 + 本地没有 + 不是我们生成的 → 这张图只在服务器上，包里带不走
        unrecoverable.push(`${r.openid}  ->  ${urlRel}`)
        continue
      }
    }
  } catch (e) {
    console.log(`  失败  ${pad} ${e.message}`)
    stat.failed++
    continue
  }

  const target = `${BASE_URL}/uploads/${rel}`
  if (target === r.avatar) {
    stat.unchanged++
  } else {
    updates.push(`UPDATE user SET avatar = ${sqlStr(target)} WHERE openid = ${sqlStr(r.openid)};`)
  }
}

const sql = [
  '-- 把头像地址改写成本站自托管地址，由 deploy/scripts/fetch-avatars.mjs 生成。',
  '-- 对应的图片文件已经打进部署包的 server/uploads/ 里。',
  `-- 生产地址：${BASE_URL}`,
  'SET NAMES utf8mb4;',
  '',
  ...(updates.length ? updates : ['-- 地址都已经是自托管形式，本次无需改写。']),
  '',
].join('\n')

fs.mkdirSync(path.dirname(args['sql-out']), { recursive: true })
fs.writeFileSync(args['sql-out'], sql, 'utf8')

console.log(
  `\n头像处理完成：下载 ${stat.downloaded}，重建 ${stat.rebuilt}，沿用本地 ${stat.reused}，` +
    `地址已正确 ${stat.unchanged}，失败 ${stat.failed}`,
)
console.log(`改写语句 ${updates.length} 条 -> ${args['sql-out']}`)

if (unrecoverable.length > 0) {
  console.log('\n以下头像只存在于服务器上，本地和包里都没有，线上若显示空头像需要单独处理：')
  unrecoverable.forEach((x) => console.log(`  ${x}`))
}
if (stat.failed > 0) process.exitCode = 2
