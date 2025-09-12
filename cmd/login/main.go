package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

func main() {

	// 登录的时候，需要界面，所以不能无头模式
	// 关闭 go-rod 的 leakless，避免被 Windows Defender 误杀
	_ = os.Setenv("ROD_LAUNCH_LEAKLESS", "0")

	b := browser.NewBrowser(false)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := xiaohongshu.NewLogin(page)

	status, err := action.CheckLoginStatus(context.Background())
	if err != nil {
		logrus.Fatalf("failed to check login status: %v", err)
	}

	logrus.Infof("当前登录状态: %v", status)

	// 无论是否已登录，都保存一次 cookies，确保 ./cookies 目录与文件创建成功
	if err := saveCookies(page); err != nil {
		logrus.Warnf("保存 cookies 失败（将继续流程）：%v", err)
	}

	if status {
		logrus.Info("已登录，已写入 cookies 文件。")
		return
	}

	// 开始登录流程
	logrus.Info("开始登录流程...")
	if err = action.Login(context.Background()); err != nil {
		logrus.Fatalf("登录失败: %v", err)
	} else {
		if err := saveCookies(page); err != nil {
			logrus.Fatalf("failed to save cookies: %v", err)
		}
	}

	// 再次检查登录状态确认成功
	status, err = action.CheckLoginStatus(context.Background())
	if err != nil {
		logrus.Fatalf("failed to check login status after login: %v", err)
	}

	if status {
		logrus.Info("登录成功！")
	} else {
		logrus.Error("登录流程完成但仍未登录")
	}

}

func saveCookies(page *rod.Page) error {
	cks, err := page.Browser().GetCookies()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cks)
	if err != nil {
		return err
	}

	// 自动命名保存路径（若未指定 MCP_SESSION_ID，则使用 session-001/002...）
	path := cookies.GetCookiePathForSaving()
	cookieLoader := cookies.NewLoadCookie(path)
	return cookieLoader.SaveCookies(data)
}
