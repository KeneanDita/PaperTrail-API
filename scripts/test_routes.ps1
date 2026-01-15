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
    $args += @('-H', "$k: $($Headers[$k])")
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
  $body = ($out -replace "`r?`nHTTP_STATUS:\d{3}`r?`n$", '')

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
$createdUser = Invoke-CurlJson -Method POST -Url "$ApiBase/users" -JsonBody ("{`"email`":`"$email`"}")

$userPublicId = $createdUser.Json.id
if (-not $userPublicId) {
  Write-Warning "Could not read created user id from response; some follow-up calls will be skipped."
} else {
  Invoke-CurlJson -Method GET -Url "$ApiBase/users/$userPublicId" | Out-Null
}

Invoke-CurlJson -Method GET -Url "$ApiBase/users" | Out-Null

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
Invoke-CurlJson -Method GET -Url "$ApiBase/papers" -Headers $authHeaders | Out-Null

# NOTE: The current codebase appears to use numeric IDs for papers/reviews/comments internally.
# These next calls assume paper id = 1 exists; they may 404/500 on a fresh DB.
Invoke-CurlJson -Method GET -Url "$ApiBase/papers/1" -Headers $authHeaders | Out-Null

# Reviews
Invoke-CurlJson -Method GET -Url "$ApiBase/papers/1/reviews" -Headers $authHeaders | Out-Null
Invoke-CurlJson -Method POST -Url "$ApiBase/papers/1/reviews" -Headers $authHeaders -JsonBody '{"reviewer_id":"1","rating":5,"comments":"Looks good"}' | Out-Null

# Comments
Invoke-CurlJson -Method GET -Url "$ApiBase/papers/1/comments" -Headers $authHeaders | Out-Null
Invoke-CurlJson -Method POST -Url "$ApiBase/papers/1/comments" -Headers $authHeaders -JsonBody '{"user_id":"1","body":"Nice paper"}' | Out-Null

Write-Host "\nDone."