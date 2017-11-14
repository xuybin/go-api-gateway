@echo off

FOR /F "delims=" %%I IN ("govendor.exe") DO (
if exist %%~$PATH:I (
    	echo 'govendor exist'
    ) else (
	echo 'govendor does not exist'
    	go get -u -v github.com/kardianos/govendor
    )
)

set "pwd=%~dp0"
for /f "delims=" %%i in ("%pwd:~0,-1%") do (set "pwd=%%~ni")

set GOOS=linux
set GOARCH=amd64
echo 'govendor build -o "%pwd%-%GOOS%-%GOARCH%'
govendor build -o "%pwd%-%GOOS%-%GOARCH%"


set GOOS=windows
set GOARCH=amd64
echo 'govendor build -o "%pwd%-%GOOS%-%GOARCH%.exe"'
govendor build -o "%pwd%-%GOOS%-%GOARCH%.exe"


set GOOS=darwin
set GOARCH=amd64
echo 'govendor build -o "%pwd%-%GOOS%-%GOARCH%'
govendor build -o "%pwd%-%GOOS%-%GOARCH%"