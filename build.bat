@echo off
setlocal enabledelayedexpansion

REM Variables
set BINARY_NAME=main
set SRC_DIR=.
set OUT_DIR=out
set FINGERPRINTS_DIR=.\global\fingerprints
set TEMPLATES_DIR=.\assets\html

REM Define OS and architectures
set "OS_LIST=linux windows freebsd openbsd"
set "ARCH_LIST=386 amd64 arm64"


REM Check the input argument
if "%1" == "clean" goto clean
if "%1" == "all" goto build_all

REM Find the index of the platform in the OS and ARCH lists
for %%O in (%OS_LIST%) do (
    for %%A in (%ARCH_LIST%) do (
        if "%1" == "%%O-%%A" (
            call :build_platform %%O %%A
            goto :eof
        )
    )
)
goto usage

REM Clean up the build
:clean
echo Cleaning up...
if exist "%OUT_DIR%" rd /S /Q "%OUT_DIR%"
goto :eof

REM Build for a specific platform
:build_platform
setlocal
set "OS=%1"
set "ARCH=%2"
set "EXT=.exe"

REM Ensure the output directory exists
if not exist "%OUT_DIR%" (
    mkdir "%OUT_DIR%"
)

REM Determine the file extension based on OS
if "%OS%" == "windows" set "EXT=.exe"
if "%OS%" == "darwin" set "EXT="
if "%OS%" == "linux" set "EXT="
if "%OS%" == "freebsd" set "EXT="
if "%OS%" == "openbsd" set "EXT="

set "OUT_FILE=%OUT_DIR%\%BINARY_NAME%.%OS%-%ARCH%%EXT%"

echo Building for %OS% %ARCH%...
set GOOS=%OS%
set GOARCH=%ARCH%
go build -o "%OUT_FILE%" "%SRC_DIR%"
if errorlevel 1 goto :error

REM Copy static files
xcopy /Y /E "%FINGERPRINTS_DIR%\*" "%OUT_DIR%\fingerprints\" >nul
xcopy /Y /E "%TEMPLATES_DIR%\*" "%OUT_DIR%\html\" >nul
if errorlevel 1 goto :error

endlocal
goto :eof

REM Build all platforms
:build_all
echo Building for all platforms...
for %%O in (%OS_LIST%) do (
    for %%A in (%ARCH_LIST%) do (
        call :build_platform %%O %%A
    )
)
goto :eof

REM Error handling
:error
echo An error occurred during the build process.
exit /b 1

REM Usage information
:usage
echo Usage: %0 ^<command^>
echo Commands:

REM Print commands dynamically
for %%O in (%OS_LIST%) do (
    for %%A in (%ARCH_LIST%) do (
        echo   %%O-%%A       Build for %%O %%A
    )
)

echo   all               Build for all platforms
echo   clean             Clean the output directory
exit /b 1

:eof
endlocal
