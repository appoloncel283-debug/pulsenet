@echo off
setlocal
if "%VERSION%"=="" set VERSION=2.0.0
if not exist bin mkdir bin
go test ./... || exit /b 1
go vet ./... || exit /b 1
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.version=%VERSION%" -o bin\pulsenet-windows-amd64.exe .\cmd\pulsenet || exit /b 1
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags "-s -w -X main.version=%VERSION%" -o bin\pulsenet-linux-amd64 .\cmd\pulsenet || exit /b 1
echo Built PulseNet %VERSION% for Windows and Linux.
