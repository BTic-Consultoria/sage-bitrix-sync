@echo off
echo ?? Installing Go dependencies...
go mod download
go mod tidy
echo ? Dependencies updated!
pause
