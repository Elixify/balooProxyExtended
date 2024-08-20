@echo off
setlocal

REM Variables
set BINARY_NAME=main
set SRC_DIR=.
set OUT_DIR=out
set FINGERPRINTS_DIR=.\global\fingerprints
set TEMPLATES_DIR=.\assets\html

REM Ensure the output directory exists
if not exist %OUT_DIR% (
    mkdir %OUT_DIR%
)

REM Check the input argument
if "%1" == "windows-amd64" goto build_windows_amd64
if "%1" == "windows-arm64" goto build_windows_arm64
if "%1" == "linux-amd64" goto build_linux_amd64
if "%1" == "linux-arm64" goto build_linux_arm64
if "%1" == "darwin-amd64" goto build_darwin_amd64
if "%1" == "darwin-arm64" goto build_darwin_arm64
if "%1" == "freebsd-amd64" goto build_freebsd_amd64
if "%1" == "freebsd-arm64" goto build_freebsd_arm64
if "%1" == "openbsd-amd64" goto build_openbsd_amd64
if "%1" == "openbsd-arm64" goto build_openbsd_arm64
if "%1" == "all" goto build_all
if "%1" == "clean" goto clean
goto :usage

REM Clean up the build
:clean
echo Cleaning up...
if exist %OUT_DIR% rd /S /Q %OUT_DIR%
goto :eof

REM Build for Windows AMD64
:build_windows_amd64
echo Building for Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.windows-amd64.exe %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for Windows ARM64
:build_windows_arm64
echo Building for Windows ARM64...
set GOOS=windows
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.windows-arm64.exe %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for Linux AMD64
:build_linux_amd64
echo Building for Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.linux-amd64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for Linux ARM64
:build_linux_arm64
echo Building for Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.linux-arm64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for macOS AMD64
:build_darwin_amd64
echo Building for macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.darwin-amd64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for macOS ARM64
:build_darwin_arm64
echo Building for macOS ARM64...
set GOOS=darwin
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.darwin-arm64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for Freebsd ARM64
:build_freebsd_arm64
echo Building for FreeBSD ARM64...
set GOOS=freebsd
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.freebsd-arm64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for Freebsd AMD64
:build_freebsd_amd64
echo Building for FreeBSD AMD64...
set GOOS=freebsd
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.freebsd-amd64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for OpenBSD ARM64
:build_openbsd_arm64
echo Building for OpenBSD ARM64...
set GOOS=openbsd
set GOARCH=arm64
go build -o %OUT_DIR%\%BINARY_NAME%.openbsd-arm64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Build for OpenBSD AMD64
:build_openbsd_amd64
echo Building for OpenBSD AMD64...
set GOOS=openbsd
set GOARCH=amd64
go build -o %OUT_DIR%\%BINARY_NAME%.openbsd-amd64 %SRC_DIR%
if errorlevel 1 goto :error
goto :copy_files

REM Copy static files
:copy_files
xcopy /Y /E %FINGERPRINTS_DIR%\* %OUT_DIR%\fingerprints\ >nul
xcopy /Y /E %TEMPLATES_DIR%\* %OUT_DIR%\html\ >nul
if errorlevel 1 goto :error
goto :eof

REM Build all platforms
:build_all
echo Building for all platforms...

call :build_windows_amd64
call :build_windows_arm64
call :build_linux_amd64
call :build_linux_arm64
call :build_darwin_amd64
call :build_darwin_arm64
call :build_freebsd_arm64
call :build_freebsd_amd64
call :build_openbsd_arm64
call :build_openbsd_amd64

goto :eof

REM Error handling
:error
echo An error occurred during the build process.
exit /b 1

REM Usage information
:usage
echo Usage: %0 ^<command^>
echo Commands:
echo   windows-amd64     Build for Windows AMD64
echo   windows-arm64     Build for Windows ARM64
echo   linux-amd64       Build for Linux AMD64
echo   linux-arm64       Build for Linux ARM64
echo   darwin-amd64      Build for macOS AMD64
echo   darwin-arm64      Build for macOS ARM64
echo   freebsd-amd64     Build for FreeBSD AMD64
echo   freebsd-arm64     Build for FreeBSD ARM64
echo   openbsd-amd64     Build for OpenBSD AMD64
echo   openbsd-arm64     Build for OpenBSD ARM64
echo   all               Build for all platforms
echo   clean             Clean the output directory
exit /b 1

:eof
endlocal
