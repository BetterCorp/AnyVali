[CmdletBinding()]
param(
    [string]$OutputDir = "dist-release"
)

$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$OutputRoot = Join-Path $RepoRoot $OutputDir
$Version = $null

function Write-Step {
    param([string]$Message)
    Write-Host "[release] $Message" -ForegroundColor Cyan
}

function Ensure-Command {
    param([string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command not found: $Name"
    }
}

function Reset-OutputRoot {
    if (Test-Path $OutputRoot) {
        Remove-Item -Recurse -Force $OutputRoot
    }

    $paths = @(
        "npm",
        "pypi",
        "crates",
        "maven/java",
        "maven/kotlin",
        "nuget",
        "packagist",
        "rubygems",
        "go",
        "cpp",
        "metadata"
    )

    foreach ($path in $paths) {
        New-Item -ItemType Directory -Force -Path (Join-Path $OutputRoot $path) | Out-Null
    }
}

function Invoke-DockerStep {
    param(
        [string]$Name,
        [string]$Image,
        [string]$WorkDir,
        [string]$Script,
        [hashtable]$EnvVars = @{},
        [string]$EntryPoint = ""
    )

    Write-Step "Building $Name with $Image"

    $dockerArgs = @(
        "run",
        "--rm",
        "-v", "${RepoRoot}:/workspace",
        "-w", $WorkDir
    )

    foreach ($key in $EnvVars.Keys) {
        $dockerArgs += @("-e", "${key}=$($EnvVars[$key])")
    }

    if ($EntryPoint) {
        $dockerArgs += @("--entrypoint", $EntryPoint)
    }

    $dockerArgs += @($Image)

    if ($EntryPoint) {
        $dockerArgs += @("-lc", $Script)
    } else {
        $dockerArgs += @("sh", "-lc", $Script)
    }

    & docker @dockerArgs
    if ($LASTEXITCODE -ne 0) {
        throw "Docker build step failed: $Name"
    }
}

function Get-Version {
    $packageJson = Get-Content (Join-Path $RepoRoot "sdk/js/package.json") -Raw | ConvertFrom-Json
    return [string]$packageJson.version
}

function Write-Manifest {
    param([object[]]$Packages)

    $manifest = [ordered]@{
        version = $Version
        generatedAtUtc = (Get-Date).ToUniversalTime().ToString("o")
        outputDir = $OutputDir
        packages = $Packages
    }

    $manifest | ConvertTo-Json -Depth 8 | Set-Content -Path (Join-Path $OutputRoot "manifest.json")
}

function Write-ReleaseNotes {
    $notes = @"
# Release Artifacts

Version: $Version

This directory contains locally built release artifacts produced with Docker.

Registries with uploadable package artifacts:
- npm: upload the `.tgz` in `npm/`
- PyPI: upload the `.whl` and `.tar.gz` in `pypi/`
- crates.io: upload the `.crate` in `crates/`
- Maven Central (Java): build output is in `maven/java/`
- Maven Central (Kotlin): build output is in `maven/kotlin/`
- NuGet: upload the `.nupkg` in `nuget/`
- RubyGems: upload the `.gem` in `rubygems/`

Registries without a manual package upload artifact in this repo:
- Go: publish via git tag/module source; `go/` contains a source archive for review
- Packagist: publishes from the tagged repository; `packagist/` contains a source archive for review
- C++: no package registry is configured; `cpp/` contains a source archive for a GitHub release or manual distribution

Notes:
- Java/Kotlin Maven Central publishing usually also requires signing and Central-specific metadata. These builds produce the local artifacts only.
- See `manifest.json` for the exact files generated.
"@

    Set-Content -Path (Join-Path $OutputRoot "README.md") -Value $notes
}

Ensure-Command docker
Reset-OutputRoot
$Version = Get-Version

$packageResults = @()

function Invoke-PackageBuild {
    param(
        [string]$Id,
        [string]$Registry,
        [string]$PublishMode,
        [string]$ArtifactSubdir,
        [string]$Name,
        [string]$Image,
        [string]$WorkDir,
        [string]$Script,
        [string[]]$Notes = @(),
        [hashtable]$EnvVars = @{},
        [string]$EntryPoint = ""
    )

    $result = [ordered]@{
        id = $Id
        registry = $Registry
        publishMode = $PublishMode
        status = "built"
        artifacts = @()
    }

    if ($Notes.Count -gt 0) {
        $result.notes = @($Notes)
    }

    try {
        Invoke-DockerStep -Name $Name -Image $Image -WorkDir $WorkDir -Script $Script -EnvVars $EnvVars -EntryPoint $EntryPoint
        $artifactDir = Join-Path $OutputRoot $ArtifactSubdir
        if (Test-Path $artifactDir) {
            $result.artifacts = @(Get-ChildItem $artifactDir | ForEach-Object { "$ArtifactSubdir/$($_.Name)" })
        }
    }
    catch {
        $result.status = "failed"
        $result.error = $_.Exception.Message
    }

    $script:packageResults += $result
}

Invoke-PackageBuild -Id "js" -Registry "npm" -PublishMode "upload artifact" -ArtifactSubdir "npm" -Name "npm package" -Image "node:20-bookworm" -WorkDir "/workspace/sdk/js" -Script @'
set -eu
npm ci
npm run build
npm pack --pack-destination /workspace/dist-release/npm
'@

Invoke-PackageBuild -Id "python" -Registry "PyPI" -PublishMode "upload artifact" -ArtifactSubdir "pypi" -Name "PyPI package" -Image "python:3.12-bookworm" -WorkDir "/workspace/sdk/python" -Script @'
set -eu
python -m pip install --upgrade pip build
python -m build --outdir /workspace/dist-release/pypi
'@

Invoke-PackageBuild -Id "rust" -Registry "crates.io" -PublishMode "upload artifact" -ArtifactSubdir "crates" -Name "crates.io package" -Image "rust:1.77-bookworm" -WorkDir "/workspace/sdk/rust" -EnvVars @{ PATH = "/usr/local/cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" } -Script @'
set -eu
export PATH="/usr/local/cargo/bin:$PATH"
cargo package --allow-dirty --no-verify || true
test -f target/package/*.crate
cp target/package/*.crate /workspace/dist-release/crates/
'@

Invoke-PackageBuild -Id "java" -Registry "Maven Central" -PublishMode "upload artifact bundle" -ArtifactSubdir "maven/java" -Name "Java package" -Image "maven:3.9.9-eclipse-temurin-17" -WorkDir "/workspace/sdk/java" -Notes @(
    "Local artifacts only. Maven Central publishing still needs signing and Central-required metadata."
) -Script @'
set -eu
mvn -B -Dmaven.test.skip=true package
cp pom.xml /workspace/dist-release/maven/java/
cp target/*.jar /workspace/dist-release/maven/java/
'@

Invoke-PackageBuild -Id "csharp" -Registry "NuGet" -PublishMode "upload artifact" -ArtifactSubdir "nuget" -Name "NuGet package" -Image "mcr.microsoft.com/dotnet/sdk:8.0" -WorkDir "/workspace/sdk/csharp" -Script @'
set -eu
dotnet restore src/AnyVali/AnyVali.csproj
dotnet build src/AnyVali/AnyVali.csproj --configuration Release --no-restore
dotnet pack src/AnyVali/AnyVali.csproj --configuration Release --no-build -o /workspace/dist-release/nuget
'@

Invoke-PackageBuild -Id "ruby" -Registry "RubyGems" -PublishMode "upload artifact" -ArtifactSubdir "rubygems" -Name "RubyGems package" -Image "ruby:3.3-bookworm" -WorkDir "/workspace/sdk/ruby" -Script @'
set -eu
bundle install
gem build anyvali.gemspec
cp anyvali-*.gem /workspace/dist-release/rubygems/
'@

Invoke-PackageBuild -Id "kotlin" -Registry "Maven Central" -PublishMode "upload artifact bundle" -ArtifactSubdir "maven/kotlin" -Name "Kotlin package" -Image "gradle:8.7.0-jdk8" -WorkDir "/workspace/sdk/kotlin" -Notes @(
    "Local artifacts only. Kotlin Maven Central publishing still needs maven-publish/signing configuration."
) -Script @'
set -eu
gradle --no-daemon clean jar
cp build.gradle.kts /workspace/dist-release/maven/kotlin/
cp build/libs/*.jar /workspace/dist-release/maven/kotlin/
'@

Invoke-PackageBuild -Id "php" -Registry "Packagist" -PublishMode "tagged repository sync" -ArtifactSubdir "packagist" -Name "PHP source archive" -Image "composer:2" -WorkDir "/workspace/sdk/php" -Notes @(
    "Packagist does not use manual artifact uploads. This archive is for review only."
) -Script @'
set -eu
composer archive --format=zip --dir /workspace/dist-release/packagist --file anyvali-php
'@

Invoke-PackageBuild -Id "go" -Registry "Go module" -PublishMode "git tag" -ArtifactSubdir "go" -Name "Go source archive" -Image "alpine/git:2.47.2" -WorkDir "/workspace" -EntryPoint "sh" -Notes @(
    "Go modules are released from the repository and version tag, not by uploading an archive."
) -Script @"
set -eu
git config --global --add safe.directory /workspace
git archive --format=tar.gz --prefix=anyvali-go-v$Version/ HEAD:sdk/go > /workspace/dist-release/go/anyvali-go-v$Version.tar.gz
"@

Invoke-PackageBuild -Id "cpp" -Registry "none" -PublishMode "manual distribution" -ArtifactSubdir "cpp" -Name "C++ source archive" -Image "alpine/git:2.47.2" -WorkDir "/workspace" -EntryPoint "sh" -Notes @(
    "No C++ registry is configured yet. Use the archive for GitHub Releases or another distribution channel."
) -Script @"
set -eu
git config --global --add safe.directory /workspace
git archive --format=tar.gz --prefix=anyvali-cpp-v$Version/ HEAD:sdk/cpp > /workspace/dist-release/cpp/anyvali-cpp-v$Version.tar.gz
"@

Write-Manifest -Packages $packageResults
Write-ReleaseNotes
Write-Step "Release artifacts written to $OutputRoot"

if (($packageResults | Where-Object { $_.status -eq "failed" }).Count -gt 0) {
    throw "One or more package builds failed. See dist-release/manifest.json."
}
