# Usage (1-liner):
#   powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"

param(
  [string]$Owner = 'Hangell',
  [string]$Repo  = 'gommit',
  [switch]$PreRelease # usa o último pré-release
)

$ErrorActionPreference = 'Stop'
$ua = @{ 'User-Agent' = 'gommit-installer' }

try {
  Write-Host "Installing gommit..." -ForegroundColor Green

  # Detecta SO/arch
  $os = 'windows'
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLower()
  switch ($arch) {
    'x64'   { $arch = 'amd64' }
    'arm64' { $arch = 'arm64' }
    default { $arch = 'amd64' }
  }

  # Pega release
  $api = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
  if ($PreRelease) { $api = "https://api.github.com/repos/$Owner/$Repo/releases" }
  
  $resp = Invoke-RestMethod -UseBasicParsing -Headers $ua -Uri $api -TimeoutSec 30
  if ($PreRelease) { 
    $resp = ($resp | Where-Object { $_.prerelease -eq $true } | Select-Object -First 1)
    if (-not $resp) { throw "No pre-release found" }
  }

  # Escolhe asset .zip do OS/arch
  $pattern = "gommit_.*_${os}_${arch}\.zip$"
  $asset = $resp.assets | Where-Object { $_.name -match $pattern } | Select-Object -First 1
  if (-not $asset) { 
    throw "No matching asset for ${os}_${arch}. Available: $($resp.assets.name -join ', ')"
  }

  # Baixa e extrai
  $tmp = Join-Path $env:TEMP ("gommit-" + [guid]::NewGuid())
  New-Item -ItemType Directory -Force -Path $tmp | Out-Null
  $zip = Join-Path $tmp "gommit.zip"

  Write-Host "Downloading $($asset.name)..." -ForegroundColor Yellow
  Invoke-WebRequest -UseBasicParsing -Uri $asset.browser_download_url -OutFile $zip -TimeoutSec 60
  Expand-Archive -Path $zip -DestinationPath $tmp -Force

  # Procura o binário
  $exe = Get-ChildItem -Path $tmp -Recurse -Filter gommit.exe | Select-Object -First 1
  if (-not $exe) { 
    throw "gommit.exe not found in package"
  }

  # Instala
  Write-Host "Installing..." -ForegroundColor Yellow
  & $exe.FullName --install
  if ($LASTEXITCODE -ne 0) {
    throw "Installation failed with exit code: $LASTEXITCODE"
  }

  # Limpa
  Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue

  Write-Host ""
  Write-Host "✅ Installation completed successfully!" -ForegroundColor Green
  Write-Host "Close and reopen the terminal, then run: gommit --version" -ForegroundColor Cyan

} catch {
  Write-Host ""
  Write-Host "❌ Installation failed: $($_.Exception.Message)" -ForegroundColor Red
  Write-Host "Please check: https://github.com/$Owner/$Repo/releases" -ForegroundColor Yellow
  exit 1
}