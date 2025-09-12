package cookies

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Cookier interface {
	LoadCookies() ([]byte, error)
	SaveCookies(data []byte) error
}

type localCookie struct {
	path string
}

func NewLoadCookie(path string) Cookier {
	if path == "" {
		panic("path is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		panic(err)
	}

	return &localCookie{
		path: path,
	}
}

// LoadCookies 从文件中加载 cookies。
func (c *localCookie) LoadCookies() ([]byte, error) {

	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cookies from tmp file")
	}

	return data, nil
}

// SaveCookies 保存 cookies 到文件中。
func (c *localCookie) SaveCookies(data []byte) error {
	return os.WriteFile(c.path, data, 0644)
}

// GetCookiesFilePath 获取 cookies 文件路径（默认会话）。
func GetCookiesFilePath() string {
	return GetCookiesFilePathWithSession("")
}

// GetCookiesFilePathWithSession 根据会话ID获取 cookies 路径。
// 默认：程序运行目录下 cookies/{sessionID|default}.json
func GetCookiesFilePathWithSession(sessionID string) string {
	baseDir := getCookiesBaseDir()
	if sessionID == "" {
		// 支持从环境变量读取（用于服务端按请求选择会话）
		if v := os.Getenv("MCP_SESSION_ID"); v != "" {
			sessionID = v
		}
	}
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join(baseDir, sessionID+".json")
}

// GetCookiePathForSaving 在保存新 cookies 时决定文件路径：
// 1) 若 MCP_SESSION_ID 存在，则返回该会话路径
// 2) 否则自动生成新会话名（session-001, session-002, ...）并返回对应路径
func GetCookiePathForSaving() string {
	if v := os.Getenv("MCP_SESSION_ID"); strings.TrimSpace(v) != "" {
		return GetCookiesFilePathWithSession(strings.TrimSpace(v))
	}
	name := nextAutoSessionName("session")
	return GetCookiesFilePathWithSession(name)
}

// ResolveCookiePath 允许通过“前缀或完整文件名（不含.json）”解析到具体 cookies 文件。
// 规则：
// 1) 若存在 ./cookies/{name}.json 则直接使用
// 2) 否则在 ./cookies 下寻找以 {name} 开头且扩展名为 .json 的第一个文件
// 3) 若仍无，返回默认 ./cookies/default.json
func ResolveCookiePath(nameOrPrefix string) string {
	baseDir := getCookiesBaseDir()
	if nameOrPrefix == "" {
		return filepath.Join(baseDir, "default.json")
	}
	candidate := filepath.Join(baseDir, nameOrPrefix+".json")
	if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
		return candidate
	}
	entries, err := os.ReadDir(baseDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if filepath.Ext(name) == ".json" && hasPrefixInsensitive(name, nameOrPrefix) {
				return filepath.Join(baseDir, name)
			}
		}
	}
	return filepath.Join(baseDir, "default.json")
}

func hasPrefixInsensitive(file string, prefix string) bool {
	if prefix == "" {
		return false
	}
	// 比较不区分大小写，且仅比对不含后缀名的部分
	base := file
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}
	fb := strings.ToLower(base)
	pb := strings.ToLower(prefix)
	return strings.HasPrefix(fb, pb)
}

// ListSessions 列出本地已存在的会话ID（基于 cookies 文件名）。
func ListSessions() ([]string, error) {
	baseDir := getCookiesBaseDir()
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var sessions []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) == ".json" {
			sessions = append(sessions, name[:len(name)-len(".json")])
		}
	}
	return sessions, nil
}

// nextAutoSessionName 生成下一个可用会话名（按 session-001 递增）
func nextAutoSessionName(prefix string) string {
	if strings.TrimSpace(prefix) == "" {
		prefix = "session"
	}
	baseDir := getCookiesBaseDir()
	_ = os.MkdirAll(baseDir, 0755)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return prefix + "-001"
	}
	var nums []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix+"-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		mid := strings.TrimSuffix(strings.TrimPrefix(name, prefix+"-"), ".json")
		if n, err := strconv.Atoi(mid); err == nil {
			nums = append(nums, n)
		}
	}
	if len(nums) == 0 {
		return prefix + "-001"
	}
	sort.Ints(nums)
	next := nums[len(nums)-1] + 1
	return fmt.Sprintf("%s-%03d", prefix, next)
}

// getCookiesBaseDir 返回 cookies 存储的基础目录（程序运行目录下的 cookies/）
func getCookiesBaseDir() string {
	wd, err := os.Getwd()
	if err != nil || wd == "" {
		if exe, e := os.Executable(); e == nil {
			wd = filepath.Dir(exe)
		} else {
			return filepath.Join(os.TempDir(), "xiaohongshu-mcp", "cookies")
		}
	}
	return filepath.Join(wd, "cookies")
}
