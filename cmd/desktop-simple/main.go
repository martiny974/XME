package main

import (
	"embed"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/webview/webview"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

//go:embed XhsMcpWeb.html
var webContent embed.FS

func main() {
	var (
		headless bool
		port     string
	)

	flag.BoolVar(&headless, "headless", true, "是否无头模式")
	flag.StringVar(&port, "port", "18060", "服务端口")
	flag.Parse()

	configs.InitHeadless(headless)

	// 启动HTTP服务器（复用现有的main.go逻辑）
	go startHTTPServer(port)

	// 等待服务器启动
	time.Sleep(3 * time.Second)

	// 创建webview桌面应用
	createDesktopApp(port)
}

func startHTTPServer(port string) {
	// 这里复用现有的HTTP服务器启动逻辑
	// 为了简化，我们直接启动现有的main程序
	cmd := exec.Command("go", "run", ".", "-port", port, "-headless=true")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func createDesktopApp(port string) {
	// 创建webview
	w := webview.New(true)
	defer w.Destroy()

	// 设置窗口属性
	w.SetTitle("小红书管理平台 - 桌面版")
	w.SetSize(1200, 800, webview.HintNone)

	// 加载本地HTML页面
	url := fmt.Sprintf("http://localhost:%s", port)
	logrus.Infof("桌面应用加载页面: %s", url)
	w.Navigate(url)

	// 添加JavaScript API绑定
	w.Bind("openExternal", func(url string) {
		openExternal(url)
	})

	w.Bind("showMessage", func(message string) {
		logrus.Info("JavaScript消息:", message)
	})

	// 设置窗口关闭时的处理
	w.OnClose(func() {
		logrus.Info("桌面应用关闭")
		os.Exit(0)
	})

	// 运行webview
	w.Run()
}

func openExternal(url string) {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		logrus.Errorf("不支持的操作系统: %s", runtime.GOOS)
		return
	}
	
	if err := cmd.Start(); err != nil {
		logrus.Errorf("打开外部链接失败: %v", err)
	}
}
