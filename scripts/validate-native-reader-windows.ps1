[CmdletBinding()]
param([Parameter(Mandatory=$true)][string]$WorkDirectory)
$ErrorActionPreference = 'Stop'
$work = (Resolve-Path -LiteralPath $WorkDirectory).Path
if (Get-Process nvda -ErrorAction SilentlyContinue) { throw "Use a separate session or close an existing NVDA before this owned portable-reader test." }
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
$reader = $null
$probe = $null
try {
    $probe = Start-Process -FilePath (Join-Path $work 'input-probe.exe') -ArgumentList '-seconds 25' -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $work 'windows-probe-final.jsonl') -RedirectStandardError (Join-Path $work 'windows-probe-final-stderr.txt')
    $reader = Start-Process -FilePath (Join-Path $work 'nvda/nvda.exe') -ArgumentList @('--minimal','--disable-addons','--no-sr-flag','--debug-logging','--config-path',(Join-Path $work 'nvda-profile'),'--log-file',(Join-Path $work 'nvda-final.log')) -WindowStyle Hidden -PassThru
    $attempt = 0
    do {
        Start-Sleep -Milliseconds 200
        $probe.Refresh()
        $attempt++
    } while ($probe.MainWindowHandle -eq 0 -and $attempt -lt 50)
    if ($probe.MainWindowHandle -eq 0) { throw 'Owned Fyne probe has no window' }
    Start-Sleep -Seconds 3
    $root = [Windows.Automation.AutomationElement]::FromHandle($probe.MainWindowHandle)
    $queue = [Collections.Generic.Queue[Windows.Automation.AutomationElement]]::new()
    $queue.Enqueue($root)
    $walker = [Windows.Automation.TreeWalker]::RawViewWalker
    $password = $null
    $visited = 0
    while ($queue.Count -gt 0 -and $visited -lt 128) {
        $element = $queue.Dequeue()
        $visited++
        if ($element.Current.Name -eq 'Private input') { $password = $element; break }
        if ($element.Current.ControlType -in @([Windows.Automation.ControlType]::List, [Windows.Automation.ControlType]::Tree, [Windows.Automation.ControlType]::Table, [Windows.Automation.ControlType]::DataGrid)) { continue }
        $child = $walker.GetFirstChild($element)
        while ($child -and $queue.Count -lt 128) {
            $queue.Enqueue($child)
            $child = $walker.GetNextSibling($child)
        }
    }
    if ($null -eq $password -or -not $password.Current.IsPassword) { throw 'Missing protected native UIA field' }
    $password.SetFocus()
    $value = $password.GetCurrentPropertyValue([Windows.Automation.ValuePattern]::ValueProperty)
    if ($value -is [string] -and $value.Length -ne 0) { throw 'Native password value was exposed' }
    $pattern = $null
    if ($password.TryGetCurrentPattern([Windows.Automation.ValuePattern]::Pattern, [ref]$pattern) -and $pattern.Current.Value.Length -ne 0) { throw 'Native ValuePattern exposed the password' }
    $pattern = $null
    if ($password.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern, [ref]$pattern) -and $pattern.DocumentRange.GetText(-1).Length -ne 0) { throw 'Native TextPattern exposed the password' }
    Start-Sleep -Seconds 3
    [ordered]@{ protected = $true; password_value_redacted = $true; source_password_nonempty = $true } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $work 'windows-reader-client.json') -Encoding utf8
} finally {
    if ($probe -and -not $probe.HasExited) { Stop-Process -Id $probe.Id -ErrorAction SilentlyContinue }
    if ($reader -and -not $reader.HasExited) { Stop-Process -Id $reader.Id -ErrorAction SilentlyContinue }
}
$speech = Get-Content -LiteralPath (Join-Path $work 'nvda-final.log') | Where-Object { $_ -match '^Speaking ' -and $_ -match 'Fyne native input acceptance|CJK input|Private input' }
$speech | Set-Content -LiteralPath (Join-Path $work 'nvda-final-speech.txt') -Encoding utf8
if (-not ($speech | Where-Object { $_ -match 'Private input' -and $_ -match 'protected' })) { throw 'Real NVDA did not announce the protected field' }
Write-Output 'PASS real NVDA protected field and native UIA password redaction'
