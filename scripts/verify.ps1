. "$PSScriptRoot/common.ps1"
Push-Location $ProjectRoot
try {
    Invoke-Checked $GoExe @('test', './...', '-count=1')
    Invoke-Checked $GoExe @('vet', './...')
    Invoke-Checked 'npm.cmd' @('--prefix', 'web', 'test')
    Invoke-Checked 'npm.cmd' @('--prefix', 'web', 'run', 'typecheck')
    Invoke-Checked 'npm.cmd' @('--prefix', 'web', 'run', 'build')
} finally { Pop-Location }
