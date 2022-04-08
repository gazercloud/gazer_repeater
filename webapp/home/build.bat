cd home
call npm run build
cd ..
Xcopy home\build ..\bin\www\home /E /I /Y
rem go-bindata -pkg httpdata -o res.go www/...
pause
