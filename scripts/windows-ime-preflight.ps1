# Read-only preflight; does not install capabilities or alter language profiles.
[CmdletBinding()]
param([string]$OutputPath = (Join-Path $PSScriptRoot '../docs/native-input-validation/windows-ime-capabilities.json'))
$ErrorActionPreference = 'Stop'
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]::new($identity)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw 'Open PowerShell as administrator to query online Windows language capabilities.'
}
$names = @('Language.Basic~~~ja-JP~0.0.1.0', 'Language.Basic~~~zh-CN~0.0.1.0', 'Language.Basic~~~ko-KR~0.0.1.0')
$capabilities = foreach ($name in $names) {
    $capability = Get-WindowsCapability -Online -Name $name
    [ordered]@{ Name = $capability.Name; State = [string]$capability.State }
}
$result = [ordered]@{
    OS = [Environment]::OSVersion.VersionString
    Capabilities = @($capabilities)
    LanguageProfiles = @(Get-WinUserLanguageList | ForEach-Object {
        [ordered]@{ LanguageTag = $_.LanguageTag; InputMethodTips = @($_.InputMethodTips) }
    })
}
$result | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $OutputPath -Encoding utf8
Write-Output ('Saved read-only IME preflight to ' + $OutputPath)
