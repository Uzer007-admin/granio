[CmdletBinding()]
param(
    [ValidateRange(10, 60)]
    [int]$Seconds = 20,

    [ValidatePattern('^[\p{L}\p{N}_.-]+$')]
    [string]$ProcessName = 'granio-app',

    [string]$OutputDirectory = '',

    [switch]$ValidateOnly
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) {
    throw 'Сбор GPU-трассы поддерживается только в Windows.'
}
if ([string]::IsNullOrWhiteSpace($OutputDirectory)) {
    $OutputDirectory = Join-Path (Split-Path -Parent $PSScriptRoot) ('graphics-trace-' + (Get-Date -Format 'yyyyMMdd-HHmmss'))
}

$recorder = (Get-Command wpr.exe -ErrorAction Stop).Source
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
try {
    $principal = [Security.Principal.WindowsPrincipal]::new($identity)
    $administrator = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
} finally {
    $identity.Dispose()
}

if ($ValidateOnly) {
    [pscustomobject]@{
        recorder = $recorder
        administrator = $administrator
        duration_seconds = $Seconds
        process_name = $ProcessName
        output_directory = [IO.Path]::GetFullPath($OutputDirectory)
        profiles = @('GPU', 'CPU.light')
        starts_application = $false
        changes_driver_settings = $false
    } | ConvertTo-Json -Depth 3
    return
}

if (-not $administrator) {
    throw 'Нужен PowerShell, открытый от имени администратора: WPR не может включить GPU/CPU-трассу без соответствующих системных привилегий. Скрипт не повышает права самостоятельно.'
}

$destination = [IO.Path]::GetFullPath($OutputDirectory)
if (Test-Path -LiteralPath $destination) {
    throw 'Каталог результата уже существует. Укажите новый OutputDirectory; файлы не будут перезаписаны.'
}
New-Item -ItemType Directory -Path $destination | Out-Null

$instance = 'GranioGraphics_' + [Guid]::NewGuid().ToString('N')
$traceFile = Join-Path $destination 'graphics.etl'
$started = $false
$samples = [Collections.Generic.List[object]]::new()
$record = [ordered]@{
    instance = $instance
    process_name = $ProcessName
    duration_seconds = $Seconds
    started_utc = [DateTime]::UtcNow.ToString('O')
    trace_saved = $false
    error = $null
}
$record | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $destination 'capture.json') -Encoding utf8

try {
    # Уникальное имя изолирует эту запись от других сеансов WPR.
    # Параметр instancename обязан быть последним в каждой команде WPR.
    $messages = & $recorder -start GPU -start CPU.light -filemode -recordtempto $destination -instancename $instance 2>&1
    $result = $LASTEXITCODE
    $messages | Out-File -LiteralPath (Join-Path $destination 'wpr-start.log') -Encoding utf8
    if ($result -ne 0) {
        throw "WPR не начал запись, код $result. Причина сохранена в wpr-start.log."
    }
    $started = $true

    Write-Host "Запись GPU/CPU началась на $Seconds секунд. Запустите проблемную Granio-сборку сейчас."
    Write-Host 'Приложение не закрывайте и не перемещайте окно во время проверки запуска.'
    $clock = [Diagnostics.Stopwatch]::StartNew()
    while ($clock.Elapsed.TotalSeconds -lt $Seconds) {
        $now = [DateTime]::UtcNow.ToString('O')
        foreach ($process in @(Get-Process -Name $ProcessName -ErrorAction SilentlyContinue)) {
            try {
                $process.Refresh()
                $samples.Add([pscustomobject]@{
                    utc = $now
                    elapsed_ms = $clock.Elapsed.TotalMilliseconds
                    process_id = $process.Id
                    cpu_seconds = $process.TotalProcessorTime.TotalSeconds
                    working_set_bytes = $process.WorkingSet64
                    private_bytes = $process.PrivateMemorySize64
                })
            } catch [InvalidOperationException] {
                # Процесс мог завершиться между перечислением и чтением счётчиков.
            } finally {
                $process.Dispose()
            }
        }
        Start-Sleep -Milliseconds 250
    }
} catch {
    $record.error = $_.Exception.Message
    throw
} finally {
    if ($started) {
        try {
            $messages = & $recorder -stop $traceFile -skipPdbGen -instancename $instance 2>&1
            $result = $LASTEXITCODE
        } catch {
            $messages = $_.Exception.Message
            $result = -1
            if (-not $record.error) {
                $record.error = $_.Exception.Message
            }
        }
        $messages | Out-File -LiteralPath (Join-Path $destination 'wpr-stop.log') -Encoding utf8
        $record.trace_saved = $result -eq 0 -and (Test-Path -LiteralPath $traceFile)
        if (-not $record.trace_saved) {
            Write-Warning "Не удалось сохранить трассу. Данные сеанса $instance оставлены для восстановления; другие сеансы WPR не затронуты."
        }
    }
    $record.ended_utc = [DateTime]::UtcNow.ToString('O')
    $record.process_seen = $samples.Count -gt 0
    $record.process_sample_count = $samples.Count
    ConvertTo-Json -InputObject @($samples.ToArray()) -Depth 3 | Set-Content -LiteralPath (Join-Path $destination 'process-samples.json') -Encoding utf8
    $record | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $destination 'capture.json') -Encoding utf8
}

if (-not $record.trace_saved) {
    throw 'Запись GPU-трассы не подтверждена. Проверьте wpr-stop.log.'
}
Write-Host "Трасса и счётчики сохранены: $destination"
