# working_test.ps1
Write-Host "РАБОЧИЙ ТЕСТ API" -ForegroundColor Green

Write-Host "1. Health check..." -ForegroundColor Yellow
Invoke-WebRequest -Uri "http://localhost:8000/health" -Method Get

Write-Host "2. Registration..." -ForegroundColor Yellow  
Invoke-RestMethod -Uri "http://localhost:8000/api/auth/register" -Method Post -Headers @{"Content-Type"="application/json"} -Body '{"username":"testuser","email":"test@test.com","password":"123456"}'

Write-Host "3. Login..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "http://localhost:8000/api/auth/login" -Method Post -Headers @{"Content-Type"="application/json"} -Body '{"username":"testuser","password":"123456"}'

Write-Host "4. Stories..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "http://localhost:8000/api/stories/" -Method Get

Write-Host "ТЕСТ ЗАВЕРШЁН!" -ForegroundColor Green