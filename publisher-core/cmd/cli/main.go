// Package main 提供命令行工具入�?
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	publisher "publisher-core/interfaces"
	"publisher-core/adapters"
	"publisher-core/storage"
	"publisher-core/task"
	"github.com/sirupsen/logrus"
)

var (
	platform  string
	headless  bool
	login     bool
	check     bool
	title     string
	content   string
	images    string
	video     string
	tags      string
	async     bool
	taskID    string
	status    bool
	list      bool
	cookieDir string
	debug     bool
)

func init() {
	flag.StringVar(&platform, "platform", "", "平台: douyin(抖音), toutiao(今日头条), xiaohongshu(小红�?")
	flag.BoolVar(&headless, "headless", true, "无头模式")
	flag.BoolVar(&login, "login", false, "登录")
	flag.BoolVar(&check, "check", false, "检查登录状�?)
	flag.StringVar(&title, "title", "", "标题")
	flag.StringVar(&content, "content", "", "正文内容")
	flag.StringVar(&images, "images", "", "图片路径(逗号分隔)")
	flag.StringVar(&video, "video", "", "视频路径")
	flag.StringVar(&tags, "tags", "", "话题标签(逗号分隔)")
	flag.BoolVar(&async, "async", false, "异步发布")
	flag.StringVar(&taskID, "task-id", "", "任务ID(用于查询状�?")
	flag.BoolVar(&status, "status", false, "查询任务状�?)
	flag.BoolVar(&list, "list", false, "列出任务")
	flag.StringVar(&cookieDir, "cookie-dir", "./cookies", "Cookie存储目录")
	flag.BoolVar(&debug, "debug", false, "调试模式")
}

func main() {
	flag.Parse()

	setupLogger()

	if platform == "" && !list && taskID == "" {
		printUsage()
		os.Exit(1)
	}

	// 创建发布�?
	factory := adapters.DefaultFactory()

	// 任务管理�?
	taskMgr := task.NewTaskManager(task.NewMemoryStorage())

	// 存储服务
	store, _ := storage.NewLocalStorage("./uploads", "")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 列出任务
	if list {
		listTasks(taskMgr)
		return
	}

	// 查询任务状�?
	if status && taskID != "" {
		queryTaskStatus(taskMgr, taskID)
		return
	}

	// 创建平台发布�?
	pub, err := factory.Create(platform, publisher.WithHeadless(headless), publisher.WithCookieDir(cookieDir))
	if err != nil {
		logrus.Fatalf("创建发布器失�? %v", err)
	}
	defer pub.Close()

	// 登录
	if login {
		doLogin(ctx, pub)
		return
	}

	// 检查登录状�?
	if check {
		doCheckLogin(ctx, pub)
		return
	}

	// 发布内容
	if images != "" || video != "" {
		doPublish(ctx, pub, taskMgr, store)
		return
	}

	printUsage()
}

func setupLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func printUsage() {
	fmt.Println("多平台内容发布工�?)
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  publisher [选项]")
	fmt.Println()
	fmt.Println("登录:")
	fmt.Println("  publisher -platform douyin -login")
	fmt.Println("  publisher -platform xiaohongshu -login")
	fmt.Println()
	fmt.Println("检查登录状�?")
	fmt.Println("  publisher -platform douyin -check")
	fmt.Println()
	fmt.Println("发布图文:")
	fmt.Println("  publisher -platform douyin -title \"标题\" -content \"正文\" -images \"img1.jpg,img2.jpg\" -tags \"美食,生活\"")
	fmt.Println()
	fmt.Println("发布视频:")
	fmt.Println("  publisher -platform douyin -title \"标题\" -content \"正文\" -video \"video.mp4\" -tags \"生活\"")
	fmt.Println()
	fmt.Println("异步发布:")
	fmt.Println("  publisher -platform douyin -title \"标题\" -content \"正文\" -video \"video.mp4\" -async")
	fmt.Println()
	fmt.Println("查询任务状�?")
	fmt.Println("  publisher -task-id <task_id> -status")
	fmt.Println()
	fmt.Println("列出任务:")
	fmt.Println("  publisher -list")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
}

func doLogin(ctx context.Context, pub publisher.Publisher) {
	logrus.Infof("开�?%s 登录...", pub.Platform())

	result, err := pub.Login(ctx)
	if err != nil {
		logrus.Fatalf("登录失败: %v", err)
	}

	if result.Success {
		logrus.Info("已登�?)
		return
	}

	if result.QrcodeURL != "" {
		fmt.Printf("请使�?%s App 扫码登录\n", pub.Platform())
		fmt.Printf("登录页面: %s\n", result.QrcodeURL)
	}

	logrus.Info("等待扫码登录...")

	if err := pub.WaitForLogin(ctx); err != nil {
		logrus.Fatalf("登录超时: %v", err)
	}

	logrus.Info("登录成功!")
}

func doCheckLogin(ctx context.Context, pub publisher.Publisher) {
	logrus.Info("检查登录状�?..")

	loggedIn, err := pub.CheckLoginStatus(ctx)
	if err != nil {
		logrus.Fatalf("检查失�? %v", err)
	}

	if loggedIn {
		logrus.Info("已登�?)
	} else {
		logrus.Warn("未登�?)
		logrus.Info("请先运行: publisher -platform <platform> -login")
	}
}

func doPublish(ctx context.Context, pub publisher.Publisher, taskMgr *task.TaskManager, store storage.Storage) {
	// 构建内容
	content := &publisher.Content{
		Title: title,
		Body:  content,
		Tags:  parseTags(tags),
	}

	if images != "" {
		content.Type = publisher.ContentTypeImages
		content.ImagePaths = parseImages(images)
	}

	if video != "" {
		content.Type = publisher.ContentTypeVideo
		content.VideoPath = video
	}

	// 验证内容
	if err := validateContent(pub, content); err != nil {
		logrus.Fatalf("内容验证失败: %v", err)
	}

	logrus.Infof("准备发布 %s 内容�?%s...", content.Type, pub.Platform())

	if async {
		// 异步发布
		taskID, err := pub.PublishAsync(ctx, content)
		if err != nil {
			logrus.Fatalf("创建发布任务失败: %v", err)
		}

		logrus.Info("异步发布任务已创�?)
		logrus.Infof("任务ID: %s", taskID)
		logrus.Info("使用以下命令查询状�?")
		fmt.Printf("  publisher -task-id %s -status\n", taskID)
		return
	}

	// 同步发布
	result, err := pub.Publish(ctx, content)
	if err != nil {
		logrus.Fatalf("发布失败: %v", err)
	}

	logrus.Infof("发布结果: %s", result.Status)
	if result.Error != "" {
		logrus.Errorf("错误: %s", result.Error)
	} else {
		logrus.Info("发布成功!")
		if result.PostURL != "" {
			logrus.Infof("访问链接: %s", result.PostURL)
		}
	}
}

func validateContent(pub publisher.Publisher, content *publisher.Content) error {
	if content.Title == "" {
		return fmt.Errorf("标题不能为空")
	}

	if content.Type == publisher.ContentTypeImages && len(content.ImagePaths) == 0 {
		return fmt.Errorf("图文内容必须包含图片")
	}

	if content.Type == publisher.ContentTypeVideo && content.VideoPath == "" {
		return fmt.Errorf("视频内容必须包含视频")
	}

	return nil
}

func listTasks(taskMgr *task.TaskManager) {
	tasks, err := taskMgr.ListTasks(task.TaskFilter{Limit: 20})
	if err != nil {
		logrus.Fatalf("获取任务列表失败: %v", err)
	}

	if len(tasks) == 0 {
		logrus.Info("暂无任务")
		return
	}

	fmt.Println("任务列表:")
	fmt.Println("----------------------------------------")
	for _, t := range tasks {
		fmt.Printf("ID: %s\n", t.ID)
		fmt.Printf("  类型: %s | 平台: %s | 状�? %s\n", t.Type, t.Platform, t.Status)
		fmt.Printf("  创建时间: %s\n", t.CreatedAt.Format(time.RFC3339))
		if t.Error != "" {
			fmt.Printf("  错误: %s\n", t.Error)
		}
		fmt.Println("----------------------------------------")
	}
}

func queryTaskStatus(taskMgr *task.TaskManager, id string) {
	task, err := taskMgr.GetTask(id)
	if err != nil {
		logrus.Fatalf("获取任务失败: %v", err)
	}

	data, _ := json.MarshalIndent(task, "", "  ")
	fmt.Println(string(data))
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
