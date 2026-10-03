. "$PSScriptRoot/common.ps1"
Push-Location $ProjectRoot
$Backend = $null
$Frontend = $null
try {
    Invoke-Checked $GoExe @('build', '-o', 'portfolio-server.exe', './cmd/server')
    $Backend = Start-Process (Join-Path $ProjectRoot 'portfolio-server.exe') -WorkingDirectory $ProjectRoot -WindowStyle Hidden -PassThru
    $NodeExe = (Get-Command node.exe).Source
    $Frontend = Start-Process $NodeExe -ArgumentList @('node_modules/vite/bin/vite.js', '--host', '127.0.0.1') -WorkingDirectory (Join-Path $ProjectRoot 'web') -WindowStyle Hidden -PassThru
    Write-Host 'Open http://127.0.0.1:5173 — Ctrl+C stops both development processes.'
    while (-not $Backend.HasExited -and -not $Frontend.HasExited) { Start-Sleep -Milliseconds 500 }
    if (($Backend.HasExited -and $Backend.ExitCode -ne 0) -or ($Frontend.HasExited -and $Frontend.ExitCode -ne 0)) { throw 'A development process exited unsuccessfully.' }
} finally {
    if ($Backend -and -not $Backend.HasExited) { Stop-Process -Id $Backend.Id }
    if ($Frontend -and -not $Frontend.HasExited) { Stop-Process -Id $Frontend.Id }
    Pop-Location
}
