@echo off
setlocal
set REPO_URL=%~1
if "%REPO_URL%"=="" set REPO_URL=https://github.com/appoloncel283-debug/pulsenet.git
git init || exit /b 1
git branch -M main || exit /b 1
git add . || exit /b 1
git commit -m "Release PulseNet 2.0.0" || exit /b 1
git remote remove origin >nul 2>nul
git remote add origin %REPO_URL% || exit /b 1
git push -u origin main || exit /b 1
echo Published to %REPO_URL%
