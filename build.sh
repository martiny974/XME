#!/bin/bash

echo "正在构建小红书MCP服务..."

# 设置环境变量
export CGO_ENABLED=0

# 创建输出目录
mkdir -p dist

# 编译主程序
echo "编译主程序..."
go build -ldflags "-s -w" -o dist/xiaohongshu-mcp ./cmd/xiaohongshu-mcp

if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi

# 复制必要的文件到输出目录
echo "复制配置文件..."
mkdir -p dist/cookies
mkdir -p dist/logs

# 创建启动脚本
echo "创建启动脚本..."
cat > dist/start.sh << 'EOF'
#!/bin/bash
echo "启动小红书MCP服务..."
./xiaohongshu-mcp
EOF

chmod +x dist/start.sh
chmod +x dist/xiaohongshu-mcp

echo "构建完成！"
echo "可执行文件位置: dist/xiaohongshu-mcp"
echo "启动脚本位置: dist/start.sh"
echo ""
echo "使用方法:"
echo "1. 运行 ./dist/start.sh 启动服务"
echo "2. 或者直接运行 ./dist/xiaohongshu-mcp"
echo "3. 程序会自动打开浏览器访问管理界面"
echo ""
