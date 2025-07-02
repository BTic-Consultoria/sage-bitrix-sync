@echo off
echo ?? Building executable...
if not exist bin mkdir bin
go build -o bin\test.exe cmd\test\main.go
echo ? Built bin\test.exe
pause
