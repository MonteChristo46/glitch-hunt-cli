$ErrorActionPreference = "Stop"

$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
$IsAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if ($IsAdmin) {
    $InstallDir = "C:\ProgramData\huntcli"
    $PathScope = "Machine"
} else {
    $InstallDir = Join-Path $env:LOCALAPPDATA "huntcli"
    $PathScope = "User"
}

if (Test-Path $InstallDir) {
    $OldPath = [Environment]::GetEnvironmentVariable("Path", $PathScope)
    if ($OldPath -like "*$InstallDir*") {
        Write-Host "Removing $InstallDir from $PathScope PATH..."
        $NewPath = $OldPath.Replace(";$InstallDir", "").Replace("$InstallDir;", "").Replace($InstallDir, "")
        [Environment]::SetEnvironmentVariable("Path", $NewPath, $PathScope)
    }

    Write-Host "Removing installation directory: $InstallDir"
    Remove-Item -Path $InstallDir -Recurse -Force
}

Write-Host "[SUCCESS] huntcli uninstalled."
