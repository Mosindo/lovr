param(
  [string]$ApiBaseUrl = "http://localhost:18080"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$reportDir = Join-Path $repoRoot ".smoke"
$reportPath = Join-Path $reportDir "report.txt"

if (-not (Test-Path $reportDir)) {
  New-Item -ItemType Directory -Path $reportDir -Force | Out-Null
}

$results = New-Object System.Collections.Generic.List[string]
$failed = $false

function Run-Step([string]$name, [scriptblock]$action) {
  try {
    & $action
    $results.Add("PASS - $name")
  }
  catch {
    $results.Add("FAIL - $name - $($_.Exception.Message)")
    $script:failed = $true
  }
}

Push-Location $repoRoot
try {
  Run-Step "Smoke API automatise (checklist MVP courte)" {
    & powershell -ExecutionPolicy Bypass -File ".\scripts\smoke-api.ps1" -ApiBaseUrl $ApiBaseUrl
    if ($LASTEXITCODE -ne 0) { throw "smoke-api exited with code $LASTEXITCODE" }
  }

  Run-Step "Smoke mobile e2e (parcours critique)" {
    Push-Location ".\apps\mobile"
    try {
      npm run e2e:smoke | Out-Host
      if ($LASTEXITCODE -ne 0) { throw "npm run e2e:smoke exited with code $LASTEXITCODE" }
    }
    finally {
      Pop-Location
    }
  }
}
finally {
  Pop-Location
}

$stamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$header = @(
  "Lovr Smoke Report",
  "Generated: $stamp",
  "API Base URL: $ApiBaseUrl",
  ""
)

($header + $results) | Set-Content -Path $reportPath -Encoding UTF8
Get-Content $reportPath

if ($failed) {
  exit 1
}

exit 0
