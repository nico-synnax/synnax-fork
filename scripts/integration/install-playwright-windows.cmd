@echo off

rem Copyright 2025 Synnax Labs, Inc.
rem
rem Use of this software is governed by the Business Source License included in the file
rem licenses/BSL.txt.
rem
rem As of the Change Date specified in that file, in accordance with the Business Source
rem License, use of this software will be governed by the Apache License, Version 2.0,
rem included in the file licenses/APL.txt.

rem install-playwright-windows.cmd
rem Installs Playwright browsers using Poetry on Windows
rem Used by GitHub Actions workflow: test.integration.yaml

echo 📦 Installing Playwright browsers on Windows...

rem Change to the integration test directory
cd integration\test\py

rem Try to find Poetry executable
set "POETRY_CMD=poetry"
poetry --version >nul 2>nul
if %errorlevel% neq 0 (
    echo Poetry not found in PATH, searching for executable...
    if exist "%APPDATA%\Python\Scripts\poetry.exe" (
        set "POETRY_CMD=%APPDATA%\Python\Scripts\poetry.exe"
        echo Found Poetry at: %POETRY_CMD%
    ) else if exist "%APPDATA%\pypoetry\venv\Scripts\poetry.exe" (
        set "POETRY_CMD=%APPDATA%\pypoetry\venv\Scripts\poetry.exe"
        echo Found Poetry at: %POETRY_CMD%
    ) else if exist "%USERPROFILE%\.local\bin\poetry.exe" (
        set "POETRY_CMD=%USERPROFILE%\.local\bin\poetry.exe"
        echo Found Poetry at: %POETRY_CMD%
    ) else (
        echo ❌ Poetry executable not found
        exit /b 1
    )
)

rem Install Playwright browsers
echo Installing Playwright browsers...
"%POETRY_CMD%" run playwright install --with-deps
if %errorlevel% neq 0 exit /b %errorlevel%

echo ✅ Playwright browsers installed successfully