@echo off
title RTV-RTC Launcher
cd /d "%~dp0"

echo ===========================================================
echo   RTV-RTC  -  Invitaciones / Remitos divertidos
echo ===========================================================
echo.
echo  Backend : http://localhost:8080
echo  Frontend: http://localhost:4200  (se abre solo en el navegador)
echo.
echo  Se abriran DOS ventanas. Para detener la app, cerralas.
echo ===========================================================
echo.

start "RTV Backend"  cmd /k "cd backend & backend.exe"
start "RTV Frontend" cmd /k "cd frontend & if not exist node_modules npm install & npx ng serve --open"

echo Listo. Esperando a que el navegador abra http://localhost:4200 ...
