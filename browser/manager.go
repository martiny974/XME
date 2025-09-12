package browser

import (
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/headless_browser"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

// BrowserManager 浏览器实例管理器
type BrowserManager struct {
	browsers map[string]*headless_browser.Browser // sessionID -> browser
	mutex    sync.RWMutex
	headless bool
}

var (
	globalManager *BrowserManager
	once          sync.Once
)

// GetManager 获取全局浏览器管理器实例
func GetManager() *BrowserManager {
	once.Do(func() {
		globalManager = &BrowserManager{
			browsers: make(map[string]*headless_browser.Browser),
			headless: false, // 默认有头模式
		}

		// 禁用 go-rod 的 leakless，避免被杀软/Defender 拦截
		_ = os.Setenv("ROD_LAUNCH_LEAKLESS", "0")

		// 检查环境变量
		if v := os.Getenv("MCP_HEADLESS"); v != "" {
			if b, err := parseBool(v); err == nil {
				globalManager.headless = b
			}
		}

		logrus.Infof("浏览器管理器初始化完成，无头模式: %v", globalManager.headless)
	})
	return globalManager
}

// GetBrowser 获取或创建浏览器实例
func (m *BrowserManager) GetBrowser(sessionID string) (*headless_browser.Browser, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果已存在，直接返回
	if browser, exists := m.browsers[sessionID]; exists {
		logrus.Debugf("复用现有浏览器实例，会话: %s", sessionID)
		return browser, nil
	}

	// 创建新的浏览器实例
	logrus.Infof("创建新的浏览器实例，会话: %s，无头模式: %v", sessionID, m.headless)
	
	opts := []headless_browser.Option{
		headless_browser.WithHeadless(m.headless),
	}

	// 加载 cookies
	cookiePath := cookies.ResolveCookiePath(sessionID)
	cookieLoader := cookies.NewLoadCookie(cookiePath)

	if data, err := cookieLoader.LoadCookies(); err == nil {
		opts = append(opts, headless_browser.WithCookies(string(data)))
		logrus.Debugf("加载cookies成功，会话: %s", sessionID)
	} else {
		logrus.Warnf("加载cookies失败，会话: %s，错误: %v", sessionID, err)
	}

	browser := headless_browser.New(opts...)
	m.browsers[sessionID] = browser

	return browser, nil
}

// CloseBrowser 关闭指定会话的浏览器
func (m *BrowserManager) CloseBrowser(sessionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if browser, exists := m.browsers[sessionID]; exists {
		logrus.Infof("关闭浏览器实例，会话: %s", sessionID)
		browser.Close()
		delete(m.browsers, sessionID)
	}
}

// CloseAll 关闭所有浏览器实例
func (m *BrowserManager) CloseAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logrus.Info("关闭所有浏览器实例")
	for sessionID, browser := range m.browsers {
		logrus.Debugf("关闭浏览器实例，会话: %s", sessionID)
		browser.Close()
	}
	m.browsers = make(map[string]*headless_browser.Browser)
}

// SetHeadless 设置无头模式
func (m *BrowserManager) SetHeadless(headless bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	oldHeadless := m.headless
	m.headless = headless
	
	if oldHeadless != headless {
		logrus.Infof("无头模式设置已更改: %v -> %v，关闭所有现有浏览器实例", oldHeadless, headless)
		
		// 关闭所有现有浏览器实例，强制重新创建
		for sessionID, browser := range m.browsers {
			logrus.Debugf("关闭浏览器实例以应用新的无头模式，会话: %s", sessionID)
			browser.Close()
		}
		// 清空浏览器映射，强制重新创建
		m.browsers = make(map[string]*headless_browser.Browser)
	}
}

// IsHeadless 获取当前无头模式设置
func (m *BrowserManager) IsHeadless() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.headless
}

// GetSessionCount 获取当前活跃的会话数量
func (m *BrowserManager) GetSessionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.browsers)
}

// CleanupInactiveSessions 清理不活跃的会话（可选功能）
func (m *BrowserManager) CleanupInactiveSessions(maxIdleTime time.Duration) {
	// 这里可以实现清理逻辑，比如检测长时间未使用的浏览器实例
	// 暂时不实现，因为需要跟踪最后使用时间
}

// parseBool 解析布尔值字符串
func parseBool(s string) (bool, error) {
	switch s {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, nil
	}
}
