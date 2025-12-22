# test_api_local.ps1
$ErrorActionPreference = "Stop"

# Настройки
$baseUrl = "http://localhost:8080"
$rand = Get-Random -Minimum 1000 -Maximum 9999
$username = "user$rand"
$email = "user$rand@example.com"
$password = "secret123"

Write-Host "🚀 STARTING API TEST" -ForegroundColor Cyan
Write-Host "Target: $baseUrl"
Write-Host "User: $username"
Write-Host "--------------------------------"

try {
    # 1. REGISTER
    Write-Host "1. Registering..." -NoNewline
    $registerBody = @{
        username = $username
        email    = $email
        password = $password
    } | ConvertTo-Json

    $regResponse = Invoke-RestMethod -Uri "$baseUrl/register" -Method POST -ContentType "application/json" -Body $registerBody
    Write-Host " OK" -ForegroundColor Green
    # Write-Host ($regResponse | Out-String)

    Start-Sleep -Seconds 1

    # 2. LOGIN
    Write-Host "2. Logging in..." -NoNewline
    $loginBody = @{
        username = $username
        password = $password
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method POST -ContentType "application/json" -Body $loginBody
    Write-Host " OK" -ForegroundColor Green
    
    $token = $loginResponse.tokens.access_token
    if (-not $token) {
        throw "No access token received!"
    }
    Write-Host "   Token received: $(($token).Substring(0, 15))..." -ForegroundColor Gray

    Start-Sleep -Seconds 1

    # 3. GET PROFILE (Protected Route)
    Write-Host "3. Fetching Profile (Protected)..." -NoNewline
    $headers = @{
        Authorization = "Bearer $token"
    }

    $profile = Invoke-RestMethod -Uri "$baseUrl/profile" -Method GET -Headers $headers
    Write-Host " OK" -ForegroundColor Green
    
    Write-Host "--------------------------------"
    Write-Host "✅ ALL TESTS PASSED SUCCESSFULLY!" -ForegroundColor Green
    Write-Host "User Profile Data:"
    Write-Host ($profile | Out-String)
}
catch {
    Write-Host "`n❌ TEST FAILED" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host "Details: " ($_.ErrorDetails | Out-String) -ForegroundColor Red
    }
    exit 1
}
