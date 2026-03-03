param(
  [string]$ApiBaseUrl = "http://localhost:18080"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$password = "Password123"

function New-UniqueEmail([string]$prefix) {
  $stamp = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
  $rand = Get-Random -Maximum 100000
  return "${prefix}_${stamp}_${rand}@lovr.test"
}

function Invoke-Api([string]$Method, [string]$Path, [object]$Body = $null, [hashtable]$Headers = @{}) {
  $uri = "$ApiBaseUrl$Path"
  if ($null -eq $Body) {
    return Invoke-RestMethod -Method $Method -Uri $uri -Headers $Headers
  }
  return Invoke-RestMethod -Method $Method -Uri $uri -Headers $Headers -ContentType "application/json" -Body ($Body | ConvertTo-Json -Depth 5)
}

try {
  Write-Output "[smoke-api] API=$ApiBaseUrl"
  $health = Invoke-RestMethod -Method Get -Uri "$ApiBaseUrl/health"
  if ($health.status -ne "ok") {
    throw "health check failed"
  }

  $email1 = New-UniqueEmail "smoke_api_a"
  $email2 = New-UniqueEmail "smoke_api_b"

  $auth1 = Invoke-Api "Post" "/auth/register" @{ email = $email1; password = $password }
  $auth2 = Invoke-Api "Post" "/auth/register" @{ email = $email2; password = $password }

  $login1 = Invoke-Api "Post" "/auth/login" @{ email = $email1; password = $password }
  if ($login1.user.id -ne $auth1.user.id) {
    throw "login user mismatch"
  }

  $headers1 = @{ Authorization = "Bearer $($auth1.token)" }
  $headers2 = @{ Authorization = "Bearer $($auth2.token)" }

  $discover1 = Invoke-Api "Get" "/discover" $null $headers1
  if (-not ($discover1.users | Where-Object { $_.id -eq $auth2.user.id })) {
    throw "discover should include second user before interactions"
  }

  $like12 = Invoke-Api "Post" "/likes" @{ toUserId = $auth2.user.id } $headers1
  if ($like12.matched -ne $false) {
    throw "first like should be unmatched"
  }

  $like21 = Invoke-Api "Post" "/likes" @{ toUserId = $auth1.user.id } $headers2
  if ($like21.matched -ne $true) {
    throw "second like should produce match"
  }

  $msg = Invoke-Api "Post" "/chats/$($auth2.user.id)/messages" @{ content = "hello from smoke-api" } $headers1
  if ([string]::IsNullOrWhiteSpace($msg.id)) {
    throw "message send failed"
  }

  $messages = Invoke-Api "Get" "/chats/$($auth2.user.id)/messages" $null $headers1
  if (-not $messages.messages -or $messages.messages.Count -lt 1) {
    throw "message list empty"
  }

  $block = Invoke-Api "Post" "/block" @{ toUserId = $auth2.user.id } $headers1
  if ($block.blocked -ne $true) {
    throw "block failed"
  }

  $likeAfterBlockStatus = 0
  try {
    Invoke-Api "Post" "/likes" @{ toUserId = $auth1.user.id } $headers2 | Out-Null
    $likeAfterBlockStatus = 200
  }
  catch {
    $likeAfterBlockStatus = $_.Exception.Response.StatusCode.value__
  }
  if ($likeAfterBlockStatus -ne 403) {
    throw "like after block should return 403, got $likeAfterBlockStatus"
  }

  $discoverAfter = Invoke-Api "Get" "/discover" $null $headers1
  if ($discoverAfter.users | Where-Object { $_.id -eq $auth2.user.id }) {
    throw "blocked user still visible in discover"
  }

  Write-Output "[smoke-api] PASS"
  exit 0
}
catch {
  Write-Output "[smoke-api] FAIL $($_.Exception.Message)"
  exit 1
}
