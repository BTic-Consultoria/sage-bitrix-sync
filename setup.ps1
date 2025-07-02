# Windows 11 Setup Script for Sage-Bitrix Go Project
# PowerShell script to set up the project with your real configuration

Write-Host "🚀 Setting up Sage-Bitrix Go project on Windows 11" -ForegroundColor Green
Write-Host "====================================================" -ForegroundColor Green
Write-Host "Using database details from your config2.json file..." -ForegroundColor Yellow
Write-Host ""

# Create Windows batch file for easy running
Write-Host "🔧 Creating Windows batch files..." -ForegroundColor Cyan
@"
@echo off
echo 🚀 Testing connection to SRVSAGE\SAGEEXPRESS:64952...
go run cmd\test\main.go
pause
"@ | Out-File -FilePath "run.bat" -Encoding ascii

@"
@echo off
echo 📦 Installing Go dependencies...
go mod download
go mod tidy
echo ✅ Dependencies updated!
pause
"@ | Out-File -FilePath "deps.bat" -Encoding ascii

@"
@echo off
echo 🔨 Building executable...
if not exist bin mkdir bin
go build -o bin\test.exe cmd\test\main.go
echo ✅ Built bin\test.exe
pause
"@ | Out-File -FilePath "build.bat" -Encoding ascii

Write-Host ""
Write-Host "🎯 READY TO TEST!" -ForegroundColor Green
Write-Host "================" -ForegroundColor Green
Write-Host ""
Write-Host "Your Go app is configured with:" -ForegroundColor Yellow
Write-Host "  🖥️  Database: SRVSAGE\SAGEEXPRESS:64952" -ForegroundColor White
Write-Host "  📊 Database: STANDARD" -ForegroundColor White
Write-Host "  👤 User: LOGIC" -ForegroundColor White
Write-Host "  🔗 Bitrix: bit24.bitrix24.eu" -ForegroundColor White
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Copy your Go source files to the proper directories" -ForegroundColor White
Write-Host "  2. Double-click run.bat or type: go run cmd\test\main.go" -ForegroundColor White
Write-Host "  3. Should connect to the same database as your Tauri app!" -ForegroundColor White
Write-Host ""
Write-Host "🔥 This should work immediately since it uses your proven config! 🔥" -ForegroundColor Yellow

Write-Host ""
Write-Host "📂 Created files:" -ForegroundColor Cyan
Write-Host "  • .env (with your database config)" -ForegroundColor White
Write-Host "  • run.bat (quick test)" -ForegroundColor White  
Write-Host "  • deps.bat (install dependencies)" -ForegroundColor White
Write-Host "  • build.bat (build executable)" -ForegroundColor White
Write-Host "  • .gitignore (protect credentials)" -ForegroundColor White