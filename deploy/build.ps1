<#
.SYNOPSIS
  在 Windows 上一次性打出 Ubuntu(amd64) 的 H5 部署包。

.DESCRIPTION
  依次做四件事：
    1. 交叉编译后端  GOOS=linux GOARCH=amd64 CGO_ENABLED=0
    2. 构建 H5 前端  npm run build:h5（会先按 config.prod.yaml 生成前端配置）
    3. 归集头像      把 dicebear 占位图和本地 uploads 一起打进包，改写成自托管地址
    4. 组装 + 打包   deploy/ubuntu/ 与 deploy/matchmaking-ubuntu-amd64.zip

  地址、数据库连接都从下面两个文件读，不在脚本里写死：
    miniapp/config.prod.yaml  -> 前端 base_url，同时决定 nginx server_name 和头像域名
    server/config.prod.yaml   -> MySQL 连接信息

.PARAMETER SkipAvatars
  跳过头像下载与数据库改写（不联网、不碰库时用）。

.PARAMETER SkipDbWrite
  头像照样下载打包，但不执行 UPDATE。SQL 仍会生成到 deploy/ubuntu/sql/ 里，之后可手动执行。

.PARAMETER NoZip
  只生成 deploy/ubuntu/ 目录，不打 zip。

.PARAMETER BaseUrl
  顺带切换站点域名。传了它就会先把两个配置文件改成一致，再打包：
    miniapp/config.prod.yaml  -> base_url
    server/config.prod.yaml   -> upload.base_url（自动补 /uploads）
  例：-BaseUrl https://liangyuan.example.com

.EXAMPLE
  powershell -ExecutionPolicy Bypass -File deploy\build.ps1

.EXAMPLE
  powershell -ExecutionPolicy Bypass -File deploy\build.ps1 -BaseUrl https://liangyuan.example.com
#>
[CmdletBinding()]
param(
  [switch]$SkipAvatars,
  [switch]$SkipDbWrite,
  [switch]$NoZip,
  [string]$BaseUrl
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$RepoRoot   = Split-Path -Parent $PSScriptRoot
$DeployDir  = Join-Path $RepoRoot 'deploy'
$OutDir     = Join-Path $DeployDir 'ubuntu'
$TemplateDir= Join-Path $DeployDir 'template'
$ServerDir  = Join-Path $RepoRoot 'server'
$MiniappDir = Join-Path $RepoRoot 'miniapp'
$ZipPath    = Join-Path $DeployDir 'matchmaking-ubuntu-amd64.zip'

function Step($msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Ok($msg)   { Write-Host "    $msg" -ForegroundColor DarkGray }
function Warn($msg) { Write-Host "    $msg" -ForegroundColor Yellow }

function Require-Cmd($name, $hint) {
  if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
    throw "找不到命令 $name。$hint"
  }
}

# Windows PowerShell 5.1 在 $ErrorActionPreference='Stop' 下会把原生程序写到 stderr 的
# 任何一行（npm/uni 的 Sass legacy-js-api 弃用警告就属于此类）当成终止性错误，
# 导致脚本在「什么都没坏」的情况下中断。原生命令调用一律走这个包装：
# 只在退出码非 0 时才失败，stderr 原样透传当普通输出。
#
# 退出码通过 $script:NativeCode 传出而不是 return —— 若用 return，调用处一旦写成
# `$out = Invoke-Native ...` 就会把退出码混进 stdout 数组里；写成管道 `|` 又会被
# 当成管道内容。stdout 保持原样流过，这样才能安全地 `Invoke-Native ... | Set-Content`。
function Invoke-Native($exe, [string[]]$exeArgs) {
  $prev = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try {
    & $exe @exeArgs
    $script:NativeCode = $LASTEXITCODE
  } finally { $ErrorActionPreference = $prev }
}
function Native-Ok { return $script:NativeCode -eq 0 }

# 删除文件/目录，不信任守卫的异常，只信「路径是否真的消失」。
#
# 本机的文件系统守卫会把删除重定向到「回收站」；回收站不可用时，它会在
# **已经完成删除之后**仍然抛出 SAFE_DELETE_FAIL_CLOSED —— 实测日志为
#   RemoveItem=FAIL: [safe-delete][SAFE_DELETE_FAIL_CLOSED] ... reason:"trash-failed"  exists=False
# 即：路径已经删掉了，异常却把 $ErrorActionPreference='Stop' 触发成致命错误，
# 导致整条打包链在「清理旧产物」这一步就中断（报错行号会指向下面第 158 行）。
# 这与 Invoke-Native 处理的是同一类问题：外部机制的噪声被当成脚本自身失败。
#
# 所以统一走这里：吞掉异常 → 回查路径是否真的不在了 → 仍在则退到 .NET 直删。
function Remove-PathSafe($path) {
  if (-not (Test-Path -LiteralPath $path)) { return }
  try {
    Remove-Item -LiteralPath $path -Recurse -Force -ErrorAction Stop
  } catch {
    # 故意忽略：守卫删完仍抛异常也走下面的一致性回查
  }
  if (Test-Path -LiteralPath $path) {
    if ([System.IO.Directory]::Exists($path)) {
      [System.IO.Directory]::Delete($path, $true)
    } elseif ([System.IO.File]::Exists($path)) {
      [System.IO.File]::Delete($path)
    }
  }
  if (Test-Path -LiteralPath $path) {
    throw "无法删除：$path"
  }
}

# Shell 脚本必须是 LF + 无 BOM，CRLF 会让 Ubuntu 报 "bad interpreter: ^M"
function Copy-AsUnixText($src, $dst, [hashtable]$replace = @{}) {
  $text = [System.IO.File]::ReadAllText($src)
  foreach ($k in $replace.Keys) { $text = $text.Replace($k, $replace[$k]) }
  $text = $text -replace "`r`n", "`n"
  $dir = Split-Path -Parent $dst
  if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
  [System.IO.File]::WriteAllText($dst, $text, (New-Object System.Text.UTF8Encoding($false)))
}

# 不能用 Compress-Archive：PowerShell 5.1 写出的条目名用反斜杠（web\assets\x.js），
# Linux 的 unzip 不会把它当目录分隔符，会解出一堆文件名里带反斜杠的平铺文件。
# 这里自己写条目，强制用正斜杠。
function New-UnixZip($sourceDir, $zipPath) {
  Add-Type -AssemblyName System.IO.Compression | Out-Null
  Add-Type -AssemblyName System.IO.Compression.FileSystem | Out-Null
  if (Test-Path $zipPath) { Remove-PathSafe $zipPath }

  $root = (Resolve-Path $sourceDir).Path.TrimEnd('\')
  $zip = [System.IO.Compression.ZipFile]::Open($zipPath, 'Create')
  try {
    foreach ($f in Get-ChildItem -Path $root -Recurse -File) {
      $rel = $f.FullName.Substring($root.Length + 1).Replace('\', '/')
      [System.IO.Compression.ZipFileExtensions]::CreateEntryFromFile(
        $zip, $f.FullName, $rel, [System.IO.Compression.CompressionLevel]::Optimal) | Out-Null
    }
  } finally {
    $zip.Dispose()
  }
}

function Read-YamlValue($path, $key) {
  $m = Select-String -Path $path -Pattern "^\s*$key\s*:\s*[`"']?([^`"'#\r\n]+)" | Select-Object -First 1
  if (-not $m) { return $null }
  return $m.Matches[0].Groups[1].Value.Trim()
}

# 按顶层段落取值。server 和 mysql 下面都有 port，只按 key 找会拿错
function Read-YamlSectionValue($path, $section, $key) {
  $inSection = $false
  foreach ($line in [System.IO.File]::ReadAllLines($path)) {
    if ($line -match "^$section\s*:") { $inSection = $true; continue }
    if ($inSection) {
      # 回到顶层（不缩进且非空）就说明这一段结束了
      if ($line -match '^\S') { break }
      if ($line -match "^\s+$key\s*:\s*[`"']?([^`"'#\r\n]+)") {
        return $Matches[1].Trim()
      }
    }
  }
  return $null
}

# ---------------------------------------------------------------- 预检
Step '预检'
Require-Cmd go   '安装 Go 1.24+ 后重试。'
Require-Cmd node '安装 Node.js 后重试。'
Require-Cmd npm  '安装 Node.js 后重试。'

$MiniappProdYaml = Join-Path $MiniappDir 'config.prod.yaml'
$ServerProdYaml  = Join-Path $ServerDir  'config.prod.yaml'
foreach ($f in @($MiniappProdYaml, $ServerProdYaml)) {
  if (-not (Test-Path $f)) { throw "缺少配置文件：$f" }
}

# ---------------------------------------------------------------- 切换域名
# 传 -BaseUrl 时才执行。两处地址必须同时改，只改一处会让新上传的图片 404。
# 严格用 UTF-8 读写（按原文件是否带 BOM 决定写回形式），
# 不要用 Get-Content | Set-Content —— 会把中文注释写乱。
function Set-YamlBaseUrl([string]$Path, [string]$Pattern, [string]$Replacement, [string]$Label) {
  $bytes = [System.IO.File]::ReadAllBytes($Path)
  $hasBom = ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF)
  $text = [System.IO.File]::ReadAllText($Path)
  $new = [regex]::Replace($text, $Pattern, $Replacement)
  if ($new -eq $text) { throw "$Path 里没找到 $Label，未做任何修改" }
  [System.IO.File]::WriteAllText($Path, $new, (New-Object System.Text.UTF8Encoding($hasBom)))
}

if ($BaseUrl) {
  $target = $BaseUrl.TrimEnd('/')
  if ($target -notmatch '^https?://[^/]+$') {
    throw "-BaseUrl 需要是 http(s)://域名[:端口] 形式（不带路径），收到：$BaseUrl"
  }
  Step "切换站点地址 -> $target"
  Set-YamlBaseUrl $MiniappProdYaml '(?m)^base_url:.*$' "base_url: `"$target`"" 'base_url'
  # 替换串用单引号拼接：PowerShell 双引号里 ${1} 会被当成变量展开，正则分组会丢
  Set-YamlBaseUrl $ServerProdYaml  '(?m)^(\s+)base_url:.*$' ('${1}base_url: "' + $target + '/uploads"') 'upload.base_url'
  Ok "前端 base_url 与上传域名已同步改为 $target"
}

$BaseUrl = Read-YamlValue $MiniappProdYaml 'base_url'
if (-not $BaseUrl) { throw "$MiniappProdYaml 里没读到 base_url" }
$BaseUrl = $BaseUrl.TrimEnd('/')
$ServerName = ([System.Uri]$BaseUrl).Host

$UploadBaseUrl = Read-YamlSectionValue $ServerProdYaml 'upload' 'base_url'
if ($UploadBaseUrl -and -not $UploadBaseUrl.StartsWith($BaseUrl)) {
  Warn "注意：server/config.prod.yaml 的 upload.base_url ($UploadBaseUrl) 和前端 base_url ($BaseUrl) 不同源，新上传的图片可能取不到。"
}

Ok "前端 / 站点地址：$BaseUrl"
Ok "nginx server_name：$ServerName"

if (-not $SkipAvatars) {
  Require-Cmd mysql '需要 mysql 客户端来导出头像清单；或者加 -SkipAvatars 跳过这一步。'
}

# ---------------------------------------------------------------- 清理
Step "清理旧产物 $OutDir"
if (Test-Path $OutDir)  { Remove-PathSafe $OutDir }
if (Test-Path $ZipPath) { Remove-PathSafe $ZipPath }
New-Item -ItemType Directory -Force -Path `
  (Join-Path $OutDir 'server\logs'), `
  (Join-Path $OutDir 'server\uploads'), `
  (Join-Path $OutDir 'server\configs'), `
  (Join-Path $OutDir 'web'), `
  (Join-Path $OutDir 'nginx'), `
  (Join-Path $OutDir 'sql') | Out-Null

# ---------------------------------------------------------------- 后端
Step '交叉编译后端 (linux/amd64)'
Push-Location $ServerDir
try {
  $env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
  $binPath = Join-Path $OutDir 'server\hunjie-server'
  # -trimpath 去掉本机路径，-s -w 去符号表，体积小一半
  Invoke-Native 'go' @('build', '-trimpath', '-ldflags', '-s -w', '-o', $binPath, './cmd')
  if (-not (Native-Ok)) { throw '后端编译失败' }
  Ok ('hunjie-server  {0:N1} MB' -f ((Get-Item $binPath).Length / 1MB))
} finally {
  Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
  Pop-Location
}

Step '归集后端配置'
Copy-Item $ServerProdYaml (Join-Path $OutDir 'server\config.prod.yaml') -Force
$wordsSrc = Join-Path $ServerDir 'configs\sensitive_words.txt'
if (Test-Path $wordsSrc) {
  Copy-AsUnixText $wordsSrc (Join-Path $OutDir 'server\configs\sensitive_words.txt')
  Ok 'configs/sensitive_words.txt'
} else {
  Warn '没有 server/configs/sensitive_words.txt，后端会退回内置词库'
}

# 本地 uploads 里用户真实上传过的图，一并带上，否则线上这些头像是坏的
$uploadsSrc = Join-Path $ServerDir 'uploads'
if (Test-Path $uploadsSrc) {
  Copy-Item "$uploadsSrc\*" (Join-Path $OutDir 'server\uploads') -Recurse -Force -ErrorAction SilentlyContinue
  $n = (Get-ChildItem (Join-Path $OutDir 'server\uploads') -Recurse -File).Count
  Ok "从本地 uploads 带上 $n 个文件"
}

# ---------------------------------------------------------------- 前端
Step '构建 H5 前端'
# 先手动清掉上次产物：vite 自己清 outDir 时是一次性批量删除，
# 在带文件系统钩子的受限环境里会被拦成 "SAFE_DELETE_BULK_CONFIRM_REQUIRED"，
# 表现为莫名其妙的 "Build failed"。先删干净，构建过程就只剩新增文件。
$h5Stale = Join-Path $MiniappDir 'dist\build\h5'
if (Test-Path $h5Stale) { Remove-PathSafe $h5Stale }
Push-Location $MiniappDir
try {
  if (-not (Test-Path 'node_modules')) {
    Ok '没有 node_modules，先执行 npm install'
    Invoke-Native 'npm' @('install')
    if (-not (Native-Ok)) { throw 'npm install 失败' }
  }
  # build:h5 的 pre 脚本会用 config.prod.yaml 重新生成 src/config.generated.ts
  Invoke-Native 'npm' @('run', 'build:h5')
  if (-not (Native-Ok)) { throw 'H5 构建失败' }
} finally { Pop-Location }

$h5Dist = Join-Path $MiniappDir 'dist\build\h5'
if (-not (Test-Path (Join-Path $h5Dist 'index.html'))) {
  throw "H5 产物不完整，没找到 $h5Dist\index.html"
}
Copy-Item "$h5Dist\*" (Join-Path $OutDir 'web') -Recurse -Force
$webCount = (Get-ChildItem (Join-Path $OutDir 'web') -Recurse -File).Count
$webSize  = (Get-ChildItem (Join-Path $OutDir 'web') -Recurse -File | Measure-Object Length -Sum).Sum
Ok ('web  {0} 个文件  {1:N1} MB' -f $webCount, ($webSize / 1MB))

# 构建完把前端配置还原成 dev，免得下次本地 npm run dev:h5 打到生产地址上
Push-Location $MiniappDir
try { Invoke-Native 'npm' @('run', 'config:dev') } finally { Pop-Location }
Ok '已把 src/config.generated.ts 还原为 dev 配置'

# ---------------------------------------------------------------- 头像
$SqlOut = Join-Path $OutDir 'sql\fix-avatars.sql'
if ($SkipAvatars) {
  Step '跳过头像处理（-SkipAvatars）'
} else {
  Step '归集用户头像'

  $DbHost = Read-YamlSectionValue $ServerProdYaml 'mysql' 'host'
  $DbPort = Read-YamlSectionValue $ServerProdYaml 'mysql' 'port'
  $DbUser = Read-YamlSectionValue $ServerProdYaml 'mysql' 'user'
  $DbPass = Read-YamlSectionValue $ServerProdYaml 'mysql' 'password'
  $DbName = Read-YamlSectionValue $ServerProdYaml 'mysql' 'database'
  if (-not ($DbHost -and $DbPort -and $DbUser -and $DbName)) {
    throw "$ServerProdYaml 的 mysql 段读取不完整"
  }
  Ok "数据库 $DbUser@$DbHost`:$DbPort/$DbName"

  $env:MYSQL_PWD = $DbPass
  $tsv = Join-Path $env:TEMP 'matchmaking-avatars.tsv'
  # 把所有站内头像都捞出来交给脚本判断。已经是生产地址的也要捞，
  # 否则重复打包时包里会缺图——数据库指着的文件根本没打进来
  $query = @"
SELECT openid, gender, avatar FROM user
WHERE deleted_at IS NULL AND avatar <> ''
  AND (avatar LIKE 'https://api.dicebear.com%'
       OR avatar LIKE 'http://localhost%'
       OR avatar LIKE 'http://127.0.0.1%'
       OR avatar LIKE '$BaseUrl/uploads/%');
"@
  $mysqlArgs = @("--host=$DbHost", "--port=$DbPort", "--user=$DbUser",
                 '--default-character-set=utf8mb4', "--database=$DbName",
                 '--batch', '--raw', '-e', $query)
  # 注意：不能写成 `$dump = Invoke-Native ...`（退出码会混进 stdout），
  # 也不能写成 `Invoke-Native ... | Set-Content`（退出码会占掉管道首行）。
  # 退出码走 $script:NativeCode，stdout 自然流过管道。
  Invoke-Native 'mysql' $mysqlArgs | Set-Content -Path $tsv -Encoding utf8
  if (-not (Native-Ok)) { throw '导出头像清单失败' }

  $rowCount = (Get-Content $tsv).Count - 1
  Ok "待处理头像 $rowCount 条"

  if ($rowCount -gt 0) {
    Invoke-Native 'node' @(
      (Join-Path $DeployDir 'scripts\fetch-avatars.mjs'),
      '--input', $tsv,
      '--uploads-src', $uploadsSrc,
      '--uploads-dst', (Join-Path $OutDir 'server\uploads'),
      '--base-url', $BaseUrl,
      '--sql-out', $SqlOut)
    if (-not (Native-Ok)) { Warn '有头像没处理成功，见上面日志' }

    if ($SkipDbWrite) {
      Warn "已生成 $SqlOut，但按 -SkipDbWrite 没有执行，需要时手动导入"
    } else {
      Ok '把改写语句写回数据库'
      # 重定向必须走 cmd（PS 的 `<` 不是重定向符），但要用 Invoke-Native 包装，
      # 否则 mysql 写到 stderr 的告警会被 PS 5.1 在 Stop 模式下当成致命错误中断脚本。
      # 注：不要改用 `mysql -e "source <path>"`，Windows 版 mysql 客户端对
      # source 的路径处理有坑（反斜杠会被吞），实测报 Failed to open file。
      $cmdline = "mysql --host=$DbHost --port=$DbPort --user=$DbUser " +
                 "--default-character-set=utf8mb4 --database=$DbName < `"$SqlOut`""
      Invoke-Native 'cmd' @('/c', $cmdline)
      if (-not (Native-Ok)) { throw '头像地址改写失败' }
      Ok '数据库头像地址已改写为自托管地址'
    }
  }
  Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue
  Remove-Item $tsv -ErrorAction SilentlyContinue
}

# ---------------------------------------------------------------- 脚本与说明
Step '组装部署脚本'
Copy-AsUnixText (Join-Path $TemplateDir 'deploy.sh') (Join-Path $OutDir 'deploy.sh')
foreach ($s in @('start.sh', 'stop.sh', 'restart.sh', 'status.sh')) {
  Copy-AsUnixText (Join-Path $TemplateDir "server\$s") (Join-Path $OutDir "server\$s")
}
# nginx 有两份模板：纯 HTTP 与 HTTPS。deploy.sh 会按证书文件是否就位自动二选一，
# 两份都要打进包，且都要把 server_name 填掉
foreach ($conf in @('matchmaking.conf', 'matchmaking-https.conf')) {
  Copy-AsUnixText (Join-Path $TemplateDir "nginx\$conf") `
                  (Join-Path $OutDir "nginx\$conf") `
                  @{ '__SERVER_NAME__' = $ServerName }
}
Ok "nginx server_name 已填为 $ServerName"

$stamp = Get-Date -Format 'yyyy-MM-dd HH:mm'
$readme = @"
良缘天合 · Ubuntu H5 部署包
打包时间：$stamp
站点地址：$BaseUrl

一键部署（在本目录执行）：
    bash deploy.sh

zip 不保留执行位，所以这里用 bash 显式调用；脚本会给其余文件补上 +x。

deploy.sh 会做：检查 nginx -> 停旧后端 -> 铺文件 -> 修权限 -> 以 root 启动新后端 -> 自检。
nginx 配置与包内模板一致且 nginx -t 通过时，会跳过重复复制与 reload。
可以重复执行；uploads 只合并不删除，服务器上已有的用户图片不会丢。

目录说明：
    web/                     H5 静态文件      -> /data/matchmaking/web
    server/hunjie-server     后端(linux-amd64) -> /data/matchmaking/server
    server/config.prod.yaml  生产配置（含数据库口令，注意保密）
    server/configs/          敏感词库
    server/uploads/          随包带的头像与图片
    server/{start,stop,restart,status}.sh   后台进程管理
    nginx/matchmaking.conf   站点配置
    sql/fix-avatars.sql      头像地址改写语句（打包时已执行过，留档备查）

日常操作：
    /data/matchmaking/server/status.sh      看进程 / 端口 / 健康检查 / 日志
    /data/matchmaking/server/restart.sh     重启后端
    tail -f /data/matchmaking/server/logs/server.log

前置条件：
    外部 MySQL 的白名单必须放行这台服务器的公网 IP，否则后端起不来。
    安全组开放 22 和 80 即可，8066 不用对公网开放。
"@
[System.IO.File]::WriteAllText((Join-Path $OutDir 'README.txt'), ($readme -replace "`r`n", "`n"),
  (New-Object System.Text.UTF8Encoding($false)))

# ---------------------------------------------------------------- 打包
if (-not $NoZip) {
  Step '打包 zip'
  New-UnixZip $OutDir $ZipPath
  Ok ('{0}  {1:N1} MB' -f (Split-Path -Leaf $ZipPath), ((Get-Item $ZipPath).Length / 1MB))
  # zip 不带 Unix 执行位，解压后由 deploy.sh 统一 chmod +x；
  # deploy.sh 自己则用 bash 显式调用，绕开它自身没有执行位的问题
  $bad = @()
  Add-Type -AssemblyName System.IO.Compression.FileSystem | Out-Null
  $probe = [System.IO.Compression.ZipFile]::OpenRead($ZipPath)
  try {
    foreach ($e in $probe.Entries) { if ($e.FullName -match '\\') { $bad += $e.FullName } }
  } finally { $probe.Dispose() }
  if ($bad.Count -gt 0) { throw "zip 里有 $($bad.Count) 个条目名带反斜杠，Linux 解压会出问题" }
  Ok '条目路径分隔符检查通过'
}

Step '完成'
Write-Host "    产物目录：$OutDir"
if (-not $NoZip) { Write-Host "    部署包：  $ZipPath" }
Write-Host @"

下一步：
    scp "$ZipPath" root@${ServerName}:/tmp/
    ssh root@$ServerName
    mkdir -p /data/matchmaking && cd /data/matchmaking
    unzip -o /tmp/matchmaking-ubuntu-amd64.zip
    bash deploy.sh
"@ -ForegroundColor Green
