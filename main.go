package main

import (
	"embed"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

//go:embed XhsMcpWeb.html login.html
var webContent embed.FS

func main() {
	var (
		headless bool
		port     string
		noBrowser bool
	)

	flag.BoolVar(&headless, "headless", false, "是否无头模式")
	flag.StringVar(&port, "port", "18060", "服务端口")
	flag.BoolVar(&noBrowser, "no-browser", false, "不自动打开浏览器")
	flag.Parse()

	configs.InitHeadless(headless)

	// 初始化服务
	xiaohongshuService := NewXiaohongshuService()

	// 创建并启动应用服务器
	appServer := NewAppServer(xiaohongshuService)

	// 在 goroutine 中启动服务器
	go func() {
		logrus.Infof("服务器启动在端口 %s", port)
		if err := appServer.Start(":" + port); err != nil {
			logrus.Fatalf("failed to run server: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 自动打开浏览器（除非指定不打开）
	if !noBrowser {
		url := fmt.Sprintf("http://localhost:%s", port)
		logrus.Infof("正在打开浏览器: %s", url)
		
		if err := openBrowser(url); err != nil {
			logrus.Warnf("无法自动打开浏览器: %v", err)
			fmt.Printf("请手动打开浏览器访问: %s\n", url)
		}
	} else {
		url := fmt.Sprintf("http://localhost:%s", port)
		fmt.Printf("服务已启动，请手动打开浏览器访问: %s\n", url)
	}

	// 保持程序运行
	fmt.Println("按 Ctrl+C 退出程序")
	select {}
}

// openBrowser 打开浏览器
func openBrowser(url string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	
	return cmd.Start()
}
