# Fix localhost DNS resolution on Windows
# Run this script as Administrator

$hostsPath = "C:\Windows\System32\drivers\etc\hosts"

# Check if running as administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin) {
    Write-Host "This script requires Administrator privileges." -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator', then run this script again." -ForegroundColor Yellow
    exit 1
}

# Read hosts file
$content = Get-Content $hostsPath

# Check if localhost entries are commented
$needsUpdate = $false
$newContent = @()

foreach ($line in $content) {
    if ($line -match '^#\s+127\.0\.0\.1\s+localhost') {
        $newContent += "127.0.0.1       localhost"
        $needsUpdate = $true
    }
    elseif ($line -match '^#\s+::1\s+localhost') {
        $newContent += "::1             localhost"
        $needsUpdate = $true
    }
    else {
        $newContent += $line
    }
}

if ($needsUpdate) {
    # Backup original file
    $backupPath = "$hostsPath.backup.$(Get-Date -Format 'yyyyMMddHHmmss')"
    Copy-Item $hostsPath $backupPath
    Write-Host "Backup created: $backupPath" -ForegroundColor Green
    
    # Write updated content
    $newContent | Set-Content $hostsPath -Encoding ASCII
    Write-Host "Hosts file updated successfully!" -ForegroundColor Green
    
    # Flush DNS cache
    ipconfig /flushdns | Out-Null
    Write-Host "DNS cache flushed." -ForegroundColor Green
    
    Write-Host "`nPlease restart your browser and try http://localhost:3000" -ForegroundColor Cyan
}
else {
    Write-Host "Hosts file is already configured correctly." -ForegroundColor Yellow
}

