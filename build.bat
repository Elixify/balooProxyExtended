@echo off
setlocal

REM Variables
set BINARY_NAME=main
set SRC_DIR=.
set OUT_DIR=out
set FINGERPRINTS_DIR=.\global\fingerprints

REM Ensure the output directory exists
if not exist %OUT_DIR% (
    mkdir %OUT_DIR%
)

REM Check the input argument
if "%1" == "build" goto build_all
if "%1" == "build-all" goto build_all
if "%1" == "clean" goto clean
goto :usage

REM Clean up the build
:clean
echo Cleaning up...
if exist %OUT_DIR% rd /S /Q %OUT_DIR%
goto :eof

REM Cross-compile for Windows, Linux, and macOS
:build_all
echo Cross-compiling for all OSes...

REM Windows
echo Building for Windows...

set GOOS=windows
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.windows-amd64.exe %SRC_DIR%

set GOOS=windows
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.windows-arm64.exe %SRC_DIR%
if errorlevel 1 goto :error

REM Linux
echo Building for Linux...

set GOOS=linux
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.linux-amd64 %SRC_DIR%

set GOOS=linux
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.linux-arm64 %SRC_DIR%
if errorlevel 1 goto :error

REM macOS
echo Building for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.darwin-amd64 %SRC_DIR%

set GOOS=darwin
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.darwin-arm64 %SRC_DIR%
if errorlevel 1 goto :error

REM Copy fingerprint files
xcopy /Y /E %FINGERPRINTS_DIR%\* %OUT_DIR%\fingerprints\
if errorlevel 1 goto :error

goto :eof

REM Error handling
:error
echo An error occurred during the build process.
exit /b 1

REM Usage information
:usage
echo Usage: %0 ^<command^>
echo Commands:
echo   build       Build for Linux and Windows
echo   build-all   Build for Linux and Windows
echo   clean       Clean the output directory
exit /b 1

:eof
endlocal
