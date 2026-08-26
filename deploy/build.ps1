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

.EXAMPLE
  powershell -ExecutionPolicy Bypass -File deploy\build.ps1
#>
[CmdletBinding()]
param(
  [switch]$SkipAvatars,
  [switch]$SkipDbWrite,
  [switch]$NoZip
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
  if (Test-Path $zipPath) { Remove-Item -Force $zipPath }

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
if (Test-Path $OutDir)  { Remove-Item -Recurse -Force $OutDir }
if (Test-Path $ZipPath) { Remove-Item -Force $ZipPath }
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
  & go build -trimpath -ldflags '-s -w' -o $binPath ./cmd
  if ($LASTEXITCODE -ne 0) { throw '后端编译失败' }
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
Push-Location $MiniappDir
try {
  if (-not (Test-Path 'node_modules')) {
    Ok '没有 node_modules，先执行 npm install'
    & npm install
    if ($LASTEXITCODE -ne 0) { throw 'npm install 失败' }
  }
  # build:h5 的 pre 脚本会用 config.prod.yaml 重新生成 src/config.generated.ts
  & npm run build:h5
  if ($LASTEXITCODE -ne 0) { throw 'H5 构建失败' }
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
try { & npm run config:dev | Out-Null } finally { Pop-Location }
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
  & mysql --host=$DbHost --port=$DbPort --user=$DbUser --default-character-set=utf8mb4 `
          --database=$DbName --batch --raw -e $query | Set-Content -Path $tsv -Encoding utf8
  if ($LASTEXITCODE -ne 0) { throw '导出头像清单失败' }

  $rowCount = (Get-Content $tsv).Count - 1
  Ok "待处理头像 $rowCount 条"

  if ($rowCount -gt 0) {
    & node (Join-Path $DeployDir 'scripts\fetch-avatars.mjs') `
        --input $tsv `
        --uploads-src $uploadsSrc `
        --uploads-dst (Join-Path $OutDir 'server\uploads') `
        --base-url $BaseUrl `
        --sql-out $SqlOut
    if ($LASTEXITCODE -ne 0) { Warn '有头像没处理成功，见上面日志' }

    if ($SkipDbWrite) {
      Warn "已生成 $SqlOut，但按 -SkipDbWrite 没有执行，需要时手动导入"
    } else {
      Ok '把改写语句写回数据库'
      cmd /c "mysql --host=$DbHost --port=$DbPort --user=$DbUser --default-character-set=utf8mb4 --database=$DbName < ""$SqlOut"""
      if ($LASTEXITCODE -ne 0) { throw '头像地址改写失败' }
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
Copy-AsUnixText (Join-Path $TemplateDir 'nginx\matchmaking.conf') `
                (Join-Path $OutDir 'nginx\matchmaking.conf') `
                @{ '__SERVER_NAME__' = $ServerName }
Ok "nginx server_name 已填为 $ServerName"

$stamp = Get-Date -Format 'yyyy-MM-dd HH:mm'
$readme = @"
缘来婚恋 · Ubuntu H5 部署包
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
    安全组开放 22 和 80 即可，8080 不用对公网开放。
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
