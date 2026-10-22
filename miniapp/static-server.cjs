const http = require('http')
const fs = require('fs')
const path = require('path')
const root = process.argv[2] || process.cwd()
const port = Number(process.argv[3] || 5199)
const mime = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.svg': 'image/svg+xml',
  '.json': 'application/json',
  '.woff2': 'font/woff2',
  '.webp': 'image/webp',
}
http
  .createServer((req, res) => {
    let p = decodeURIComponent(req.url.split('?')[0])
    if (p === '/') p = '/index.html'
    let f = path.join(root, p)
    // SPA 兜底：找不到的路径一律回 index.html
    if (!fs.existsSync(f) || fs.statSync(f).isDirectory()) f = path.join(root, 'index.html')
    res.setHeader('Content-Type', mime[path.extname(f)] || 'application/octet-stream')
    res.setHeader('Cache-Control', 'no-store')
    fs.createReadStream(f).pipe(res)
  })
  .listen(port, () => console.log('static server on ' + port + ' root=' + root))
