package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/webview/webview"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
	"github.com/xpzouying/xiaohongshu-mcp/middleware"
	"github.com/xpzouying/xiaohongshu-mcp/routes"
	"github.com/xpzouying/xiaohongshu-mcp/service"
)

//go:embed XhsMcpWeb.html
var webContent embed.FS

type DesktopApp struct {
	xiaohongshuService *service.XiaohongshuService
	httpServer         *http.Server
	webview            webview.WebView
}

func main() {
	// 设置日志格式
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// 初始化服务
	configs.InitHeadless(true) // 默认无头模式
	xiaohongshuService := service.NewXiaohongshuService()

	desktopApp := &DesktopApp{
		xiaohongshuService: xiaohongshuService,
	}

	// 启动HTTP服务器
	go desktopApp.startHTTPServer()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 创建webview
	desktopApp.createWebView()

	// 运行webview
	desktopApp.webview.Run()
}

func (d *DesktopApp) startHTTPServer() {
	// 创建服务实例
	appServer := service.NewAppServer(d.xiaohongshuService)

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 添加中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ErrorHandler())

	// 静态文件服务 - 提供嵌入的 HTML 文件
	r.GET("/", func(c *gin.Context) {
		content, err := webContent.ReadFile("XhsMcpWeb.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "无法加载网页文件")
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, string(content))
	})

	// 登录页面
	r.GET("/login.html", func(c *gin.Context) {
		content, err := webContent.ReadFile("login.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "无法加载登录页面")
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, string(content))
	})

	// 设置路由
	routes.SetupRoutes(r, appServer)

	// 启动服务器
	port := "18060"
	d.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	logrus.Infof("HTTP服务器启动在端口 %s", port)
	if err := d.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("HTTP服务器启动失败: %v", err)
	}
}

func (d *DesktopApp) createWebView() {
	// 创建webview
	d.webview = webview.New(true)
	defer d.webview.Destroy()

	// 设置窗口属性
	d.webview.SetTitle("小红书管理平台")
	d.webview.SetSize(1200, 800, webview.HintNone)

	// 加载本地HTML页面
	url := "http://localhost:18060"
	logrus.Infof("加载页面: %s", url)
	d.webview.Navigate(url)

	// 添加JavaScript API绑定
	d.webview.Bind("openExternal", func(url string) {
		d.openExternal(url)
	})

	d.webview.Bind("showMessage", func(message string) {
		logrus.Info("JavaScript消息:", message)
	})

	// 设置窗口关闭时的处理
	d.webview.OnClose(func() {
		logrus.Info("窗口关闭，正在退出...")
		if d.httpServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			d.httpServer.Shutdown(ctx)
		}
		os.Exit(0)
	})
}

func (d *DesktopApp) openExternal(url string) {
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
