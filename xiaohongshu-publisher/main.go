package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

func main() {
	var (
		headless bool
		title    string
		content  string
		images   string
		video    string
		tags     string
		check    bool
	)

	flag.BoolVar(&headless, "headless", true, "是否无头模式")
	flag.StringVar(&title, "title", "", "内容标题")
	flag.StringVar(&content, "content", "", "正文内容")
	flag.StringVar(&images, "images", "", "图片路径(逗号分隔,支持本地路径或HTTP链接)")
	flag.StringVar(&video, "video", "", "视频路径(仅支持本地)")
	flag.StringVar(&tags, "tags", "", "话题标签(逗号分隔)")
	flag.BoolVar(&check, "check", false, "检查登录状态")
	flag.Parse()

	configs.InitHeadless(headless)

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	browserInstance := browser.NewBrowser(headless)
	defer browserInstance.Close()

	page := browserInstance.NewPage()
	defer page.Close()

	loginAction := xiaohongshu.NewLogin(page)

	if check {
		if err := checkLoginStatus(ctx, loginAction); err != nil {
			logrus.Fatalf("检查登录状态失败: %v", err)
		}
		return
	}

	if video != "" {
		if err := publishVideo(ctx, page, title, content, video, parseTags(tags)); err != nil {
			logrus.Fatalf("发布视频失败: %v", err)
		}
		logrus.Info("视频发布成功!")
		return
	}

	if images != "" {
		if err := publishImages(ctx, page, title, content, parseImages(images), parseTags(tags)); err != nil {
			logrus.Fatalf("发布图文失败: %v", err)
		}
		logrus.Info("图文发布成功!")
		return
	}

	logrus.Error("请指定要执行的操作: -check (检查登录), -images (发布图文), 或 -video (发布视频)")
	flag.Usage()
}

func checkLoginStatus(ctx context.Context, loginAction *xiaohongshu.LoginAction) error {
	logrus.Info("检查登录状态...")
	isLoggedIn, err := loginAction.CheckLoginStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "检查登录状态失败")
	}

	if isLoggedIn {
		logrus.Info("✓ 已登录")
	} else {
		logrus.Warn("✗ 未登录")
		logrus.Info("请先运行登录程序进行扫码登录")
	}
	return nil
}

func publishImages(ctx context.Context, page *rod.Page, title, content string, imagePaths []string, tags []string) error {
	logrus.Info("准备发布图文...")

	if len(title) > 20 {
		logrus.Warnf("标题超过20字,将被截断")
		title = title[:20]
	}
	if len(content) > 1000 {
		logrus.Warnf("正文超过1000字,将被截断")
		content = content[:1000]
	}

	publishAction, err := xiaohongshu.NewPublishImageAction(page)
	if err != nil {
		return errors.Wrap(err, "初始化发布页面失败")
	}

	publishContent := xiaohongshu.PublishImageContent{
		Title:        title,
		Content:      content,
		ImagePaths:   imagePaths,
		Tags:         tags,
		ScheduleTime: nil,
	}

	if err := publishAction.Publish(ctx, publishContent); err != nil {
		return errors.Wrap(err, "发布失败")
	}

	return nil
}

func publishVideo(ctx context.Context, page *rod.Page, title, content, videoPath string, tags []string) error {
	logrus.Info("准备发布视频...")

	if len(title) > 20 {
		logrus.Warnf("标题超过20字,将被截断")
		title = title[:20]
	}
	if len(content) > 1000 {
		logrus.Warnf("正文超过1000字,将被截断")
		content = content[:1000]
	}

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return errors.Wrapf(err, "视频文件不存在: %s", videoPath)
	}

	publishAction, err := xiaohongshu.NewPublishVideoAction(page)
	if err != nil {
		return errors.Wrap(err, "初始化发布页面失败")
	}

	publishContent := xiaohongshu.PublishVideoContent{
		Title:        title,
		Content:      content,
		VideoPath:    videoPath,
		Tags:         tags,
		ScheduleTime: nil,
	}

	if err := publishAction.PublishVideo(ctx, publishContent); err != nil {
		return errors.Wrap(err, "发布失败")
	}

	return nil
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
