. "$PSScriptRoot/common.ps1"
Push-Location $ProjectRoot
try {
    Invoke-Checked 'npm.cmd' @('--prefix', 'web', 'run', 'build')
    Invoke-Checked $GoExe @('build', '-o', 'portfolio-server.exe', './cmd/server')
    Write-Host 'Open http://127.0.0.1:8080 — Ctrl+C stops the demo.'
    Invoke-Checked './portfolio-server.exe' @()
} finally { Pop-Location }
