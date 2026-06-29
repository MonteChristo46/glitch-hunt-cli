$ErrorActionPreference = "Stop"

$VERSION="{{VERSION}}"

$ESC = [char]27

Write-Host "{{BANNER}}" -NoNewline
Write-Host " $ESC[38;2;200;200;200mCLI INSTALLER | v$VERSION$ESC[0m`n"

$Url = "https://github.com/MonteChristo46/glitch-hunt-cli/raw/main/build/huntcli-windows-amd64.exe"
$BinName = "huntcli.exe"
$InstallDir = "C:\ProgramData\huntcli"
$PathScope = "Machine"

$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
$IsAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $IsAdmin) {
    Write-Host "[SYSTEM] Not running as Administrator. Installing to user directory."
    $InstallDir = Join-Path $env:LOCALAPPDATA "huntcli"
    $PathScope = "User"
} else {
    Write-Host "[SYSTEM] Running as ADMINISTRATOR"
}

if (-not (Test-Path -Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

try {
    $Acl = Get-Acl $InstallDir
    $Ar = New-Object System.Security.AccessControl.FileSystemAccessRule("Users", "ReadAndExecute", "ContainerInherit,ObjectInherit", "None", "Allow")
    $Acl.SetAccessRule($Ar)
    Set-Acl $InstallDir $Acl
} catch {
    Write-Warning "[CONFIG] Could not set directory permissions."
}

$Target = Join-Path $InstallDir $BinName
Write-Host "[STATUS] Downloading huntcli..."
Invoke-WebRequest -Uri $Url -OutFile $Target

Unblock-File -Path $Target

$CurrentPath = [Environment]::GetEnvironmentVariable("Path", $PathScope)
if ($CurrentPath -notlike "*$InstallDir*") {
    Write-Host "[CONFIG] Adding $InstallDir to $PathScope PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", $PathScope)
    $env:Path += ";$InstallDir"
} else {
    Write-Host "[CONFIG] PATH already configured."
}

Write-Host "[STATUS] Running huntcli install..."
Write-Host "--------------------------------------------------"
& $Target install

Write-Host "--------------------------------------------------"
Write-Host "[SUCCESS] Installation complete. You can now use 'huntcli'."
Write-Host "[INFO] You may need to restart your terminal for PATH changes to take effect."
