package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type GUIApp struct {
	app            fyne.App
	window         fyne.Window
	httpServer     *http.Server
	
	// UI 组件
	sessionSelect  *widget.Select
	headlessMode   *widget.Check
	titleEntry     *widget.Entry
	contentEntry   *widget.MultiLineEntry
	imageList      *widget.List
	selectedImages []string
	statusLabel    *widget.Label
	logText        *widget.RichText
	contentList    *widget.List
	searchEntry    *widget.Entry
	searchResults  *widget.List
	
	// 数据
	sessions       []string
	feeds          []interface{}
	searchResults  []interface{}
}

func main() {
	// 创建 Fyne 应用
	myApp := app.NewWithID("com.xiaohongshu.mcp")
	myApp.SetMetadata(&fyne.AppMetadata{
		ID:      "com.xiaohongshu.mcp",
		Name:    "小红书管理平台",
		Version: "1.0.0",
	})

	guiApp := &GUIApp{
		app:            myApp,
		selectedImages: make([]string, 0),
		sessions:       make([]string, 0),
		feeds:          make([]interface{}, 0),
		searchResults:  make([]interface{}, 0),
	}

	// 初始化服务
	// 这里可以添加初始化逻辑

	// 创建主窗口
	guiApp.window = myApp.NewWindow("小红书管理平台")
	guiApp.window.Resize(fyne.NewSize(800, 600))
	guiApp.window.SetMaster()

	// 创建UI
	guiApp.createUI()

	// 启动后台HTTP服务器
	go guiApp.startHTTPServer()

	// 加载会话列表
	guiApp.loadSessions()

	// 显示窗口并运行
	guiApp.window.ShowAndRun()
}

func (g *GUIApp) createUI() {
	// 创建标签页容器
	tabs := container.NewAppTabs(
		container.NewTabItem("发布内容", g.createPublishTab()),
		container.NewTabItem("内容管理", g.createContentTab()),
		container.NewTabItem("搜索", g.createSearchTab()),
		container.NewTabItem("日志", g.createLogTab()),
	)

	g.window.SetContent(tabs)
}

func (g *GUIApp) createPublishTab() *fyne.Container {
	// 会话选择
	sessionLabel := widget.NewLabel("选择会话:")
	g.sessionSelect = widget.NewSelect([]string{}, func(value string) {
		g.updateStatus("已选择会话: " + value)
	})
	g.sessionSelect.PlaceHolder = "请选择会话"

	// 无头模式
	g.headlessMode = widget.NewCheck("无头模式", func(checked bool) {
		configs.InitHeadless(checked)
		g.updateStatus(fmt.Sprintf("无头模式: %v", checked))
	})
	g.headlessMode.SetChecked(true)

	// 标题输入
	titleLabel := widget.NewLabel("标题:")
	g.titleEntry = widget.NewEntry()
	g.titleEntry.SetPlaceHolder("请输入标题")

	// 内容输入
	contentLabel := widget.NewLabel("内容:")
	g.contentEntry = widget.NewMultiLineEntry()
	g.contentEntry.SetPlaceHolder("请输入内容")

	// 图片选择
	imageLabel := widget.NewLabel("选择图片:")
	selectImageBtn := widget.NewButton("选择图片", g.selectImages)
	clearImageBtn := widget.NewButton("清空图片", g.clearImages)
	
	g.imageList = widget.NewList(
		func() int { return len(g.selectedImages) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(g.selectedImages[id])
		},
	)

	// 发布按钮
	publishBtn := widget.NewButton("发布内容", g.publishContent)

	// 状态标签
	g.statusLabel = widget.NewLabel("就绪")

	// 布局
	topContainer := container.NewHBox(
		sessionLabel, g.sessionSelect,
		widget.NewButton("刷新会话", g.loadSessions),
		g.headlessMode,
	)

	imageContainer := container.NewVBox(
		imageLabel,
		container.NewHBox(selectImageBtn, clearImageBtn),
		g.imageList,
	)

	formContainer := container.NewVBox(
		titleLabel, g.titleEntry,
		contentLabel, g.contentEntry,
		imageContainer,
		publishBtn,
		g.statusLabel,
	)

	return container.NewVBox(
		topContainer,
		widget.NewSeparator(),
		formContainer,
	)
}

func (g *GUIApp) createContentTab() *fyne.Container {
	// 内容列表
	refreshBtn := widget.NewButton("刷新内容", g.loadFeeds)
	g.contentList = widget.NewList(
		func() int { return len(g.feeds) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(g.feeds) {
				feed := g.feeds[id].(map[string]interface{})
				title := feed["title"].(string)
				obj.(*widget.Label).SetText(title)
			}
		},
	)

	return container.NewVBox(
		refreshBtn,
		g.contentList,
	)
}

func (g *GUIApp) createSearchTab() *fyne.Container {
	// 搜索输入
	g.searchEntry = widget.NewEntry()
	g.searchEntry.SetPlaceHolder("请输入搜索关键词")
	searchBtn := widget.NewButton("搜索", g.searchContent)

	// 搜索结果
	g.searchResults = widget.NewList(
		func() int { return len(g.searchResults) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(g.searchResults) {
				result := g.searchResults[id].(map[string]interface{})
				title := result["title"].(string)
				obj.(*widget.Label).SetText(title)
			}
		},
	)

	return container.NewVBox(
		container.NewHBox(g.searchEntry, searchBtn),
		g.searchResults,
	)
}

func (g *GUIApp) createLogTab() *fyne.Container {
	g.logText = widget.NewRichText()
	g.logText.Wrapping = fyne.TextWrapWord
	
	clearLogBtn := widget.NewButton("清空日志", func() {
		g.logText.ParseMarkdown("")
	})

	return container.NewVBox(
		clearLogBtn,
		g.logText,
	)
}

func (g *GUIApp) startHTTPServer() {
	// 创建简单的HTTP服务器用于API调用
	mux := http.NewServeMux()
	
	// 模拟API端点
	mux.HandleFunc("/api/v1/sessions", g.handleSessions)
	mux.HandleFunc("/api/v1/login/status", g.handleLoginStatus)
	mux.HandleFunc("/api/v1/publish", g.handlePublish)
	mux.HandleFunc("/api/v1/feeds/list", g.handleFeedsList)
	mux.HandleFunc("/api/v1/feeds/search", g.handleFeedsSearch)

	g.httpServer = &http.Server{
		Addr:    ":18060",
		Handler: mux,
	}

	g.addLog("HTTP服务器启动在端口 18060")
	if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		g.addLog(fmt.Sprintf("HTTP服务器启动失败: %v", err))
	}
}

func (g *GUIApp) selectImages() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			g.addLog(fmt.Sprintf("选择图片失败: %v", err))
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// 这里可以添加多文件选择逻辑
		path := reader.URI().Path()
		g.selectedImages = append(g.selectedImages, path)
		g.imageList.Refresh()
		g.addLog(fmt.Sprintf("已选择图片: %s", path))
	}, g.window)
}

func (g *GUIApp) clearImages() {
	g.selectedImages = make([]string, 0)
	g.imageList.Refresh()
	g.addLog("已清空图片列表")
}

func (g *GUIApp) loadSessions() {
	// 模拟加载会话
	g.sessions = []string{"session-001", "session-002", "session-003"}
	g.sessionSelect.Options = g.sessions
	g.sessionSelect.Refresh()
	g.addLog("已加载会话列表")
}

func (g *GUIApp) loadFeeds() {
	// 模拟加载内容
	g.feeds = []interface{}{
		map[string]interface{}{"title": "测试内容1", "content": "这是测试内容1"},
		map[string]interface{}{"title": "测试内容2", "content": "这是测试内容2"},
	}
	g.contentList.Refresh()
	g.addLog("已加载内容列表")
}

func (g *GUIApp) searchContent() {
	keyword := g.searchEntry.Text
	if keyword == "" {
		dialog.ShowInformation("提示", "请输入搜索关键词", g.window)
		return
	}

	// 模拟搜索
	g.searchResults = []interface{}{
		map[string]interface{}{"title": "搜索结果1: " + keyword, "content": "这是搜索结果1"},
		map[string]interface{}{"title": "搜索结果2: " + keyword, "content": "这是搜索结果2"},
	}
	g.searchResults.Refresh()
	g.addLog(fmt.Sprintf("搜索关键词: %s", keyword))
}

func (g *GUIApp) publishContent() {
	title := g.titleEntry.Text
	content := g.contentEntry.Text
	session := g.sessionSelect.Selected

	if title == "" || content == "" {
		dialog.ShowInformation("提示", "请输入标题和内容", g.window)
		return
	}

	if session == "" {
		dialog.ShowInformation("提示", "请选择会话", g.window)
		return
	}

	g.updateStatus("正在发布...")
	g.addLog(fmt.Sprintf("发布内容: %s", title))

	// 这里调用实际的发布API
	// 模拟发布过程
	go func() {
		time.Sleep(2 * time.Second)
		g.updateStatus("发布成功")
		g.addLog("内容发布成功")
		
		// 清空表单
		g.titleEntry.SetText("")
		g.contentEntry.SetText("")
		g.clearImages()
	}()
}

func (g *GUIApp) updateStatus(message string) {
	g.statusLabel.SetText(message)
	g.addLog(message)
}

func (g *GUIApp) addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	// 在主线程中更新UI
	g.app.RunOnMain(func() {
		currentText := g.logText.ParseMarkdown("")
		g.logText.ParseMarkdown(currentText + logMessage)
	})
}

// HTTP 处理器
func (g *GUIApp) handleSessions(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"sessions": g.sessions,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (g *GUIApp) handleLoginStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
		"data":    true,
	}
	json.NewEncoder(w).Encode(response)
}

func (g *GUIApp) handlePublish(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
		"message": "发布成功",
	}
	json.NewEncoder(w).Encode(response)
}

func (g *GUIApp) handleFeedsList(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"feeds": g.feeds,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (g *GUIApp) handleFeedsSearch(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"feeds": g.searchResults,
		},
	}
	json.NewEncoder(w).Encode(response)
}
