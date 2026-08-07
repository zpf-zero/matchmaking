import { readFile, writeFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const env = process.argv[2] || 'dev'
const configPath = path.join(root, `config.${env}.yaml`)
const outputPath = path.join(root, 'src/config.generated.ts')
const yaml = await readFile(configPath, 'utf8')
const match = yaml.match(/^base_url:\s*(?:"([^"]*)"|'([^']*)'|([^#\r\n]*))/m)

if (!match) {
  throw new Error(`配置文件缺少 base_url: ${configPath}`)
}

const baseURL = (match[1] ?? match[2] ?? match[3] ?? '').trim().replace(/\/$/, '')
const output = `// 由 scripts/generate-config.mjs 自动生成，请勿手动修改。\nexport const BASE_URL = ${JSON.stringify(baseURL)}\n`
await writeFile(outputPath, output, 'utf8')
console.log(`已生成前端配置: ${env} -> ${baseURL || '(空地址)'}`)
