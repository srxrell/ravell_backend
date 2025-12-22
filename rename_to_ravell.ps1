
$baseUrl = "http://localhost:8080"
$rand = Get-Random -Minimum 1000 -Maximum 9999
$username = "testuser$rand"
$email = "test$rand@mail.com"
$password = "123456"

Write-Host "Testing against $baseUrl with user $username"

Write-Host "=== REGISTER ==="

$registerBody = @{
    username = $username
    email    = $email
    password = $password
} | ConvertTo-Json

$registerResponse = Invoke-RestMethod -Uri "$baseUrl/register" -Method POST -ContentType "application/json" -Body $registerBody -ErrorAction Stop

Write-Host "Registered OK"
Write-Host $registerResponse | Out-String

Start-Sleep -Seconds 1

Write-Host "`n=== LOGIN ==="

$loginBody = @{
    username = $username
    password = $password
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" -Method POST -ContentType "application/json" -Body $loginBody -ErrorAction Stop

Write-Host "Login OK"
Write-Host $loginResponse | Out-String

$token = $loginResponse.tokens.access_token

if (-not $token) {
    Write-Host "NO TOKEN…" -ForegroundColor Red
    exit 1
}

Write-Host "`nAccess Token:"
Write-Host $token

Start-Sleep -Seconds 1

Write-Host "`n=== PROFILE ==="

try {
    $profile = Invoke-RestMethod -Uri "$baseUrl/profile" -Headers @{ Authorization = "Bearer $token" } -Method GET -ErrorAction Stop
    Write-Host "Profile OK"
    Write-Host $profile | Out-String
}
catch {
    Write-Host "Profile request failed. Token?" -ForegroundColor Red
    Write-Host $_
    exit 1
}

Write-Host "`n=== DONE ==="