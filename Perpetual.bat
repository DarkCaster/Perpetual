@echo off
setlocal EnableDelayedExpansion

rem Perpetual.bat - launcher script that detects the current Windows version
rem and CPU architecture, then executes the matching Perpetual binary located
rem alongside this script.
rem
rem Fallback: if a matching binary cannot be found next to this script,
rem a generic "Perpetual.exe" binary in the current working directory will
rem be used instead.
rem
rem All command line arguments are passed through to the selected binary,
rem stdin/stdout/stderr are inherited naturally, and its exit code becomes
rem the exit code of this script.

set "SCRIPT_DIR=%~dp0"
if "%SCRIPT_DIR:~-1%"=="\" set "SCRIPT_DIR=%SCRIPT_DIR:~0,-1%"

rem ---------------------------------------------------------------------
rem Detect whether we are running on a legacy Windows version
rem (Windows 7 / Windows Server 2008 R2, NT kernel version 6.1, or older).
rem Only an x86 build is produced for such systems.
rem ---------------------------------------------------------------------

set "IS_LEGACY=0"

reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion" /v CurrentMajorVersionNumber >nul 2>&1
if !ERRORLEVEL! EQU 0 (
    rem CurrentMajorVersionNumber exists only on Windows 10/11 and later.
    set "IS_LEGACY=0"
) else (
    set "CURVER="
    for /f "tokens=3" %%v in ('reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion" /v CurrentVersion 2^>nul ^| findstr "REG_SZ"') do (
        set "CURVER=%%v"
    )
    if defined CURVER (
        set "VMAJOR="
        set "VMINOR="
        for /f "tokens=1,2 delims=." %%a in ("!CURVER!") do (
            set "VMAJOR=%%a"
            set "VMINOR=%%b"
        )
        if defined VMAJOR (
            if !VMAJOR! LSS 6 (
                set "IS_LEGACY=1"
            ) else if !VMAJOR! EQU 6 (
                if defined VMINOR (
                    if !VMINOR! LSS 2 (
                        set "IS_LEGACY=1"
                    )
                ) else (
                    set "IS_LEGACY=1"
                )
            )
        )
    )
)

rem ---------------------------------------------------------------------
rem Detect CPU architecture and map it to the naming scheme used by
rem release binaries.
rem ---------------------------------------------------------------------

set "RAWARCH=%PROCESSOR_ARCHITECTURE%"
if defined PROCESSOR_ARCHITEW6432 set "RAWARCH=%PROCESSOR_ARCHITEW6432%"

set "ARCH=unknown"
if /i "%RAWARCH%"=="AMD64" set "ARCH=amd64"
if /i "%RAWARCH%"=="ARM64" set "ARCH=arm64"
if /i "%RAWARCH%"=="x86" set "ARCH=x86"

rem ---------------------------------------------------------------------
rem Select the appropriate binary located next to this script.
rem ---------------------------------------------------------------------

set "TARGET_BIN="

if "%IS_LEGACY%"=="1" (
    if exist "%SCRIPT_DIR%\Perpetual_win7_x86.exe" (
        set "TARGET_BIN=%SCRIPT_DIR%\Perpetual_win7_x86.exe"
    )
) else (
    if not "%ARCH%"=="unknown" (
        if exist "%SCRIPT_DIR%\Perpetual_%ARCH%.exe" (
            set "TARGET_BIN=%SCRIPT_DIR%\Perpetual_%ARCH%.exe"
        )
    )
)

rem Fallback to a generic "Perpetual.exe" binary
if not defined TARGET_BIN (
    if exist "%SCRIPT_DIR%\Perpetual.exe" (
        set "TARGET_BIN=%SCRIPT_DIR%\Perpetual.exe"
    )
)
if not defined TARGET_BIN (
    if exist "Perpetual.exe" (
        set "TARGET_BIN=Perpetual.exe"
    )
)

if not defined TARGET_BIN (
    echo Error: could not find a suitable Perpetual binary to run. 1>&2
    echo Looked for a matching binary in "%SCRIPT_DIR%" and for "Perpetual.exe" in the current directory. 1>&2
    exit /b 1
)

"%TARGET_BIN%" %* && exit 0 || exit 1
