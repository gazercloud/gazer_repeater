set UTIL_NAME=cu_base64

cd main

set GOOS=windows
set GOARCH=386
go build -o ../bin/win32/%UTIL_NAME%.exe

set GOOS=windows
set GOARCH=amd64
go build -o ../bin/win64/%UTIL_NAME%.exe

cd ..

pause
