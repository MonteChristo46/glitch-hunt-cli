$ErrorActionPreference = "Stop"

$VERSION="0.1.0-alpha"

$ESC = [char]27

Write-Host "\033[38;2;156;39;176m██╗  ██╗██╗   ██╗███╗   ██╗████████╗     ██████╗██╗     ██╗\033[0m
\033[38;2;125;61;168m██║  ██║██║   ██║████╗  ██║╚══██╔══╝    ██╔════╝██║     ██║\033[0m
\033[38;2;94;83;160m███████║██║   ██║██╔██╗ ██║   ██║       ██║     ██║     ██║\033[0m
\033[38;2;63;105;152m██╔══██║██║   ██║██║╚██╗██║   ██║       ██║     ██║     ██║\033[0m
\033[38;2;32;127;144m██║  ██║╚██████╔╝██║ ╚████║   ██║       ╚██████╗███████╗██║\033[0m
\033[38;2;0;150;136m╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝        ╚═════╝╚══════╝╚═╝\033[0m
" -NoNewline
Write-Host " $ESC[38;2;200;200;200mCLI INSTALLER | v$VERSION$ESC[0m`n"

$Url = "https://raw.githubusercontent.com/MonteChristo46/glitch-hunt-cli/main/dist/huntcli-windows-amd64.exe"
$BinName = "huntcli.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "huntcli"

New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

$Target = Join-Path $InstallDir $BinName
Write-Host "[STATUS] Downloading huntcli..."
Invoke-WebRequest -Uri $Url -OutFile $Target

Unblock-File -Path $Target

Write-Host ""
Write-Host "[OK] Installed to: $Target"
Write-Host ""

$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    Write-Host "Note: $InstallDir is not in your PATH."
    $choice = Read-Host "Add it to your PATH now? [Y/n]"
    if ($choice -notmatch "^(n|N|no|NO)$") {
        [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
        $env:Path += ";$InstallDir"
        Write-Host "[OK] Added to PATH."
        Write-Host "     Restart your terminal or log out/in for changes to take effect."
    } else {
        Write-Host ""
        Write-Host "To add it manually, run:"
        Write-Host "  setx PATH `"%PATH%;$InstallDir`""
    }
} else {
    Write-Host "[OK] PATH already configured."
}
Write-Host ""

Write-Host "Now run 'huntcli install' to complete setup:"
Write-Host "  $Target install"
Write-Host ""
Write-Host "Or authenticate directly:"
Write-Host "  $Target login"
Write-Host "  $Target listen --forward-to http://localhost:8080/webhooks"
