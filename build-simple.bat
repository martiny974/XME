@echo off
echo 正在构建小红书MCP服务...

REM 创建输出目录
if not exist "dist" mkdir dist

REM 编译当前平台版本
echo 编译程序...
go build -ldflags "-s -w" -o dist/xiaohongshu-mcp.exe .

if %ERRORLEVEL% neq 0 (
    echo 编译失败！
    pause
    exit /b 1
)

REM 创建启动脚本
echo 创建启动脚本...
echo @echo off > dist/start.bat
echo echo 启动小红书MCP服务... >> dist/start.bat
echo echo 程序将自动打开浏览器访问管理界面 >> dist/start.bat
echo echo 按 Ctrl+C 可退出程序 >> dist/start.bat
echo echo. >> dist/start.bat
echo xiaohongshu-mcp.exe >> dist/start.bat

REM 创建使用说明
echo 创建使用说明...
echo 小红书MCP服务 > dist/使用说明.txt
echo. >> dist/使用说明.txt
echo 使用方法: >> dist/使用说明.txt
echo 1. 双击 start.bat 启动服务 >> dist/使用说明.txt
echo 2. 或者直接运行 xiaohongshu-mcp.exe >> dist/使用说明.txt
echo 3. 程序会自动打开浏览器访问管理界面 >> dist/使用说明.txt
echo. >> dist/使用说明.txt
echo 功能说明: >> dist/使用说明.txt
echo - 支持多账号登录管理 >> dist/使用说明.txt
echo - 支持发布图文内容 >> dist/使用说明.txt
echo - 支持查询已发布内容 >> dist/使用说明.txt
echo - 支持搜索功能 >> dist/使用说明.txt
echo. >> dist/使用说明.txt
echo 注意事项: >> dist/使用说明.txt
echo - 首次使用需要登录小红书账号 >> dist/使用说明.txt
echo - 默认端口: 18060 >> dist/使用说明.txt
echo - 按 Ctrl+C 退出程序 >> dist/使用说明.txt
echo - 登录信息会保存在 cookies 目录中 >> dist/使用说明.txt
echo - 如需修改端口: xiaohongshu-mcp.exe -port 8080 >> dist/使用说明.txt

echo.
echo ==========================================
echo 构建完成！
echo ==========================================
echo.
echo 输出文件:
echo - dist/xiaohongshu-mcp.exe    (主程序)
echo - dist/start.bat              (启动脚本)
echo - dist/使用说明.txt           (使用说明)
echo.
echo 文件大小: 
dir dist\xiaohongshu-mcp.exe
echo.
echo 使用方法:
echo 1. 进入 dist 目录
echo 2. 双击 start.bat 启动服务
echo 3. 程序会自动打开浏览器
echo.
pause
