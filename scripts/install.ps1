$ErrorActionPreference = "Stop"

$Repo = "monkeymonk/gdt"
$InstallDir = if ($env:GDT_HOME) { $env:GDT_HOME } else { Join-Path $env:LOCALAPPDATA "gdt" }
$BinDir = Join-Path $InstallDir "bin"

function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $release.tag_name -replace '^v', ''
}

function Get-Platform {
    return "windows"
}

function Get-Arch {
    if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
        return "amd64"
    }
    elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        return "arm64"
    }
    else {
        throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
    }
}

function Install-Gdt {
    $version = Get-LatestVersion
    $platform = Get-Platform
    $arch = Get-Arch

    Write-Host "Installing gdt $version for $platform/$arch..."

    $artifact = "gdt-$version-$platform-$arch.zip"
    $url = "https://github.com/$Repo/releases/download/v$version/$artifact"

    $tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        $tmpFile = Join-Path $tmpDir $artifact
        Write-Host "Downloading $url..."
        Invoke-WebRequest -Uri $url -OutFile $tmpFile

        New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
        Expand-Archive -Path $tmpFile -DestinationPath $BinDir -Force

        # Create shims directory
        $shimsDir = Join-Path $InstallDir "shims"
        New-Item -ItemType Directory -Path $shimsDir -Force | Out-Null

        # Copy gdt as godot shim
        Copy-Item (Join-Path $BinDir "gdt.exe") (Join-Path $shimsDir "godot.exe") -Force

        Write-Host ""
        Write-Host "gdt $version installed to $BinDir\gdt.exe"
        Write-Host ""
        Write-Host "Add gdt to your PATH:"
        Write-Host ""
        Write-Host "  [Environment]::SetEnvironmentVariable('Path', '$BinDir;$shimsDir;' + [Environment]::GetEnvironmentVariable('Path', 'User'), 'User')"
        Write-Host ""
    }
    finally {
        Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
    }
}

Install-Gdt
