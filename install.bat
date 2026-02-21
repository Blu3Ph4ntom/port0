@echo off
setlocal enabledelayedexpansion

set "INSTALL_DIR=%USERPROFILE%\bin"
set "VERSION=latest"

echo Installing port0 for Windows...

if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

set "DOWNLOAD_URL=https://github.com/blu3ph4ntom/port0/releases/latest/download/port0-windows-amd64.exe"

echo Downloading from %DOWNLOAD_URL%...
curl -fsSL "%DOWNLOAD_URL%" -o "%INSTALL_DIR%\port0.exe"

if errorlevel 1 (
    echo Error: Failed to download port0
    exit /b 1
)

echo.
echo port0 installed to %INSTALL_DIR%\port0.exe
echo.
echo Add to PATH if not already:
echo   setx PATH "%%PATH%%;%INSTALL_DIR%"
echo.
echo IMPORTANT: Run setup to configure DNS and permissions:
echo   port0 setup
echo.
echo Quick start:
echo   port0 run npm run dev
echo   port0 run python -m http.server
echo   port0 ls
echo.
echo Subdomain support (for monorepos):
echo   api.myapp.localhost ^-> routes to "api" project
echo   web.myapp.localhost ^-> routes to "web" project
echo   Run multiple projects under one parent domain.

endlocal
