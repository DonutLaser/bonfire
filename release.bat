@echo off
go build -ldflags -H=windowsgui
ResourceHacker -open bonfire.exe -save bonfire.exe -action addskip -res assets/images/icon.ico -mask ICONGROUP,MAIN,
xcopy /s /y assets\* D:\Programos\custom\bonfire\assets\
xcopy /y bonfire.exe D:\Programos\custom\bonfire\