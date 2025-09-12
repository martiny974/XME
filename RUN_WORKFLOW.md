## 运行脚本与验证流程速查

> 一次性照着跑即可；如有 Windows Defender 拦截问题，见文末“常见问题”。

### 0. 环境
- 安装 Go ≥ 1.23.5（`go version` 查看）

### 1. 首次登录（保存 Cookies）
- 使用脚本（推荐）
```bash
./run.sh --only-login
```
- 或直接运行登录程序
```bash
go run ./cmd/login/main.go
```

### 2. 启动服务（MCP HTTP Server）
- 无头模式（默认）
```bash
./run.sh
# 或
go run .
```
- 有界面模式（调试可视化）
```bash
./run.sh --with-ui
# 或
go run . -headless=false
```
- 服务地址：`http://localhost:18060/mcp`

### 3. 验证服务联通
- 方式 A：MCP Inspector（可视化最直观）
```bash
npx @modelcontextprotocol/inspector
# 浏览器打开 Inspector → URL 填入 http://localhost:18060/mcp → Connect → List Tools
```

- 方式 B：curl（命令行最简）
```bash
# 初始化
curl -X POST http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","id":1}'

# 列表工具
curl -X POST http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":2}'
```

- 方式 C：Cursor（你已经可用 `.cursor/mcp.json`）
  1) 确保服务在跑
  2) 重启 Cursor
  3) 在聊天右侧 Available Tools 中使用 `xiaohongshu-mcp`

- 方式 D：Claude Code CLI
```bash
claude mcp add --transport http xiaohongshu-mcp http://localhost:18060/mcp
```

### 4. 常用脚本参数（run.sh）
- `./run.sh`：安装依赖、检测 Cookies、启动（默认无头）
- `./run.sh --with-ui`：启动有界面浏览器（非无头）
- `./run.sh --only-login`：只执行登录，保存 Cookies 后退出
- `./run.sh --skip-login`：跳过自动登录检查

Cookies 默认路径（系统临时目录）：
- Windows：`%TEMP%\cookies.json`
- macOS/Linux：`$TMPDIR/cookies.json` 或 `/tmp/cookies.json`

### 5. 常见问题（Windows）
- Defender 拦截 leakless.exe（go-rod 启动助手）
  - 已在登录程序中禁用 leakless；如仍遇到，临时指定浏览器后运行：
```bash
# Git Bash 当前会话
ROD_BROWSER_BIN="C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe" \
  go run ./cmd/login/main.go
```
- 直接用浏览器打开 `http://localhost:18060/mcp` 显示 “SSE not requested”
  - 该路径的 GET 是 SSE 通道，需要 `Accept: text/event-stream`；请用 MCP Inspector 或 curl POST 来验证。


