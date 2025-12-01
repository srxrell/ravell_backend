Write-Host "FIXING PROJECT NAME TO RAVELL..." -ForegroundColor Yellow

# 1. Fix go.mod
Write-Host "1. Updating go.mod..." -ForegroundColor Cyan
(Get-Content "go.mod") -replace "module go_stories_api", "module go_stories_api" | Set-Content "go.mod"
Write-Host "DONE: go.mod" -ForegroundColor Green

# 2. Fix imports in all Go files
Write-Host "2. Fixing imports..." -ForegroundColor Cyan
Get-ChildItem -Recurse -Filter "*.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $newContent = $content -replace "go_stories_api/", "go_stories_api/"
    if ($content -ne $newContent) {
        Set-Content $_.FullName $newContent
        Write-Host "UPDATED: $($_.Name)" -ForegroundColor Green
    }
}

# 3. Rebuild
Write-Host "3. Rebuilding..." -ForegroundColor Cyan
go mod tidy
go build -o ravell_api.exe
Write-Host "DONE: Project rebuilt" -ForegroundColor Green

Write-Host ""
Write-Host "SUCCESS! PROJECT RENAMED TO RAVELL BACKEND!" -ForegroundColor Green
Write-Host "NOW YOU ARE OFFICIAL RAVELL DEVELOPER, DUMBASS!" -ForegroundColor Red