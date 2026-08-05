@echo off
setlocal
cd /d "%~dp0"
if not exist "bin\pulsenet-windows-amd64.exe" (
  echo PulseNet binary not found. Building it now...
  call build.bat || exit /b 1
)
"bin\pulsenet-windows-amd64.exe" %*
