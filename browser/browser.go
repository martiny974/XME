package browser

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/headless_browser"
)

// NewBrowser 创建新的浏览器实例（兼容旧接口）
func NewBrowser(headless bool) *headless_browser.Browser {
	// 获取当前会话ID
	sessionID := os.Getenv("MCP_SESSION_ID")
	if sessionID == "" {
		sessionID = "default"
	}

	// 使用浏览器管理器
	manager := GetManager()
	
	// 只有在传入的headless参数与当前设置不同时才更新
	currentHeadless := manager.IsHeadless()
	if currentHeadless != headless {
		logrus.Debugf("NewBrowser: 更新无头模式设置 %v -> %v", currentHeadless, headless)
		manager.SetHeadless(headless)
	}
	
	browser, err := manager.GetBrowser(sessionID)
	if err != nil {
		logrus.Errorf("获取浏览器实例失败: %v", err)
		// 如果获取失败，回退到直接创建
		return createDirectBrowser(headless)
	}

	return browser
}

// createDirectBrowser 直接创建浏览器实例（回退方案）
func createDirectBrowser(headless bool) *headless_browser.Browser {
	opts := []headless_browser.Option{
		headless_browser.WithHeadless(headless),
	}
	return headless_browser.New(opts...)
}
