param(
    [ValidateSet('Build', 'Test', 'All')]
    [string]$Task = 'Build',
    [switch]$Race,
    [switch]$NativeGLFW
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

if ($env:OS -ne 'Windows_NT') {
    throw 'This script builds the Windows desktop applications.'
}

$null = Get-Command go -ErrorAction Stop
$null = Get-Command gcc -ErrorAction Stop
$projectRoot = [IO.Path]::GetFullPath((Split-Path -Parent $PSScriptRoot))
foreach ($modulePath in @('extras/apps', 'deps/gl', 'deps/glfw', 'deps/glow', 'deps/dbus')) {
    if (-not (Test-Path -LiteralPath (Join-Path $projectRoot "$modulePath/go.mod"))) {
        throw "$modulePath sources are missing. Run: git submodule update --init --recursive"
    }
}

$temporaryDirectory = Join-Path $projectRoot ('.build-' + [guid]::NewGuid().ToString('N'))
$savedEnvironment = @{}
foreach ($name in @('GOBIN', 'GOWORK', 'GOTMPDIR', 'CGO_ENABLED', 'CC')) {
    $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}

function Invoke-Go {
    param([string[]]$Arguments)
    & go @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "go $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

Push-Location $projectRoot
try {
    $null = New-Item -ItemType Directory -Path $temporaryDirectory
    $env:GOTMPDIR = $temporaryDirectory
    $env:GOWORK = Join-Path $projectRoot 'go.work'
    $env:CGO_ENABLED = '1'
    $env:CC = 'gcc'

    $localPatches = @(
        @('extras/apps', 'apps-fork.patch'),
        @('deps/gl', 'gl-fork.patch'),
        @('deps/glfw', 'glfw-fork.patch'),
        @('deps/glow', 'glow-fork.patch'),
        @('deps/dbus', 'dbus-fork.patch')
    )
    foreach ($localPatch in $localPatches) {
        $patchPath = Join-Path $projectRoot ('patches/' + $localPatch[1])
        if (-not (Test-Path -LiteralPath $patchPath)) { throw "Missing local patch: $patchPath" }
        $gitCommand = (Get-Command git -ErrorAction Stop).Source
        $moduleDirectory = Join-Path $projectRoot $localPatch[0]
        $patchCheckArguments = @('-C', ('"' + $moduleDirectory + '"'), 'apply', '--reverse', '--check', ('"' + $patchPath + '"'))
        $patchCheck = Start-Process -FilePath $gitCommand -ArgumentList $patchCheckArguments -WindowStyle Hidden -Wait -PassThru -RedirectStandardError (Join-Path $temporaryDirectory ($localPatch[1] + '.log'))
        if ($patchCheck.ExitCode -ne 0) {
            & git -C $moduleDirectory apply --check $patchPath
            if ($LASTEXITCODE -ne 0) { throw "Local patch conflicts with changes in $moduleDirectory. See $patchPath." }
            & git -C $moduleDirectory apply $patchPath
            if ($LASTEXITCODE -ne 0) { throw "Cannot apply $patchPath." }
        }
    }

    if ($Task -in @('Build', 'All')) {
        $env:GOBIN = Join-Path $projectRoot 'bin'
        $null = New-Item -ItemType Directory -Force -Path $env:GOBIN
        Invoke-Go -Arguments @('install', '-ldflags=-H=windowsgui', './cmd/granio_settings')
        Invoke-Go -Arguments @('install', '-ldflags=-H=windowsgui', './extras/apps')
        Write-Host "Applications installed in $env:GOBIN"
    }

    if ($Task -in @('Test', 'All')) {
        $testTags = 'no_glfw,migrated_fynedo'
        if ($NativeGLFW) { $testTags = 'migrated_fynedo' }
        $testArguments = @('test', ('-tags=' + $testTags))
        if ($Race) { $testArguments += '-race' }
        Invoke-Go -Arguments ($testArguments + @('./...', './extras/apps/...'))
    }
}
finally {
    foreach ($name in $savedEnvironment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process')
    }
    Pop-Location

    $resolvedTemporaryDirectory = [IO.Path]::GetFullPath($temporaryDirectory)
    $expectedPrefix = $projectRoot.TrimEnd([IO.Path]::DirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar + '.build-'
    if (-not $resolvedTemporaryDirectory.StartsWith($expectedPrefix, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Temporary directory is outside the project; cleanup refused.'
    }
    if (Test-Path -LiteralPath $resolvedTemporaryDirectory) {
        Remove-Item -LiteralPath $resolvedTemporaryDirectory -Recurse -Force
    }
}
