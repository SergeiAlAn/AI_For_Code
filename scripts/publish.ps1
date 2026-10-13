<#
.SYNOPSIS
    Собирает образы Flask- и Go-версии и публикует их на Docker Hub.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User myname -Version 1.0.0

.EXAMPLE
    # Только Go-версия и только для x86-серверов
    powershell -ExecutionPolicy Bypass -File scripts\publish.ps1 -User myname -Version 1.0.1 -Only go -Platforms linux/amd64
#>
param(
    # Логин на Docker Hub (или префикс другого реестра, например ghcr.io/myname)
    [Parameter(Mandatory)][string]$User,
    # Версия в формате 1.2.3
    [Parameter(Mandatory)][ValidatePattern('^\d+\.\d+\.\d+$')][string]$Version,
    # Платформы серверов: linux/amd64 — обычные x86, linux/arm64 — ARM
    [string]$Platforms = "linux/amd64,linux/arm64",
    # Публиковать только одну версию
    [ValidateSet("all", "python", "go")][string]$Only = "all",
    # Не обновлять тег latest
    [switch]$NoLatest
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot

$images = @(
    @{ Key = "python"; Name = "users-api-python"; Context = $root },
    @{ Key = "go"; Name = "users-api-go"; Context = (Join-Path $root "go-server") }
) | Where-Object { $Only -eq "all" -or $_.Key -eq $Only }

foreach ($image in $images) {
    $repo = "$User/$($image.Name)"
    $tags = @("--tag", "${repo}:$Version")
    if (-not $NoLatest) { $tags += @("--tag", "${repo}:latest") }

    Write-Host "`n==> $repo`:$Version ($Platforms)" -ForegroundColor Cyan
    docker buildx build --platform $Platforms @tags --push $image.Context
    if ($LASTEXITCODE -ne 0) {
        throw "Не удалось собрать или отправить $repo. Выполнен ли 'docker login'?"
    }
}

Write-Host "`nГотово. Опубликованы образы:" -ForegroundColor Green
foreach ($image in $images) {
    Write-Host "  $User/$($image.Name):$Version"
}
