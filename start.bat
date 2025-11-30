@echo off
echo 🚀 Starting Stories API...

:: Проверка Go
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go is not installed. Please install Go 1.21+
    pause
    exit /b 1
)

:: Установка зависимостей
echo 📦 Installing dependencies...
go mod tidy

:: Запуск сервера
echo 🚀 Starting server...
go run main.go
pause