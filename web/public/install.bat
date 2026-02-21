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
echo   port0 npm run dev
echo   port0 python -m http.server
echo   port0 ls
echo.
echo Subdomain support (for monorepos):
echo   port0 -n api.myapp npm run dev     (creates api.myapp.localhost)
echo   port0 -n web.myapp npm run dev     (creates web.myapp.localhost)
echo   Or: port0 -n api --domain myapp npm run dev

endlocal
