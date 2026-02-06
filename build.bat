@echo off
setlocal enableextensions

cd /d %~dp0

set "DIST_DIR=%CD%\dist"
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"

set "LDFLAGS=-s -w"

if /I "%SKIP_FRONTEND%"=="1" goto build_backend

where npm >nul 2>&1
if errorlevel 1 (
  echo npm not found, skip frontend build.
) else (
  if exist "frontend" (
    echo Building frontend...
    pushd "frontend"
    if not exist "node_modules" (
      call npm install
      if errorlevel 1 exit /b 1
    )
    call npm run build
    if errorlevel 1 exit /b 1
    popd
  )
)

:build_backend
echo Building linux/amd64...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
set GOARM=
go build -trimpath -ldflags "%LDFLAGS%" -o "%DIST_DIR%\\wyapi-golang_linux_amd64" .\\cmd\\server
if errorlevel 1 exit /b 1

echo Building linux/arm64 (armv8)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm64
set GOARM=
go build -trimpath -ldflags "%LDFLAGS%" -o "%DIST_DIR%\\wyapi-golang_linux_arm64" .\\cmd\\server
if errorlevel 1 exit /b 1

echo Building linux/arm (armv7)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -trimpath -ldflags "%LDFLAGS%" -o "%DIST_DIR%\\wyapi-golang_linux_armv7" .\\cmd\\server
if errorlevel 1 exit /b 1

echo Building windows/amd64...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
set GOARM=
go build -trimpath -ldflags "%LDFLAGS%" -o "%DIST_DIR%\\wyapi-golang_windows_amd64.exe" .\\cmd\\server
if errorlevel 1 exit /b 1

echo Done. Outputs in .\\dist\\
