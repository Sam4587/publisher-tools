package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/monkeycode/douyin-toutiao-mcp/browser"
	"github.com/monkeycode/douyin-toutiao-mcp/cookies"
	"github.com/monkeycode/douyin-toutiao-mcp/douyin"
	"github.com/monkeycode/douyin-toutiao-mcp/toutiao"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		platform string
		headless bool
		title    string
		content  string
		images   string
		video    string
		tags     string
		check    bool
		login    bool
	)

	flag.StringVar(&platform, "platform", "douyin", "平台选择: douyin(抖音) 或 toutiao(今日头条)")
	flag.BoolVar(&headless, "headless", true, "是否无头模式")
	flag.StringVar(&title, "title", "", "内容标题")
	flag.StringVar(&content, "content", "", "正文内容")
	flag.StringVar(&images, "images", "", "图片路径(逗号分隔,支持本地路径)")
	flag.StringVar(&video, "video", "", "视频路径(仅支持本地)")
	flag.StringVar(&tags, "tags", "", "话题标签(逗号分隔)")
	flag.BoolVar(&check, "check", false, "检查登录状态")
	flag.BoolVar(&login, "login", false, "登录")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if err := cookies.InitCookieDir(); err != nil {
		logrus.Fatalf("初始化 cookie 目录失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	browserInstance := browser.NewBrowser(headless)
	defer browserInstance.Close()

	page := browserInstance.MustPage()
	defer page.Close()

	switch platform {
	case "douyin":
		handleDouyin(ctx, page, check, login, title, content, images, video, tags)
	case "toutiao":
		handleToutiao(ctx, page, check, login, title, content, images, video, tags)
	default:
		logrus.Errorf("不支持的平台: %s, 仅支持 douyin 或 toutiao", platform)
		os.Exit(1)
	}
}

func handleDouyin(ctx context.Context, page *rod.Page, check, login bool, title, content, images, video, tags string) {
	loginAction := douyin.NewLogin(page)

	if login {
		if err := performLogin(ctx, loginAction, "douyin"); err != nil {
			logrus.Fatalf("登录失败: %v", err)
		}
		return
	}

	if check {
		if err := checkLoginStatus(ctx, loginAction); err != nil {
			logrus.Fatalf("检查登录状态失败: %v", err)
		}
		return
	}

	if video != "" {
		publishAction, err := douyin.NewPublishAction(page)
		if err != nil {
			logrus.Fatalf("初始化发布页面失败: %v", err)
		}
		if err := publishVideo(ctx, publishAction, title, content, video, parseTags(tags)); err != nil {
			logrus.Fatalf("发布视频失败: %v", err)
		}
		logrus.Info("视频发布成功!")
		return
	}

	if images != "" {
		publishAction, err := douyin.NewPublishAction(page)
		if err != nil {
			logrus.Fatalf("初始化发布页面失败: %v", err)
		}
		if err := publishImages(ctx, publishAction, title, content, parseImages(images), parseTags(tags)); err != nil {
			logrus.Fatalf("发布图文失败: %v", err)
		}
		logrus.Info("图文发布成功!")
		return
	}

	printUsage("douyin")
}

func handleToutiao(ctx context.Context, page *rod.Page, check, login bool, title, content, images, video, tags string) {
	loginAction := toutiao.NewLogin(page)

	if login {
		if err := performLogin(ctx, loginAction, "toutiao"); err != nil {
			logrus.Fatalf("登录失败: %v", err)
		}
		return
	}

	if check {
		if err := checkLoginStatus(ctx, loginAction); err != nil {
			logrus.Fatalf("检查登录状态失败: %v", err)
		}
		return
	}

	if video != "" {
		publishAction, err := toutiao.NewPublishAction(page)
		if err != nil {
			logrus.Fatalf("初始化发布页面失败: %v", err)
		}
		if err := publishVideo(ctx, publishAction, title, content, video, parseTags(tags)); err != nil {
			logrus.Fatalf("发布视频失败: %v", err)
		}
		logrus.Info("视频发布成功!")
		return
	}

	if images != "" {
		publishAction, err := toutiao.NewPublishAction(page)
		if err != nil {
			logrus.Fatalf("初始化发布页面失败: %v", err)
		}
		if err := publishImages(ctx, publishAction, title, content, parseImages(images), parseTags(tags)); err != nil {
			logrus.Fatalf("发布图文失败: %v", err)
		}
		logrus.Info("图文发布成功!")
		return
	}

	printUsage("toutiao")
}

func performLogin(ctx context.Context, loginAction interface{}, platform string) error {
	logrus.Infof("开始 %s 登录...", platform)

	qrcodeURL, isLoggedIn, err := fetchQrcode(ctx, loginAction, platform)
	if err != nil {
		return err
	}

	if isLoggedIn {
		logrus.Info("✓ 已登录")
		return nil
	}

	if qrcodeURL != "" {
		fmt.Printf("请使用 %s App 扫码登录\n", platform)
		fmt.Printf("二维码: %s\n", qrcodeURL)
	}

	logrus.Info("等待扫码登录...")
	success := waitForLogin(ctx, loginAction, platform)
	if !success {
		return errors.New("登录超时")
	}

	cookiesData, err := extractCookies(ctx, loginAction, platform)
	if err != nil {
		return err
	}

	if err := cookies.SaveCookies(cookiesData, platform); err != nil {
		return err
	}

	logrus.Info("✓ 登录成功!")
	return nil
}

func checkLoginStatus(ctx context.Context, loginAction interface{}) error {
	logrus.Info("检查登录状态...")
	isLoggedIn, err := checkStatus(ctx, loginAction)
	if err != nil {
		return errors.Wrap(err, "检查登录状态失败")
	}

	if isLoggedIn {
		logrus.Info("✓ 已登录")
	} else {
		logrus.Warn("✗ 未登录")
		logrus.Info("请先运行登录: -login")
	}
	return nil
}

func fetchQrcode(ctx context.Context, loginAction interface{}, platform string) (string, bool, error) {
	switch action := loginAction.(type) {
	case *douyin.LoginAction:
		return action.FetchQrcodeImage(ctx)
	case *toutiao.LoginAction:
		return action.FetchQrcodeImage(ctx)
	default:
		return "", false, errors.New("无效的登录操作")
	}
}

func waitForLogin(ctx context.Context, loginAction interface{}, platform string) bool {
	switch action := loginAction.(type) {
	case *douyin.LoginAction:
		return action.WaitForLogin(ctx)
	case *toutiao.LoginAction:
		return action.WaitForLogin(ctx)
	default:
		return false
	}
}

func extractCookies(ctx context.Context, loginAction interface{}, platform string) (map[string]string, error) {
	switch action := loginAction.(type) {
	case *douyin.LoginAction:
		return action.ExtractCookies(ctx)
	case *toutiao.LoginAction:
		return action.ExtractCookies(ctx)
	default:
		return nil, errors.New("无效的登录操作")
	}
}

func checkStatus(ctx context.Context, loginAction interface{}) (bool, error) {
	switch action := loginAction.(type) {
	case *douyin.LoginAction:
		return action.CheckLoginStatus(ctx)
	case *toutiao.LoginAction:
		return action.CheckLoginStatus(ctx)
	default:
		return false, errors.New("无效的登录操作")
	}
}

func publishVideo(ctx context.Context, publishAction interface{}, title, content, videoPath string, tags []string) error {
	if len(title) > 30 {
		logrus.Warnf("标题超过30字,将被截断")
		title = title[:30]
	}
	if len(content) > 2000 {
		logrus.Warnf("正文超过2000字,将被截断")
		content = content[:2000]
	}

	switch action := publishAction.(type) {
	case *douyin.PublishAction:
		publishContent := douyin.PublishVideoContent{
			Title:        title,
			Content:      content,
			VideoPath:    videoPath,
			Tags:         tags,
			ScheduleTime: nil,
		}
		return action.PublishVideo(ctx, publishContent)
	case *toutiao.PublishAction:
		publishContent := toutiao.PublishVideoContent{
			Title:        title,
			Content:      content,
			VideoPath:    videoPath,
			Tags:         tags,
			ScheduleTime: nil,
		}
		return action.PublishVideo(ctx, publishContent)
	default:
		return errors.New("无效的发布操作")
	}
}

func publishImages(ctx context.Context, publishAction interface{}, title, content string, imagePaths []string, tags []string) error {
	if len(title) > 30 {
		logrus.Warnf("标题超过30字,将被截断")
		title = title[:30]
	}
	if len(content) > 2000 {
		logrus.Warnf("正文超过2000字,将被截断")
		content = content[:2000]
	}

	switch action := publishAction.(type) {
	case *douyin.PublishAction:
		publishContent := douyin.PublishImageContent{
			Title:        title,
			Content:      content,
			ImagePaths:   imagePaths,
			Tags:         tags,
			ScheduleTime: nil,
		}
		return action.PublishImages(ctx, publishContent)
	case *toutiao.PublishAction:
		publishContent := toutiao.PublishImageContent{
			Title:        title,
			Content:      content,
			ImagePaths:   imagePaths,
			Tags:         tags,
			ScheduleTime: nil,
		}
		return action.PublishImages(ctx, publishContent)
	default:
		return errors.New("无效的发布操作")
	}
}

func parseImages(input string) []string {
	if input == "" {
		return nil
	}
	var result []string
	for _, s := range strings.Split(input, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func parseTags(input string) []string {
	if input == "" {
		return nil
	}
	var result []string
	for _, tag := range strings.Split(input, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

func printUsage(platform string) {
	logrus.Error("请指定要执行的操作:")
	logrus.Info("  -login        登录")
	logrus.Info("  -check         检查登录状态")
	logrus.Info("  -images        发布图文")
	logrus.Info("  -video         发布视频")
	logrus.Info("  -title         标题")
	logrus.Info("  -content       正文")
	logrus.Info("  -tags          标签")
	logrus.Info(fmt.Sprintf("示例: -platform %s -login", platform))
}
