@echo off
setlocal
set "SCRIPT=%~dp0PulseNet-Security.ps1"
if not exist "%SCRIPT%" (
  echo PulseNet-Security.ps1 was not found next to pnsec.cmd.
  exit /b 1
)
powershell.exe -NoProfile -ExecutionPolicy RemoteSigned -File "%SCRIPT%" %*
