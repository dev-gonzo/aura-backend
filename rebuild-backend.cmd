@echo off
setlocal

title Aura Editora Backend - Rebuild Docker

set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
for %%I in ("%PROJECT_ROOT%") do set "PROJECT_ROOT=%%~fI"

set "COMPOSE_FILE=%PROJECT_ROOT%\docker-compose.yml"
set "SERVICE_NAME=editora-backend"
set "EXPECTED_CONTEXT=%PROJECT_ROOT%\editora-backend"
set "ACTUAL_BACKEND_DIR=%SCRIPT_DIR:~0,-1%"

echo ===============================================
echo Aura Editora Backend - rebuild do container
echo ===============================================
echo.

if not exist "%COMPOSE_FILE%" (
  echo [ERRO] Nao encontrei o docker-compose.yml em:
  echo        %COMPOSE_FILE%
  echo.
  pause
  exit /b 1
)

if /I not "%EXPECTED_CONTEXT%"=="%ACTUAL_BACKEND_DIR%" (
  echo [AVISO] O docker-compose.yml parece apontar para outro caminho de build.
  echo         Contexto esperado pelo compose:
  echo         %EXPECTED_CONTEXT%
  echo.
  echo         Pasta atual do backend:
  echo         %ACTUAL_BACKEND_DIR%
  echo.
  echo         Se o compose nao estiver alinhado com a pasta real, o rebuild pode falhar.
  echo.
)

where docker >nul 2>nul
if errorlevel 1 (
  echo [ERRO] O Docker nao esta disponivel no PATH do Windows.
  echo.
  pause
  exit /b 1
)

echo [1/3] Derrubando o container atual do backend...
docker compose -f "%COMPOSE_FILE%" stop "%SERVICE_NAME%"
if errorlevel 1 (
  echo.
  echo [ERRO] Nao foi possivel parar o container atual do backend.
  echo.
  pause
  exit /b 1
)

docker compose -f "%COMPOSE_FILE%" rm -f "%SERVICE_NAME%"
if errorlevel 1 (
  echo.
  echo [ERRO] Nao foi possivel remover o container atual do backend.
  echo.
  pause
  exit /b 1
)

echo.
echo [2/3] Rebuildando a imagem do backend...
docker compose -f "%COMPOSE_FILE%" build "%SERVICE_NAME%"
if errorlevel 1 (
  echo.
  echo [ERRO] O build do backend falhou.
  echo.
  pause
  exit /b 1
)

echo.
echo [3/3] Subindo o backend atualizado...
docker compose -f "%COMPOSE_FILE%" up -d --force-recreate --no-deps "%SERVICE_NAME%"
if errorlevel 1 (
  echo.
  echo [ERRO] Nao foi possivel subir o backend atualizado.
  echo.
  pause
  exit /b 1
)

echo.
echo [OK] Backend rebuildado e container iniciado com sucesso.
echo.
pause
