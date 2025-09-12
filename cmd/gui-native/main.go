package main

import (
	"fmt"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/sirupsen/logrus"
)

type MainWindow struct {
	*walk.MainWindow
	sessionCombo    *walk.ComboBox
	headlessCheck   *walk.CheckBox
	titleEdit       *walk.LineEdit
	contentEdit     *walk.TextEdit
	imageList       *walk.ListBox
	statusLabel     *walk.Label
	logText         *walk.TextEdit
	contentList     *walk.ListBox
	searchEdit      *walk.LineEdit
	searchList      *walk.ListBox
	
	// 数据
	sessions        []string
	selectedImages  []string
	feeds           []map[string]interface{}
	searchResults   []map[string]interface{}
}

func main() {
	// 设置日志
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	mw := &MainWindow{
		sessions:       make([]string, 0),
		selectedImages: make([]string, 0),
		feeds:          make([]map[string]interface{}, 0),
		searchResults:  make([]map[string]interface{}, 0),
	}

	// 创建主窗口
	if err := MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "小红书管理平台",
		Size:     Size{Width: 1000, Height: 700},
		Layout:   VBox{},
		Children: []Widget{
			// 顶部工具栏
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "会话:"},
					ComboBox{
						AssignTo:      &mw.sessionCombo,
						Model:         mw.sessions,
						Editable:      false,
						OnCurrentIndexChanged: mw.onSessionChanged,
					},
					PushButton{
						Text:      "刷新会话",
						OnClicked: mw.loadSessions,
					},
					CheckBox{
						AssignTo:  &mw.headlessCheck,
						Text:      "无头模式",
						Checked:   true,
						OnCheckedChanged: mw.onHeadlessChanged,
					},
				},
			},
			
			// 标签页
			TabWidget{
				Pages: []TabPage{
					{
						Title:  "发布内容",
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "标题:"},
							LineEdit{
								AssignTo: &mw.titleEdit,
								Text:     "",
							},
							Label{Text: "内容:"},
							TextEdit{
								AssignTo: &mw.contentEdit,
								Text:     "",
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									Label{Text: "图片:"},
									PushButton{
										Text:      "选择图片",
										OnClicked: mw.selectImages,
									},
									PushButton{
										Text:      "清空图片",
										OnClicked: mw.clearImages,
									},
								},
							},
							ListBox{
								AssignTo: &mw.imageList,
								Model:    &StringListModel{items: mw.selectedImages},
							},
							PushButton{
								Text:      "发布内容",
								OnClicked: mw.publishContent,
							},
							Label{
								AssignTo: &mw.statusLabel,
								Text:     "就绪",
							},
						},
					},
					{
						Title:  "内容管理",
						Layout: VBox{},
						Children: []Widget{
							PushButton{
								Text:      "刷新内容",
								OnClicked: mw.loadFeeds,
							},
							ListBox{
								AssignTo: &mw.contentList,
								Model:    &MapListModel{items: mw.feeds},
								OnCurrentIndexChanged: mw.onFeedSelected,
							},
						},
					},
					{
						Title:  "搜索",
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: HBox{},
								Children: []Widget{
									LineEdit{
										AssignTo: &mw.searchEdit,
										Text:     "",
									},
									PushButton{
										Text:      "搜索",
										OnClicked: mw.searchContent,
									},
								},
							},
							ListBox{
								AssignTo: &mw.searchList,
								Model:    &MapListModel{items: mw.searchResults},
								OnCurrentIndexChanged: mw.onSearchResultSelected,
							},
						},
					},
					{
						Title:  "日志",
						Layout: VBox{},
						Children: []Widget{
							PushButton{
								Text:      "清空日志",
								OnClicked: mw.clearLog,
							},
							TextEdit{
								AssignTo: &mw.logText,
								ReadOnly: true,
								Text:     "",
							},
						},
					},
				},
			},
		},
	}.Create(); err != nil {
		logrus.Fatalf("创建主窗口失败: %v", err)
	}

	// 初始化
	mw.loadSessions()
	mw.addLog("小红书管理平台启动")

	// 显示窗口并运行
	mw.Run()
}

func (mw *MainWindow) onSessionChanged() {
	if mw.sessionCombo.CurrentIndex() >= 0 {
		session := mw.sessions[mw.sessionCombo.CurrentIndex()]
		mw.addLog(fmt.Sprintf("已选择会话: %s", session))
	}
}

func (mw *MainWindow) onHeadlessChanged() {
	headless := mw.headlessCheck.Checked()
	mw.addLog(fmt.Sprintf("无头模式: %v", headless))
}

func (mw *MainWindow) selectImages() {
	dlg := new(walk.FileDialog)
	dlg.Title = "选择图片文件"
	dlg.Filter = "图片文件 (*.jpg;*.jpeg;*.png;*.gif)|*.jpg;*.jpeg;*.png;*.gif|所有文件 (*.*)|*.*"
	dlg.FilterIndex = 1

	if ok, err := dlg.ShowOpen(mw); err != nil {
		mw.addLog(fmt.Sprintf("选择图片失败: %v", err))
		return
	} else if !ok {
		return
	}

	mw.selectedImages = append(mw.selectedImages, dlg.FilePath)
	mw.imageList.SetModel(&StringListModel{items: mw.selectedImages})
	mw.addLog(fmt.Sprintf("已选择图片: %s", dlg.FilePath))
}

func (mw *MainWindow) clearImages() {
	mw.selectedImages = make([]string, 0)
	mw.imageList.SetModel(&StringListModel{items: mw.selectedImages})
	mw.addLog("已清空图片列表")
}

func (mw *MainWindow) loadSessions() {
	// 模拟加载会话
	mw.sessions = []string{"session-001", "session-002", "session-003"}
	mw.sessionCombo.SetModel(mw.sessions)
	mw.addLog("已加载会话列表")
}

func (mw *MainWindow) loadFeeds() {
	// 模拟加载内容
	mw.feeds = []map[string]interface{}{
		{"title": "测试内容1", "content": "这是测试内容1"},
		{"title": "测试内容2", "content": "这是测试内容2"},
	}
	mw.contentList.SetModel(&MapListModel{items: mw.feeds})
	mw.addLog("已加载内容列表")
}

func (mw *MainWindow) onFeedSelected() {
	if mw.contentList.CurrentIndex() >= 0 {
		feed := mw.feeds[mw.contentList.CurrentIndex()]
		mw.addLog(fmt.Sprintf("选择内容: %s", feed["title"]))
	}
}

func (mw *MainWindow) searchContent() {
	keyword := mw.searchEdit.Text()
	if keyword == "" {
		walk.MsgBox(mw, "提示", "请输入搜索关键词", walk.MsgBoxOK)
		return
	}

	// 模拟搜索
	mw.searchResults = []map[string]interface{}{
		{"title": "搜索结果1: " + keyword, "content": "这是搜索结果1"},
		{"title": "搜索结果2: " + keyword, "content": "这是搜索结果2"},
	}
	mw.searchList.SetModel(&MapListModel{items: mw.searchResults})
	mw.addLog(fmt.Sprintf("搜索关键词: %s", keyword))
}

func (mw *MainWindow) onSearchResultSelected() {
	if mw.searchList.CurrentIndex() >= 0 {
		result := mw.searchResults[mw.searchList.CurrentIndex()]
		mw.addLog(fmt.Sprintf("选择搜索结果: %s", result["title"]))
	}
}

func (mw *MainWindow) publishContent() {
	title := mw.titleEdit.Text()
	content := mw.contentEdit.Text()
	
	if title == "" || content == "" {
		walk.MsgBox(mw, "提示", "请输入标题和内容", walk.MsgBoxOK)
		return
	}

	if mw.sessionCombo.CurrentIndex() < 0 {
		walk.MsgBox(mw, "提示", "请选择会话", walk.MsgBoxOK)
		return
	}

	session := mw.sessions[mw.sessionCombo.CurrentIndex()]
	
	mw.statusLabel.SetText("正在发布...")
	mw.addLog(fmt.Sprintf("发布内容: %s", title))

	// 模拟发布过程
	go func() {
		time.Sleep(2 * time.Second)
		mw.statusLabel.SetText("发布成功")
		mw.addLog("内容发布成功")
		
		// 清空表单
		mw.titleEdit.SetText("")
		mw.contentEdit.SetText("")
		mw.clearImages()
	}()
}

func (mw *MainWindow) clearLog() {
	mw.logText.SetText("")
}

func (mw *MainWindow) addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	currentText := mw.logText.Text()
	mw.logText.SetText(currentText + logMessage)
}

// 实现 walk.ListModel 接口
type StringListModel struct {
	items []string
}

func (m *StringListModel) ItemCount() int {
	return len(m.items)
}

func (m *StringListModel) Value(index int) interface{} {
	return m.items[index]
}

// 实现 walk.ListModel 接口
type MapListModel struct {
	items []map[string]interface{}
}

func (m *MapListModel) ItemCount() int {
	return len(m.items)
}

func (m *MapListModel) Value(index int) interface{} {
	if index < len(m.items) {
		return m.items[index]["title"]
	}
	return ""
}
