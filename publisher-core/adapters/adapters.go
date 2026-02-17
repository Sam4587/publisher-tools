package adapters

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"publisher-core/browser"
	"publisher-core/cookies"
	publisher "publisher-core/interfaces"
	"publisher-core/storage"
	"publisher-core/task"
)

type BaseAdapter struct {
	mu         sync.Mutex
	browser    *browser.Browser
	cookieMgr  *cookies.Manager
	taskMgr    *task.TaskManager
	storage    storage.Storage

	platform   string
	loginURL   string
	publishURL string
	limits     publisher.ContentLimits
	cookieKeys []string
	domain     string

	headless  bool
	cookieDir string
}

func NewBaseAdapter(platform string, opts *publisher.Options) *BaseAdapter {
	if opts == nil {
		opts = publisher.DefaultOptions()
	}
	if opts.CookieDir == "" {
		opts.CookieDir = "./cookies"
	}

	return &BaseAdapter{
		platform:   platform,
		headless:   opts.Headless,
		cookieDir:  opts.CookieDir,
		cookieMgr:  cookies.NewManager(opts.CookieDir),
		taskMgr:    task.NewTaskManager(task.NewMemoryStorage()),
		storage:    nil,
	}
}

func (a *BaseAdapter) Platform() string {
	return a.platform
}

func (a *BaseAdapter) initBrowser() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.browser == nil {
		a.browser = browser.NewBrowser(&browser.Config{
			Headless: a.headless,
		})
	}
	return nil
}

func (a *BaseAdapter) Login(ctx context.Context) (*publisher.LoginResult, error) {
	if err := a.initBrowser(); err != nil {
		return nil, err
	}

	loggedIn, err := a.CheckLoginStatus(ctx)
	if err != nil {
		logrus.Warnf("[%s] Check login status failed: %v", a.platform, err)
	}

	if loggedIn {
		logrus.Infof("[%s] Already logged in", a.platform)
		return &publisher.LoginResult{Success: true}, nil
	}

	page := a.browser.MustPage()
	defer page.Close()

	helper := browser.NewPageHelper(page)
	if err := helper.Navigate(a.loginURL); err != nil {
		return nil, errors.Wrap(err, "navigate to login page failed")
	}

	time.Sleep(2 * time.Second)

	qrcodeURL, err := a.getQrcodeURL(page)
	if err != nil {
		logrus.Warnf("[%s] Get qrcode failed: %v", a.platform, err)
	}

	return &publisher.LoginResult{
		Success:   false,
		QrcodeURL: qrcodeURL,
	}, nil
}

func (a *BaseAdapter) WaitForLogin(ctx context.Context) error {
	if err := a.initBrowser(); err != nil {
		return err
	}

	page := a.browser.MustPage()
	defer page.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	loginCheckSelector := a.getLoginCheckSelector()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			has, _, _ := page.Has(loginCheckSelector)
			if has {
				cookiesData, err := page.Cookies([]string{})
				if err != nil {
					return errors.Wrap(err, "get cookies failed")
				}

				keyCookies := cookies.ExtractCookies(cookiesData, a.cookieKeys)
				if len(keyCookies) == 0 {
					logrus.Warnf("[%s] Key cookies not found", a.platform)
					continue
				}

				if err := a.cookieMgr.Save(ctx, a.platform, cookiesData); err != nil {
					return errors.Wrap(err, "save cookies failed")
				}

				logrus.Infof("[%s] Login success, saved %d cookies", a.platform, len(keyCookies))
				return nil
			}
		}
	}
}

func (a *BaseAdapter) CheckLoginStatus(ctx context.Context) (bool, error) {
	exists, err := a.cookieMgr.Exists(ctx, a.platform)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	return true, nil
}

func (a *BaseAdapter) Logout(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	logrus.Infof("[%s] Executing logout", a.platform)

	if err := a.cookieMgr.Delete(ctx, a.platform); err != nil {
		logrus.Warnf("[%s] Delete cookies failed: %v", a.platform, err)
		return err
	}

	logrus.Infof("[%s] Logout success", a.platform)
	return nil
}

func (a *BaseAdapter) Publish(ctx context.Context, content *publisher.Content) (*publisher.PublishResult, error) {
	if err := a.validateContent(content); err != nil {
		return nil, err
	}

	taskID := fmt.Sprintf("%s_%d", a.platform, time.Now().UnixNano())
	result := &publisher.PublishResult{
		TaskID:    taskID,
		Status:    publisher.StatusProcessing,
		Platform:  a.platform,
		CreatedAt: time.Now(),
	}

	err := a.doPublish(ctx, content)
	if err != nil {
		result.Status = publisher.StatusFailed
		result.Error = err.Error()
		return result, err
	}

	result.Status = publisher.StatusSuccess
	now := time.Now()
	result.FinishedAt = &now

	return result, nil
}

func (a *BaseAdapter) PublishAsync(ctx context.Context, content *publisher.Content) (string, error) {
	if err := a.validateContent(content); err != nil {
		return "", err
	}

	payload := map[string]interface{}{
		"title":  content.Title,
		"body":   content.Body,
		"type":   content.Type,
		"images": content.ImagePaths,
		"video":  content.VideoPath,
		"tags":   content.Tags,
	}

	t, err := a.taskMgr.CreateTask("publish", a.platform, payload)
	if err != nil {
		return "", err
	}

	go a.taskMgr.Execute(context.Background(), t.ID)

	return t.ID, nil
}

func (a *BaseAdapter) QueryStatus(ctx context.Context, taskID string) (*publisher.PublishResult, error) {
	t, err := a.taskMgr.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	result := &publisher.PublishResult{
		TaskID:    t.ID,
		Platform:  a.platform,
		CreatedAt: t.CreatedAt,
	}

	switch t.Status {
	case task.TaskStatusPending:
		result.Status = publisher.StatusPending
	case task.TaskStatusRunning:
		result.Status = publisher.StatusProcessing
	case task.TaskStatusCompleted:
		result.Status = publisher.StatusSuccess
		if t.FinishedAt != nil {
			result.FinishedAt = t.FinishedAt
		}
	case task.TaskStatusFailed:
		result.Status = publisher.StatusFailed
		result.Error = t.Error
	case task.TaskStatusCancelled:
		result.Status = publisher.StatusFailed
		result.Error = "task cancelled"
	}

	return result, nil
}

func (a *BaseAdapter) Cancel(ctx context.Context, taskID string) error {
	return a.taskMgr.Cancel(taskID)
}

func (a *BaseAdapter) Close() error {
	if a.browser != nil {
		return a.browser.Close()
	}
	return nil
}

func (a *BaseAdapter) GetLimits() publisher.ContentLimits {
	return a.limits
}

func (a *BaseAdapter) validateContent(content *publisher.Content) error {
	if content == nil {
		return fmt.Errorf("content cannot be empty")
	}

	if len(content.Title) > a.limits.TitleMaxLength {
		return fmt.Errorf("title exceeds max length %d", a.limits.TitleMaxLength)
	}

	if len(content.Body) > a.limits.BodyMaxLength {
		return fmt.Errorf("body exceeds max length %d", a.limits.BodyMaxLength)
	}

	if content.Type == publisher.ContentTypeImages && len(content.ImagePaths) == 0 {
		return fmt.Errorf("image content must include images")
	}

	if content.Type == publisher.ContentTypeVideo && content.VideoPath == "" {
		return fmt.Errorf("video content must include video")
	}

	return nil
}

func (a *BaseAdapter) getQrcodeURL(page *rod.Page) (string, error) {
	return "", nil
}

func (a *BaseAdapter) getLoginCheckSelector() string {
	return ""
}

func (a *BaseAdapter) doPublish(ctx context.Context, content *publisher.Content) error {
	return nil
}

type DouyinAdapter struct {
	BaseAdapter
}

func NewDouyinAdapter(opts *publisher.Options) *DouyinAdapter {
	base := NewBaseAdapter("douyin", opts)
	base.loginURL = "https://creator.douyin.com/creator-micro/content/publish"
	base.publishURL = "https://creator.douyin.com/creator-micro/content/publish"
	base.domain = ".douyin.com"
	base.cookieKeys = cookies.DouyinCookieKeys
	base.limits = publisher.ContentLimits{
		TitleMaxLength:      30,
		BodyMaxLength:       2000,
		MaxImages:           12,
		MaxVideoSize:        4 * 1024 * 1024 * 1024,
		MaxTags:             5,
		AllowedVideoFormats: []string{".mp4", ".mov", ".avi", ".mkv"},
		AllowedImageFormats: []string{".jpg", ".jpeg", ".png", ".webp"},
	}

	return &DouyinAdapter{BaseAdapter: *base}
}

func (a *DouyinAdapter) getQrcodeURL(page *rod.Page) (string, error) {
	has, elem, err := page.Has(".login-avatar")
	if err == nil && has {
		return "", nil
	}

	elem, err = page.Element(".qrcode-img")
	if err != nil {
		return "", errors.Wrap(err, "find qrcode element failed")
	}

	src, err := elem.Attribute("src")
	if err != nil || src == nil {
		return "", errors.New("get qrcode link failed")
	}

	return *src, nil
}

func (a *DouyinAdapter) getLoginCheckSelector() string {
	return ".login-avatar"
}

func (a *DouyinAdapter) doPublish(ctx context.Context, content *publisher.Content) error {
	if err := a.initBrowser(); err != nil {
		return err
	}

	cookieParams, err := a.cookieMgr.LoadAsProto(ctx, a.platform, a.domain)
	if err != nil {
		return errors.Wrap(err, "load cookies failed")
	}

	page := a.browser.MustPage()
	defer page.Close()

	if len(cookieParams) > 0 {
		if err := page.SetCookies(cookieParams); err != nil {
			logrus.Warnf("[%s] Set cookies failed: %v", a.platform, err)
		}
	}

	helper := browser.NewPageHelper(page)

	if err := helper.Navigate(a.publishURL); err != nil {
		return errors.Wrap(err, "navigate to publish page failed")
	}

	time.Sleep(3 * time.Second)

	has, _, _ := page.Has(".login-avatar")
	if !has {
		return errors.New("not logged in, please login first")
	}

	if content.Type == publisher.ContentTypeVideo {
		if err := a.uploadVideo(page, content.VideoPath); err != nil {
			return errors.Wrap(err, "upload video failed")
		}
	} else {
		if err := a.uploadImages(page, content.ImagePaths); err != nil {
			return errors.Wrap(err, "upload images failed")
		}
	}

	if err := a.fillContent(page, content); err != nil {
		return errors.Wrap(err, "fill content failed")
	}

	if err := a.submitPublish(page); err != nil {
		return errors.Wrap(err, "publish failed")
	}

	logrus.Infof("[%s] Publish success", a.platform)
	return nil
}

func (a *DouyinAdapter) uploadVideo(page *rod.Page, videoPath string) error {
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("video file not found: %s", videoPath)
	}

	logrus.Infof("[%s] Uploading video: %s", a.platform, videoPath)

	fileInput, err := page.Element("input[type='file'][accept*='video']")
	if err != nil {
		return errors.Wrap(err, "find video upload input failed")
	}

	if err := fileInput.SetFiles([]string{videoPath}); err != nil {
		return errors.Wrap(err, "set video file failed")
	}

	logrus.Infof("[%s] Waiting for video upload...", a.platform)
	time.Sleep(5 * time.Second)

	return nil
}

func (a *DouyinAdapter) uploadImages(page *rod.Page, imagePaths []string) error {
	helper := browser.NewPageHelper(page)

	for i, imgPath := range imagePaths {
		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			return fmt.Errorf("image file not found: %s", imgPath)
		}

		logrus.Infof("[%s] Uploading image %d/%d: %s", a.platform, i+1, len(imagePaths), imgPath)

		fileInput, err := page.Element("input[type='file'][accept*='image']")
		if err != nil {
			return errors.Wrap(err, "find image upload input failed")
		}

		if err := fileInput.SetFiles([]string{imgPath}); err != nil {
			return errors.Wrap(err, "set image file failed")
		}

		helper.RandomDelay(1, 2)
	}

	return nil
}

func (a *DouyinAdapter) fillContent(page *rod.Page, content *publisher.Content) error {
	helper := browser.NewPageHelper(page)

	titleInput, err := page.Element("input[placeholder*='title']")
	if err == nil {
		if err := titleInput.Input(content.Title); err != nil {
			logrus.Warnf("[%s] Input title failed: %v", a.platform, err)
		}
		helper.RandomDelay(0.5, 1)
	}

	contentInput, err := page.Element("textarea[placeholder*='content']")
	if err == nil {
		if err := contentInput.Input(content.Body); err != nil {
			logrus.Warnf("[%s] Input body failed: %v", a.platform, err)
		}
		helper.RandomDelay(0.5, 1)
	}

	for _, tag := range content.Tags {
		tagInput, err := page.Element("input[placeholder*='topic']")
		if err != nil {
			logrus.Warnf("[%s] Find topic input failed: %v", a.platform, err)
			continue
		}

		tagInput.Input("#" + tag)
		time.Sleep(500 * time.Millisecond)
		helper.RandomDelay(0.3, 0.7)
	}

	return nil
}

func (a *DouyinAdapter) submitPublish(page *rod.Page) error {
	helper := browser.NewPageHelper(page)

	publishBtn, err := page.Element("button[type='submit']")
	if err != nil {
		return errors.Wrap(err, "find publish button failed")
	}

	vis, err := publishBtn.Visible()
	if err != nil || !vis {
		return errors.New("publish button not visible")
	}

	helper.RandomDelay(1, 2)

	if err := publishBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return errors.Wrap(err, "click publish button failed")
	}

	logrus.Infof("[%s] Clicked publish button, waiting...", a.platform)
	time.Sleep(5 * time.Second)

	return nil
}

type ToutiaoAdapter struct {
	DouyinAdapter
}

func NewToutiaoAdapter(opts *publisher.Options) *ToutiaoAdapter {
	base := NewBaseAdapter("toutiao", opts)
	base.loginURL = "https://mp.toutiao.com/"
	base.publishURL = "https://mp.toutiao.com/profile_v4/pub_article"
	base.domain = ".toutiao.com"
	base.cookieKeys = cookies.ToutiaoCookieKeys

	return &ToutiaoAdapter{DouyinAdapter: DouyinAdapter{BaseAdapter: *base}}
}

func (a *ToutiaoAdapter) getLoginCheckSelector() string {
	return ".user-avatar"
}

func (a *ToutiaoAdapter) getQrcodeURL(page *rod.Page) (string, error) {
	has, elem, err := page.Has(".user-avatar")
	if err == nil && has {
		return "", nil
	}

	elem, err = page.Element(".qrcode-img, .qr-code")
	if err != nil {
		return "", errors.Wrap(err, "find qrcode element failed")
	}

	src, err := elem.Attribute("src")
	if err != nil || src == nil {
		return "", errors.New("get qrcode link failed")
	}

	return *src, nil
}

type XiaohongshuAdapter struct {
	BaseAdapter
}

func NewXiaohongshuAdapter(opts *publisher.Options) *XiaohongshuAdapter {
	base := NewBaseAdapter("xiaohongshu", opts)
	base.loginURL = "https://creator.xiaohongshu.com/"
	base.publishURL = "https://creator.xiaohongshu.com/publish/publish"
	base.domain = ".xiaohongshu.com"
	base.cookieKeys = cookies.XiaohongshuCookieKeys
	base.limits = publisher.ContentLimits{
		TitleMaxLength:      20,
		BodyMaxLength:       1000,
		MaxImages:           18,
		MaxVideoSize:        500 * 1024 * 1024,
		MaxTags:             5,
		AllowedVideoFormats: []string{".mp4", ".mov"},
		AllowedImageFormats: []string{".jpg", ".jpeg", ".png", ".webp"},
	}

	return &XiaohongshuAdapter{BaseAdapter: *base}
}

func (a *XiaohongshuAdapter) getLoginCheckSelector() string {
	return ".avatar-wrapper, .user-info"
}

func (a *XiaohongshuAdapter) getQrcodeURL(page *rod.Page) (string, error) {
	has, _, err := page.Has(".avatar-wrapper, .user-info")
	if err == nil && has {
		return "", nil
	}

	elem, err := page.Element(".qrcode-img, img[class*='qrcode']")
	if err != nil {
		return "", errors.Wrap(err, "find qrcode element failed")
	}

	src, err := elem.Attribute("src")
	if err != nil || src == nil {
		return "", errors.New("get qrcode link failed")
	}

	return *src, nil
}

func (a *XiaohongshuAdapter) doPublish(ctx context.Context, content *publisher.Content) error {
	if err := a.initBrowser(); err != nil {
		return err
	}

	cookieParams, err := a.cookieMgr.LoadAsProto(ctx, a.platform, a.domain)
	if err != nil {
		return errors.Wrap(err, "load cookies failed")
	}

	page := a.browser.MustPage()
	defer page.Close()

	if len(cookieParams) > 0 {
		if err := page.SetCookies(cookieParams); err != nil {
			logrus.Warnf("[%s] Set cookies failed: %v", a.platform, err)
		}
	}

	helper := browser.NewPageHelper(page)

	if err := helper.Navigate(a.publishURL); err != nil {
		return errors.Wrap(err, "navigate to publish page failed")
	}

	time.Sleep(3 * time.Second)

	has, _, _ := page.Has(".avatar-wrapper, .user-info")
	if !has {
		return errors.New("not logged in, please login first")
	}

	if content.Type == publisher.ContentTypeVideo {
		if err := a.uploadVideo(page, content.VideoPath); err != nil {
			return errors.Wrap(err, "upload video failed")
		}
	} else {
		if err := a.uploadImages(page, content.ImagePaths); err != nil {
			return errors.Wrap(err, "upload images failed")
		}
	}

	if err := a.fillContent(page, content); err != nil {
		return errors.Wrap(err, "fill content failed")
	}

	if err := a.submitPublish(page); err != nil {
		return errors.Wrap(err, "publish failed")
	}

	logrus.Infof("[%s] Publish success", a.platform)
	return nil
}

func (a *XiaohongshuAdapter) uploadVideo(page *rod.Page, videoPath string) error {
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("video file not found: %s", videoPath)
	}

	logrus.Infof("[%s] Uploading video: %s", a.platform, videoPath)

	fileInput, err := page.Element("input[type='file'][accept*='video']")
	if err != nil {
		return errors.Wrap(err, "find video upload input failed")
	}

	if err := fileInput.SetFiles([]string{videoPath}); err != nil {
		return errors.Wrap(err, "set video file failed")
	}

	logrus.Infof("[%s] Waiting for video upload...", a.platform)
	time.Sleep(5 * time.Second)

	return nil
}

func (a *XiaohongshuAdapter) uploadImages(page *rod.Page, imagePaths []string) error {
	helper := browser.NewPageHelper(page)

	for i, imgPath := range imagePaths {
		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			return fmt.Errorf("image file not found: %s", imgPath)
		}

		logrus.Infof("[%s] Uploading image %d/%d: %s", a.platform, i+1, len(imagePaths), imgPath)

		fileInput, err := page.Element("input[type='file'][accept*='image']")
		if err != nil {
			return errors.Wrap(err, "find image upload input failed")
		}

		if err := fileInput.SetFiles([]string{imgPath}); err != nil {
			return errors.Wrap(err, "set image file failed")
		}

		helper.RandomDelay(1, 2)
	}

	return nil
}

func (a *XiaohongshuAdapter) fillContent(page *rod.Page, content *publisher.Content) error {
	helper := browser.NewPageHelper(page)

	title := content.Title
	if len(title) > 20 {
		title = title[:20]
	}

	titleInput, err := page.Element("input[placeholder*='title'], input[name*='title']")
	if err == nil {
		if err := titleInput.Input(title); err != nil {
			logrus.Warnf("[%s] Input title failed: %v", a.platform, err)
		}
		helper.RandomDelay(0.5, 1)
	}

	body := content.Body
	if len(body) > 1000 {
		body = body[:1000]
	}

	contentInput, err := page.Element("textarea[placeholder*='content'], textarea[name*='content']")
	if err == nil {
		if err := contentInput.Input(body); err != nil {
			logrus.Warnf("[%s] Input body failed: %v", a.platform, err)
		}
		helper.RandomDelay(0.5, 1)
	}

	for _, tag := range content.Tags {
		tagInput, err := page.Element("input[placeholder*='tag'], input[placeholder*='topic']")
		if err != nil {
			logrus.Warnf("[%s] Find tag input failed: %v", a.platform, err)
			continue
		}

		tagInput.Input("#" + tag)
		time.Sleep(500 * time.Millisecond)
		helper.RandomDelay(0.3, 0.7)
	}

	return nil
}

func (a *XiaohongshuAdapter) submitPublish(page *rod.Page) error {
	helper := browser.NewPageHelper(page)

	publishBtn, err := page.Element("button[type='submit'], button[class*='publish']")
	if err != nil {
		return errors.Wrap(err, "find publish button failed")
	}

	vis, err := publishBtn.Visible()
	if err != nil || !vis {
		return errors.New("publish button not visible")
	}

	helper.RandomDelay(1, 2)

	if err := publishBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return errors.Wrap(err, "click publish button failed")
	}

	logrus.Infof("[%s] Clicked publish button, waiting...", a.platform)
	time.Sleep(5 * time.Second)

	return nil
}

type PublisherFactory struct {
	mu      sync.RWMutex
	creators map[string]func(*publisher.Options) publisher.Publisher
}

func NewPublisherFactory() *PublisherFactory {
	return &PublisherFactory{
		creators: make(map[string]func(*publisher.Options) publisher.Publisher),
	}
}

func (f *PublisherFactory) Register(platform string, creator func(*publisher.Options) publisher.Publisher) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[platform] = creator
}

func (f *PublisherFactory) Create(platform string, opts *publisher.Options) (publisher.Publisher, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	creator, ok := f.creators[platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	return creator(opts), nil
}

func (f *PublisherFactory) Platforms() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	platforms := make([]string, 0, len(f.creators))
	for p := range f.creators {
		platforms = append(platforms, p)
	}
	return platforms
}

func DefaultFactory() *PublisherFactory {
	f := NewPublisherFactory()

	f.Register("douyin", func(opts *publisher.Options) publisher.Publisher {
		return NewDouyinAdapter(opts)
	})

	f.Register("toutiao", func(opts *publisher.Options) publisher.Publisher {
		return NewToutiaoAdapter(opts)
	})

	f.Register("xiaohongshu", func(opts *publisher.Options) publisher.Publisher {
		return NewXiaohongshuAdapter(opts)
	})

	return f
}
