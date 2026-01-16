# PaperTrail API route smoke-test (curl-based)
# Usage examples:
#   powershell -ExecutionPolicy Bypass -File .\scripts\test_routes.ps1
#   $env:PAPERTRAIL_BASE_URL='http://localhost:8080'; .\scripts\test_routes.ps1
#   $env:JWT_SECRET='super-secret-jwt-key'; .\scripts\test_routes.ps1
#   $env:PAPERTRAIL_JWT='<paste token>'; .\scripts\test_routes.ps1
#   .\scripts\test_routes.ps1 -SkipPrivate

[CmdletBinding()]
param(
  [string]$BaseUrl = $(if ($env:PAPERTRAIL_BASE_URL) { $env:PAPERTRAIL_BASE_URL } else { 'http://localhost:8080' }),
  [string]$Jwt = $env:PAPERTRAIL_JWT,
  [string]$JwtSecret = $env:JWT_SECRET,
  [string]$Role = $(if ($env:PAPERTRAIL_ROLE) { $env:PAPERTRAIL_ROLE } else { 'admin' }),
  [switch]$SkipPrivate
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Assert-Status {
  param(
    [Parameter(Mandatory=$true)]$Response,
    [Parameter(Mandatory=$true)][int]$Expected,
    [string]$Context = ''
  )

  if ($Response.Status -ne $Expected) {
    $prefix = if ($Context) { "${Context}: " } else { '' }
    throw "${prefix}Expected HTTP $Expected, got $($Response.Status). Body: $($Response.Body)"
  }
}

function ConvertTo-Base64Url {
  param([Parameter(Mandatory=$true)][byte[]]$Bytes)

  $b64 = [Convert]::ToBase64String($Bytes)
  $b64 = $b64.TrimEnd('=')
  $b64 = $b64.Replace('+', '-').Replace('/', '_')
  return $b64
}

function New-JwtHS256 {
  param(
    [Parameter(Mandatory=$true)][string]$Secret,
    [Parameter(Mandatory=$true)][hashtable]$Payload
  )

  $headerJson = '{"alg":"HS256","typ":"JWT"}'
  $payloadJson = ($Payload | ConvertTo-Json -Compress)

  $headerB64 = ConvertTo-Base64Url ([Text.Encoding]::UTF8.GetBytes($headerJson))
  $payloadB64 = ConvertTo-Base64Url ([Text.Encoding]::UTF8.GetBytes($payloadJson))

  $toSign = "$headerB64.$payloadB64"
  $hmac = [System.Security.Cryptography.HMACSHA256]::new([Text.Encoding]::UTF8.GetBytes($Secret))
  try {
    $sigBytes = $hmac.ComputeHash([Text.Encoding]::UTF8.GetBytes($toSign))
  } finally {
    $hmac.Dispose()
  }

  $sigB64 = ConvertTo-Base64Url $sigBytes
  return "$toSign.$sigB64"
}

function Invoke-CurlJson {
  param(
    [Parameter(Mandatory=$true)][ValidateSet('GET','POST','PUT','PATCH','DELETE')][string]$Method,
    [Parameter(Mandatory=$true)][string]$Url,
    [string]$JsonBody,
    [hashtable]$Headers = @{}
  )

  Write-Host "\n==> $Method $Url"

  $args = @(
    '-sS',
    '-X', $Method,
    $Url,
    '-H', 'Accept: application/json'
  )

  foreach ($k in $Headers.Keys) {
    $args += @('-H', "${k}: $($Headers[$k])")
  }

  if ($null -ne $JsonBody -and $JsonBody.Trim().Length -gt 0) {
    $args += @(
      '-H', 'Content-Type: application/json',
      '--data', $JsonBody
    )
  }

  # Append a sentinel line with the HTTP status so we can parse it.
  $args += @('-w', "`nHTTP_STATUS:%{http_code}`n")

  $out = & curl.exe @args
  if ($LASTEXITCODE -ne 0) {
    throw "curl.exe failed (exit $LASTEXITCODE)"
  }

  $statusMatch = [regex]::Match($out, 'HTTP_STATUS:(\d{3})')
  $status = if ($statusMatch.Success) { [int]$statusMatch.Groups[1].Value } else { -1 }
  $body = [regex]::Replace($out, "`r?`nHTTP_STATUS:\d{3}.*\z", '')

  Write-Host "HTTP $status"
  if ($body) {
    Write-Host $body
  }

  $json = $null
  try {
    if ($body -and ($body.Trim().StartsWith('{') -or $body.Trim().StartsWith('['))) {
      $json = $body | ConvertFrom-Json
    }
  } catch {
    $json = $null
  }

  return [pscustomobject]@{
    Status = $status
    Body   = $body
    Json   = $json
  }
}

$BaseUrl = $BaseUrl.TrimEnd('/')
$ApiBase = "$BaseUrl/api"

# 1) Public health check
Invoke-CurlJson -Method GET -Url "$BaseUrl/health" | Out-Null

# 2) Public user bootstrapping routes
$email = "curltest+$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"

$createUserBody = @{ email = $email } | ConvertTo-Json -Compress
$createdUser = Invoke-CurlJson -Method POST -Url "$ApiBase/users" -JsonBody $createUserBody
Assert-Status -Response $createdUser -Expected 201 -Context 'Create user'

$userPublicId = $createdUser.Json.id
if (-not $userPublicId) {
  throw "Create user response did not include 'id'. Body: $($createdUser.Body)"
}

$fetchedUser = Invoke-CurlJson -Method GET -Url "$ApiBase/users/$userPublicId"
Assert-Status -Response $fetchedUser -Expected 200 -Context 'Get user'
if ($fetchedUser.Json.email -ne $email) {
  throw "Get user email mismatch. Expected '$email', got '$($fetchedUser.Json.email)'"
}

$userList = Invoke-CurlJson -Method GET -Url "$ApiBase/users"
Assert-Status -Response $userList -Expected 200 -Context 'List users'

$found = $false
try {
  foreach ($u in $userList.Json) {
    if ($u.id -eq $userPublicId -or $u.email -eq $email) {
      $found = $true
      break
    }
  }
} catch {
  $found = $false
}

if (-not $found) {
  Write-Warning "Created user was not found in list. This might be due to pagination/ordering differences."
}

if ($SkipPrivate) {
  Write-Host "\nSkipping authenticated /api/* routes (-SkipPrivate)."
  exit 0
}

# 3) Authenticated routes: use PAPERTRAIL_JWT if provided; otherwise generate from JWT_SECRET.
if (-not $Jwt -or $Jwt.Trim().Length -eq 0) {
  if (-not $JwtSecret -or $JwtSecret.Trim().Length -eq 0) {
    Write-Warning "No PAPERTRAIL_JWT and no JWT_SECRET found. Skipping authenticated routes."
    exit 0
  }

  if (-not $userPublicId) {
    Write-Warning "Cannot generate JWT without a user public id. Skipping authenticated routes."
    exit 0
  }

  $now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
  $Jwt = New-JwtHS256 -Secret $JwtSecret -Payload @{
    sub  = $userPublicId
    role = $Role
    iat  = $now
    exp  = $now + 3600
  }

  Write-Host "\nGenerated JWT (HS256) for sub=$userPublicId role=$Role"
}

$authHeaders = @{ Authorization = "Bearer $Jwt" }

# Papers

$papersList = Invoke-CurlJson -Method GET -Url "$ApiBase/papers" -Headers $authHeaders
if ($papersList.Status -eq 401) {
  throw "Authenticated routes are still protected. Provide PAPERTRAIL_JWT or JWT_SECRET, or keep -SkipPrivate. Body: $($papersList.Body)"
}

if ($papersList.Status -eq 200 -and $papersList.Json -and $papersList.Json.Count -gt 0) {
  $paperId = $papersList.Json[0].id
  if ($paperId) {
    Invoke-CurlJson -Method GET -Url "$ApiBase/papers/$paperId" -Headers $authHeaders | Out-Null
    Invoke-CurlJson -Method GET -Url "$ApiBase/papers/$paperId/reviews" -Headers $authHeaders | Out-Null
    Invoke-CurlJson -Method GET -Url "$ApiBase/papers/$paperId/comments" -Headers $authHeaders | Out-Null
  }
} else {
  Write-Host "\nNo papers found; skipping paper/review/comment detail checks."
}

Write-Host "\nDone."