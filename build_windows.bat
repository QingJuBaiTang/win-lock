@echo off
echo 正在编译课堂锁屏...
go build -ldflags="-H windowsgui -s -w" -o win-lock.exe .
if %ERRORLEVEL% == 0 (
    echo 编译成功：win-lock.exe
) else (
    echo 编译失败，请检查是否安装了 Go 和 MinGW-w64
)
pause
