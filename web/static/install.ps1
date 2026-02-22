# port0 Windows installer (PowerShell)
#
# One-liner:
#   irm https://port0.bluephantom.dev/install.ps1 | iex
#
# Optional env vars:
#   $env:PORT0_VERSION = "latest"   # or a tag, e.g. "v1.2.3"
#   $env:PORT0_INSTALL_DIR = "$env:USERPROFILE\bin"
#   $env:PORT0_ADD_TO_PATH = "1"    # set to "1" to persist PATH with setx
#   $env:PORT0_FORCE = "1"          # overwrite existing port0.exe
#
# Notes:
# - Requires PowerShell 5+ (Windows PowerShell) or PowerShell 7+.
# - Downloads from GitHub Releases.

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-Info([string]$Message) {
  Write-Host $Message
}

function Write-Warn([string]$Message) {
  Write-Host $Message -ForegroundColor Yellow
}

function Write-Ok([string]$Message) {
  Write-Host $Message -ForegroundColor Green
}

function Get-Arch {
  # Normalize to: amd64 | arm64
  # Prefer OS reported bitness/architecture over PROCESSOR_ARCHITECTURE quirks.
  try {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
    if ($arch -match "arm64") { return "arm64" }
    if ($arch -match "x64|amd64") { return "amd64" }
  } catch {
    # Fallback below
  }

  $envArch = ($env:PROCESSOR_ARCHITECTURE | ForEach-Object { $_.ToLowerInvariant() })
  if ($envArch -match "arm64") { return "arm64" }
  return "amd64"
}

function Get-DownloadUrl([string]$Version, [string]$Arch) {
  $base = "https://github.com/blu3ph4ntom/port0"
  $asset = "port0-windows-$Arch.exe"

  if ([string]::IsNullOrWhiteSpace($Version) -or $Version -eq "latest") {
    return "$base/releases/latest/download/$asset"
  }

  # Accept "1.2.3" or "v1.2.3" – do not mutate the tag.
  return "$base/releases/download/$Version/$asset"
}

function Ensure-Tls {
  # Ensure TLS 1.2+ for older Windows PowerShell.
  try {
    $tls12 = [Net.SecurityProtocolType]::Tls12
    if (-not ([Net.ServicePointManager]::SecurityProtocol.HasFlag($tls12))) {
      [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor $tls12
    }
  } catch {
    # Best-effort.
  }
}

function Add-ToPathSession([string]$Dir) {
  $path = [Environment]::GetEnvironmentVariable("PATH", "User")
  $parts = @()
  if ($path) { $parts = $path -split ";" }
  if ($parts -notcontains $Dir) {
    # Update current session immediately
    $env:PATH = "$env:PATH;$Dir"
  }
}

function Add-ToPathPersist([string]$Dir) {
  $path = [Environment]::GetEnvironmentVariable("PATH", "User")
  $parts = @()
  if ($path) { $parts = $path -split ";" }
  if ($parts -contains $Dir) { return }

  # Use setx to persist for future terminals (User scope).
  # NOTE: setx truncates very large PATH values on some Windows versions; user opted-in via env var.
  $newPath = if ([string]::IsNullOrEmpty($path)) { $Dir } else { "$path;$Dir" }
  & setx PATH "$newPath" | Out-Null
}

function Main {
  Ensure-Tls

  $version = $env:PORT0_VERSION
  if ([string]::IsNullOrWhiteSpace($version)) { $version = "latest" }

  $installDir = $env:PORT0_INSTALL_DIR
  if ([string]::IsNullOrWhiteSpace($installDir)) {
    $installDir = Join-Path $env:USERPROFILE "bin"
  }

  $force = $false
  if ($env:PORT0_FORCE -eq "1" -or $env:PORT0_FORCE -eq "true") { $force = $true }

  $addToPath = $false
  if ($env:PORT0_ADD_TO_PATH -eq "1" -or $env:PORT0_ADD_TO_PATH -eq "true") { $addToPath = $true }

  $arch = Get-Arch
  $url = Get-DownloadUrl -Version $version -Arch $arch

  $exePath = Join-Path $installDir "port0.exe"

  Write-Info "Installing port0 for Windows ($arch)..."
  Write-Info ""

  if (-not (Test-Path -LiteralPath $installDir)) {
    Write-Info "Creating install directory: $installDir"
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
  }

  if ((Test-Path -LiteralPath $exePath) -and (-not $force)) {
    Write-Warn "port0 is already installed at:"
    Write-Info "  $exePath"
    Write-Info ""
    Write-Info "Set `$env:PORT0_FORCE=1 and re-run to overwrite."
    return
  }

  $tempFile = Join-Path ([IO.Path]::GetTempPath()) ("port0-" + [Guid]::NewGuid().ToString("n") + ".exe")

  try {
    Write-Info "Downloading:"
    Write-Info "  $url"
    $ProgressPreference = "SilentlyContinue"

    # Prefer Invoke-WebRequest for binary downloads.
    Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing

    if (-not (Test-Path -LiteralPath $tempFile)) {
      throw "download failed: temp file not created"
    }

    $len = (Get-Item -LiteralPath $tempFile).Length
    if ($len -lt 1024 * 1024) {
      # Heuristic: a real binary should be > 1MB; small files are often HTML error pages.
      $head = ""
      try { $head = (Get-Content -LiteralPath $tempFile -Raw -TotalCount 1) } catch { }
      throw "download looks invalid (size ${len} bytes). Response may be an error page."
    }

    Move-Item -Force -LiteralPath $tempFile -Destination $exePath

    Write-Ok ""
    Write-Ok "✓ Installed:"
    Write-Info "  $exePath"

    # Update current session PATH so 'port0' works immediately in the same terminal.
    Add-ToPathSession -Dir $installDir

    if ($addToPath) {
      Write-Info ""
      Write-Info "Persisting PATH update (User scope)..."
      Add-ToPathPersist -Dir $installDir
      Write-Ok "✓ PATH updated (restart terminal to ensure it applies everywhere)"
    } else {
      Write-Info ""
      Write-Info "If 'port0' isn't found, add this folder to PATH:"
      Write-Info "  $installDir"
      Write-Info ""
      Write-Info "To persist it automatically, re-run with:"
      Write-Info "  `$env:PORT0_ADD_TO_PATH=1; irm https://port0.bluephantom.dev/install.ps1 | iex"
    }

    Write-Info ""
    Write-Info "Next:"
    Write-Info "  port0 daemon status"
    Write-Info "  port0 ls"
    Write-Info ""
    Write-Info "Optional system setup (Admin PowerShell):"
    Write-Info "  port0 setup"
  }
  finally {
    if (Test-Path -LiteralPath $tempFile) {
      Remove-Item -Force -LiteralPath $tempFile -ErrorAction SilentlyContinue | Out-Null
    }
  }
}

Main
