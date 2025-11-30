@echo off
setlocal

set OUTPUT_BASE=build
set PROJECT_NAME=fntv-proxy

echo Cleaning up %OUTPUT_BASE% directory...
if exist %OUTPUT_BASE% rmdir /s /q %OUTPUT_BASE%
mkdir %OUTPUT_BASE%

call :build windows amd64
call :build windows arm64
call :build linux amd64
call :build linux arm64
call :build darwin amd64
call :build darwin arm64

echo Build complete. Artifacts are in %OUTPUT_BASE%/
goto :eof

:build
set OS=%1
set ARCH=%2

set FOLDER_ARCH=%ARCH%
if "%ARCH%"=="arm64" set FOLDER_ARCH=aarch64

set OUTPUT_DIR=%OUTPUT_BASE%\%OS%_%FOLDER_ARCH%
mkdir %OUTPUT_DIR%

set EXE_NAME=%PROJECT_NAME%
if "%OS%"=="windows" set EXE_NAME=%PROJECT_NAME%.exe

echo Building for %OS%/%ARCH% -^> %OUTPUT_DIR%...

set GOOS=%OS%
set GOARCH=%ARCH%
set CGO_ENABLED=0

go build -trimpath -ldflags "-s -w" -o "%OUTPUT_DIR%\%EXE_NAME%" .

if %errorlevel% neq 0 (
    echo Failed to build for %OS%/%ARCH%
    exit /b 1
)

goto :eof
