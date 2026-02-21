A:\Projects\port0\install.bat#L1-200
@echo off
setlocal enabledelayedexpansion

rem -------------------------
rem Basic install settings
rem -------------------------
set "INSTALL_DIR=%USERPROFILE%\bin"
set "VERSION=latest"

echo Installing port0 for Windows...
echo.

rem Create install dir if missing
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
    if errorlevel 1 (
        echo Error: Could not create install directory "%INSTALL_DIR%".
        echo Ensure you have permission to create that folder.
        exit /b 1
    )
)

set "DOWNLOAD_URL=https://github.com/blu3ph4ntom/port0/releases/latest/download/port0-windows-amd64.exe"

echo Downloading port0 from:
echo   %DOWNLOAD_URL%
echo.

rem Try curl first (quiet). If curl not available, fall back to PowerShell Invoke-WebRequest.
curl -fsSL "%DOWNLOAD_URL%" -o "%INSTALL_DIR%\port0.exe" 2>nul
if errorlevel 1 (
    echo curl failed or not present - trying PowerShell download...
    powershell -NoProfile -Command "try { Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%INSTALL_DIR%\port0.exe' -UseBasicParsing; exit 0 } catch { exit 1 }"
)

if errorlevel 1 (
    echo Error: Failed to download port0 to "%INSTALL_DIR%\port0.exe"
    echo Please visit: https://github.com/blu3ph4ntom/port0/releases to download manually.
    exit /b 1
)

echo.
echo port0 installed to: %INSTALL_DIR%\port0.exe
echo.

rem -------------------------
rem PATH / environment notes
rem -------------------------
echo PATH notes:
echo - To use %INSTALL_DIR%\port0.exe from any shell, add it to your PATH.
echo - To add now for the current session run:
echo     set PATH="%%PATH%%;%INSTALL_DIR%"
echo - To persistently add it for your user (recommended), run:
echo     setx PATH "%%PATH%%;%INSTALL_DIR%"
echo   After running setx you must open a new terminal window to see the change.
echo.

rem -------------------------
rem Post-install: system setup
rem -------------------------
echo IMPORTANT: system integration step
echo - port0 needs to configure local DNS and (on Windows) NRPT/firewall rules to map *.web and *.local to your machine and to allow binding to ports 80 and 53.
echo - Run the following in an elevated (Administrator) Command Prompt or PowerShell:
echo     "%INSTALL_DIR%\port0.exe" setup
echo   or, if you added port0 to PATH:
echo     port0 setup
echo.
echo - On Windows this step will attempt to add NRPT rules (for .web/.local) and open firewall ports required by port0.
echo - If you cannot run the setup step with admin rights, port0 will still work for the .localhost suffix (browsers resolve *.localhost to 127.0.0.1 by RFC 6761), but .web/.local may not function until you run setup as Administrator.
echo.

rem -------------------------
rem Subdomains / Monorepo guidance
rem -------------------------
echo Subdomains and monorepos:
echo - port0 uses a short project NAME (derived from each project's folder name by default).
echo   That NAME is what you use as the hostname prefix in your browser:
echo     NAME.localhost
echo     NAME.web
echo     NAME.local
echo - For multi-app repositories (monorepos), the recommended patterns are:
echo   1) cd into the subproject directory and run port0 there:
echo        cd path\to\my-monorepo\apps\api
echo        port0 run npm run dev
echo      The project NAME will be derived from the subfolder (for example: "api").
echo   2) If you prefer a custom alias (no process spawn) use the link escape hatch to register a name:
echo        port0 link my-alias
echo      This writes an alias into port0 state and lets you reserve or point a hostname without creating a per-project config file.
echo - Naming tips for subdomains / multi-services:
echo   * Use clear, short names for each service (api, web, admin, worker).
echo   * If you need multiple processes for the same folder, spawn them from different cwd basenames (or use aliases).
echo   * port0's routing looks up the left-most hostname label (the part before the first dot) to find a project entry.
echo     Because of that, choose names so the intended hostname prefix matches the registered project NAME.
echo.

rem -------------------------
rem Quick usage examples
rem -------------------------
echo Quick start examples:
echo   cd C:\path\to\your\project
echo   port0 run npm run dev
echo   (open in browser) http://<project-name>.localhost
echo.
echo   For a Python simple server:
echo   port0 run python -m http.server
echo   (open) http://<project-name>.localhost
echo.

echo If you run into issues, try running "%INSTALL_DIR%\port0.exe" --help for commands and flags.
echo.

endlocal
exit /b 0
