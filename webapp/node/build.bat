cd node
call npm run build
cd ..
Xcopy node\build ..\bin\www\node /E /I /Y
rem go-bindata -pkg httpdata -o res.go www/...

echo Complete. Press Enter.
pause
