param(
    [string]$Version = $env:GEX_VERSION,
    [string]$InstallDir = $(if ($env:GEX_INSTALL_DIR) { $env:GEX_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\gex-dev-kit' }),
    [switch]$NoPath
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repo = 'Gex-Development/gex-dev-kit'
$tempDir = $null

function Step([int]$Number, [string]$Text) {
    if ($env:NO_COLOR -or [Console]::IsOutputRedirected) { Write-Host "[$Number/5] $Text" }
    else { Write-Host "[$Number/5]" -ForegroundColor DarkGray -NoNewline; Write-Host " $Text" }
}
function Success([string]$Text) {
    if ($env:NO_COLOR -or [Console]::IsOutputRedirected) { Write-Host "ok: $Text" }
    else { Write-Host '✓' -ForegroundColor Green -NoNewline; Write-Host " $Text" }
}
function Show-NextStep([string]$Command, [string]$Description) {
    if ($env:NO_COLOR -or [Console]::IsOutputRedirected) {
        Write-Host ('  {0,-21}{1}' -f $Command, $Description)
    } else {
        Write-Host ('  ' + [char]27 + '[1m' + ('{0,-21}' -f $Command) + [char]27 + '[0m' + $Description)
    }
}
function Invoke-Gh {
    $previousPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        & gh @args 2>$null
    } finally {
        $ErrorActionPreference = $previousPreference
    }
}
function Detect-Arch {
    $machine = $env:PROCESSOR_ARCHITEW6432
    if (-not $machine) { $machine = $env:PROCESSOR_ARCHITECTURE }
    if (-not $machine) { throw 'Arquitetura sem build. Suportados: darwin/arm64, darwin/amd64, linux/arm64, linux/amd64, windows/arm64, windows/amd64.' }
    switch ($machine.ToUpperInvariant()) {
        'ARM64' { return 'arm64' }
        'AMD64' { return 'amd64' }
        default { throw 'Arquitetura sem build. Suportados: darwin/arm64, darwin/amd64, linux/arm64, linux/amd64, windows/arm64, windows/amd64.' }
    }
}
function Resolve-Version {
    if ($Version) { return $Version }
    $json = Invoke-Gh release view --repo $repo --json tagName
    if ($LASTEXITCODE -ne 0) { throw 'Release indisponível. Peça acesso à organização Gex-Development.' }
    return ($json | ConvertFrom-Json).tagName
}
function Download-Asset([string]$Name) {
    Invoke-Gh release download $Version --repo $repo --pattern $Name --dir $tempDir | Out-Null
    if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath (Join-Path $tempDir $Name))) {
        throw 'Release incompleta. Confira a versão e peça acesso à organização Gex-Development.'
    }
}
function Verify-Checksum([string]$Name) {
    $matchesFound = @(
        foreach ($line in (Get-Content -LiteralPath (Join-Path $tempDir 'SHA256SUMS'))) {
            if ($line -match '^([a-fA-F0-9]{64})  \*?(.+)$' -and $Matches[2] -ceq $Name) { $Matches[1] }
        }
    )
    if ($matchesFound.Count -ne 1) { throw "SHA256SUMS não contém uma linha única para $Name. Instalação cancelada." }
    $actual = (Get-FileHash -LiteralPath (Join-Path $tempDir $Name) -Algorithm SHA256).Hash
    if ($actual -ine $matchesFound[0]) { throw 'SHA-256 diferente do SHA256SUMS. Instalação cancelada.' }
}
function Install-Binary([string]$Name) {
    $target = Join-Path $InstallDir 'gex-dev-kit.exe'
    $previous = ''
    if (Test-Path -LiteralPath $target) {
        try { $previous = (& $target --version 2>$null | Out-String).Trim() } catch { $previous = '' }
    }
    [System.IO.Directory]::CreateDirectory($InstallDir) | Out-Null
    $staged = Join-Path $InstallDir ('.gex-dev-kit.' + [guid]::NewGuid().ToString('N') + '.exe')
    try {
        Copy-Item -LiteralPath (Join-Path $tempDir $Name) -Destination $staged
        $observed = (& $staged --version | Out-String).Trim()
        if ($LASTEXITCODE -ne 0 -or $observed -cne $Version) { throw "Versão divergente: $observed (esperada: $Version). Instalação cancelada." }
        try {
            if (Test-Path -LiteralPath $target) { [System.IO.File]::Replace($staged, $target, $null) }
            else { [System.IO.File]::Move($staged, $target) }
        } catch [System.IO.IOException] {
            throw 'Não foi possível trocar o executável: feche o gex-dev-kit e rode de novo.'
        }
        $observed = (& $target --version | Out-String).Trim()
        if ($LASTEXITCODE -ne 0 -or $observed -cne $Version) { throw "Versão instalada divergente: $observed (esperada: $Version)." }
        if ($previous) { Write-Host "Versão: $previous → $observed" } else { Write-Host "Versão: $observed" }
    } finally {
        if (Test-Path -LiteralPath $staged) { Remove-Item -LiteralPath $staged -Force }
    }
}
function Update-Path {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    $parts = @($current -split ';' | Where-Object { $_ })
    if ($parts -inotcontains $InstallDir) {
        $newPath = (@($parts) + $InstallDir) -join ';'
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    }
    if (@($env:Path -split ';') -inotcontains $InstallDir) { $env:Path += ";$InstallDir" }
}

try {
    Write-Host 'Kit GEX · instalador'
    Step 1 'Conferindo GitHub CLI e acesso…'
    if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
        throw 'Instale gh com winget install --id GitHub.cli; depois rode gh auth login.'
    }
    Invoke-Gh auth status | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'GitHub CLI sem autenticação. Rode gh auth login.' }
    Step 2 'Identificando release e plataforma…'
    $Version = Resolve-Version
    if ($Version -cnotmatch '^v[0-9]+\.[0-9]+\.[0-9]+$') { throw 'Versão inválida. Use vX.Y.Z.' }
    $arch = Detect-Arch
    $asset = "gex-dev-kit_${Version}_windows_${arch}.exe"
    Success "$Version (windows/$arch)"
    Step 3 'Baixando binário e SHA256SUMS…'
    $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ('gex-dev-kit-' + [guid]::NewGuid().ToString('N'))
    [System.IO.Directory]::CreateDirectory($tempDir) | Out-Null
    Download-Asset $asset
    Download-Asset 'SHA256SUMS'
    Step 4 'Conferindo SHA-256…'
    Verify-Checksum $asset
    Success 'integridade confirmada'
    Step 5 'Instalando e verificando versão…'
    Install-Binary $asset
    $displayDir = $InstallDir
    if ($env:USERPROFILE -and $displayDir.StartsWith($env:USERPROFILE + '\', [System.StringComparison]::OrdinalIgnoreCase)) {
        $displayDir = '~' + $displayDir.Substring($env:USERPROFILE.Length)
    }
    Success "gex-dev-kit $Version instalado em $(Join-Path $displayDir 'gex-dev-kit.exe')"
    if (-not $NoPath) { Update-Path }
    Write-Host ''
    Write-Host 'Próximos passos'
    Show-NextStep 'gex-dev-kit' 'abre o assistente'
    Show-NextStep 'gex-dev-kit doctor' 'confere a instalação'
    if ($NoPath -and @($env:Path -split ';') -inotcontains $InstallDir) {
        Write-Host ''
        Write-Host 'PATH'
        Write-Host "  `$env:Path += ';$InstallDir'"
    }
} catch {
    if ($env:NO_COLOR -or [Console]::IsErrorRedirected) { [Console]::Error.WriteLine("erro: $($_.Exception.Message)") }
    else { [Console]::Error.WriteLine("✗ $($_.Exception.Message)") }
    if ($PSCommandPath) { exit 1 } else { $global:LASTEXITCODE = 1; return }
} finally {
    if ($tempDir -and (Test-Path -LiteralPath $tempDir)) { Remove-Item -LiteralPath $tempDir -Recurse -Force }
}
